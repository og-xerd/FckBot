package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func ExampleEndpoint(c *fiber.Ctx) error {
	answer := c.Get("x-answer")
	if answer == "" {
		return errors.New("no answer headers")
	}

	var answerPayload = AnswerPayload{
		Answer: answer,
	}

	payload, err := json.Marshal(answerPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://localhost:4000/verifyChallenge", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header = http.Header{
		"ApiKey":       {settings.ApiKey},
		"Content-Type": {"text/plain"},
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(result))

	return nil
}
