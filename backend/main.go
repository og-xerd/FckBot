package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha3"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	rand2 "math/rand"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/zeebo/blake3"
)

var (
	algorithms = []string{"sha256", "sha3-256", "blake3"}
	secret     []byte
	privateKey [32]byte
	publicKey  [32]byte
	settings   Settings
)

type Challenge struct {
	Type       string `json:"type"`
	Challenge  string `json:"challenge"`
	Difficulty int    `json:"difficulty"`
	Algorithm  string `json:"algorithm"`
	Timestamp  int    `json:"timestamp"`
	Latency    int    `json:"latency"`
	Signature  string `json:"signature"`
}

type Answer struct {
	Challenge
	Answer int `json:"answer"`
}

type AnswerPayload struct {
	Answer string `json:"answer"`
}

func generateChallenge() ([]byte, error) {
	var challenge Challenge

	challenge.Type = "pow"

	challengeDecoded := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, challengeDecoded)
	if err != nil {
		return nil, err
	}
	challenge.Challenge = hex.EncodeToString(challengeDecoded)

	challenge.Difficulty = rand2.Intn(settings.Difficulty[1]-settings.Difficulty[0]+1) + settings.Difficulty[0]

	challenge.Algorithm = algorithms[rand2.Intn(3)]

	challenge.Timestamp = int(time.Now().UnixMilli())

	challenge.Latency = rand2.Intn(settings.Latency[1]-settings.Latency[0]+1) + settings.Latency[0]

	challenge.Signature = hex.EncodeToString(signatureChallenge(challenge))

	challengeJSON, err := json.Marshal(&challenge)
	if err != nil {
		return nil, err
	}

	return challengeJSON, nil
}

func getChallengeHandler(c *fiber.Ctx) error {
	if c.Body() == nil {
		return errors.New("body is empty")
	}

	peerPublicKeyDecode := make([]byte, base64.RawURLEncoding.DecodedLen(len(c.Body())))
	_, err := base64.RawURLEncoding.Decode(peerPublicKeyDecode, c.Body())
	if err != nil {
		return err
	}

	key := getSharedSecret(peerPublicKeyDecode)

	challenge, err := generateChallenge()
	if err != nil {
		return err
	}

	encryptedChallenge, err := encrypt(key, challenge)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"challenge": base64.RawURLEncoding.EncodeToString(encryptedChallenge),
		"publicKey": base64.RawURLEncoding.EncodeToString(publicKey[:]),
	})
}

type Result struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func verifyChallengeHandler(c *fiber.Ctx) error {
	if c.Body() == nil {
		return c.JSON(Result{
			Success: false,
			Error:   "body is empty",
		})
	}

	var answerPayload AnswerPayload
	err := json.Unmarshal(c.Body(), &answerPayload)
	if err != nil {
		return c.JSON(Result{
			Success: false,
			Error:   "invalid payload",
		})
	}

	apikey := c.Get("ApiKey")
	if apikey != settings.ApiKey {
		return c.JSON(Result{
			Success: false,
			Error:   "invalid api key",
		})
	}

	body, err := base64.RawURLEncoding.DecodeString(answerPayload.Answer)
	if err != nil {
		return c.JSON(Result{
			Success: false,
			Error:   err.Error(),
		})
	}

	if len(body) < 45 {
		return c.JSON(Result{
			Success: false,
			Error:   "body is incorrect",
		})
	}

	peerPublicKey := body[0:32]
	ciphertextWithNonce := body[32:]

	key := getSharedSecret(peerPublicKey)

	plaintext, err := decrypt(key, ciphertextWithNonce)
	if err != nil {
		return c.JSON(Result{
			Success: false,
			Error:   err.Error(),
		})
	}

	var answer Answer
	err = json.Unmarshal(plaintext, &answer)
	if err != nil {
		return c.JSON(Result{
			Success: false,
			Error:   err.Error(),
		})
	}

	signature, err := hex.DecodeString(answer.Signature)
	if err != nil {
		return c.JSON(Result{
			Success: false,
			Error:   err.Error(),
		})
	}

	currentSignature := signatureChallenge(answer.Challenge)

	if !hmac.Equal(signature, currentSignature) {
		return c.JSON(Result{
			Success: false,
			Error:   "signature is incorrect",
		})
	}

	currentTimestamp := int(time.Now().UnixMilli())

	if answer.Timestamp+60000 < currentTimestamp {
		return c.JSON(Result{
			Success: false,
			Error:   "challenge expired",
		})
	}

	if answer.Timestamp+answer.Latency > currentTimestamp {
		return c.JSON(Result{
			Success: false,
			Error:   "latency is incorrect",
		})
	}

	if answer.Type == "pow" {
		challenge, err := hex.DecodeString(answer.Challenge.Challenge)
		if err != nil {
			return c.JSON(Result{
				Success: false,
				Error:   err.Error(),
			})
		}

		answerEncode := make([]byte, 4)
		binary.BigEndian.PutUint32(answerEncode, uint32(answer.Answer))
		var result [32]byte

		switch answer.Algorithm {
		case "sha256":
			result = sha256.Sum256(append([]byte(challenge), answerEncode...))

		case "sha3-256":
			result = sha3.Sum256(append([]byte(challenge), answerEncode...))

		case "blake3":
			result = blake3.Sum256(append([]byte(challenge), answerEncode...))

		default:
			return c.JSON(Result{
				Success: false,
				Error:   "algorithm is incorrect",
			})
		}

		if !hashMeetsDifficulty(result[:], answer.Difficulty) {
			return c.JSON(Result{
				Success: false,
				Error:   "invalid challenge",
			})
		}

	} else {
		return c.JSON(Result{
			Success: false,
			Error:   "invalid type challenge",
		})

	}

	return c.JSON(Result{
		Success: true,
	})
}

func main() {
	var err error
	settings, err = loadSettings()
	if err != nil {
		log.Fatal(err)
	}

	publicKey, privateKey, err = generateKeyPair()
	if err != nil {
		log.Fatal(err)
	}

	secret = make([]byte, 32)
	_, err = io.ReadFull(rand.Reader, secret)
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	if settings.Example {
		app.Use(cors.New())
	}

	app.Post(settings.GetChallenge, getChallengeHandler)

	app.Post(settings.VerifyChallenge, verifyChallengeHandler)

	if settings.Example {
		log.Println("Run with example")
		app.Post("/exampleEndpoint", ExampleEndpoint)
	}

	log.Println("Listen", settings.Host+":"+strconv.Itoa(settings.Port))

	err = app.Listen(settings.Host + ":" + strconv.Itoa(settings.Port))
	if err != nil {
		log.Fatal(err)
	}
}
