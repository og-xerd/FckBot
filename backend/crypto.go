package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

func generateKeyPair() ([32]byte, [32]byte, error) {
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return [32]byte{}, [32]byte{}, err
	}

	return *publicKey, *privateKey, nil
}

func getSharedSecret(peerPublicKey []byte) []byte {
	var publicKey [32]byte
	copy(publicKey[:], peerPublicKey)

	var sharedSecret [32]byte
	box.Precompute(&sharedSecret, &publicKey, &privateKey)

	return sharedSecret[:]
}

func signatureChallenge(challenge Challenge) []byte {
	hmac := hmac.New(sha256.New, secret)

	hmac.Write([]byte(challenge.Type))
	hmac.Write([]byte(challenge.Challenge))
	hmac.Write([]byte{byte(challenge.Difficulty)})
	hmac.Write([]byte(challenge.Algorithm))

	timestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(timestamp, uint64(challenge.Timestamp))
	hmac.Write(timestamp)

	latency := make([]byte, 2)
	binary.BigEndian.PutUint16(latency, uint16(challenge.Latency))
	hmac.Write(latency)

	signature := hmac.Sum(nil)

	return signature
}

func encrypt(key, plaintext []byte) (ciphertextWithNonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aes, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aes.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aes.Seal(nil, nonce, plaintext, nil)

	ciphertextWithNonce = append(nonce, ciphertext...)
	return
}

func decrypt(key, ciphertextWithNonce []byte) (plaintext []byte, err error) {
	if len(ciphertextWithNonce) < 12 {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertextWithNonce[:12]
	ciphertext := ciphertextWithNonce[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err = gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return
}

func countLeadingZeros(hashBuffer []byte) int {
	zeros := 0

	for _, b := range hashBuffer {
		for i := 7; i >= 0; i-- {
			if (b>>i)&1 == 1 {
				return zeros
			}
			zeros++
		}
	}

	return zeros
}

func hashMeetsDifficulty(hashBuffer []byte, difficulty int) bool {
	return countLeadingZeros(hashBuffer) >= difficulty
}
