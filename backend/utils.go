package main

import (
	"encoding/json"
	"errors"
	"os"
)

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

type Settings struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	ApiKey            string `json:"apikey"`
	Example           bool   `json:"example"`
	ChallengeSettings `json:"challenge"`
	PathsSettings     `json:"paths"`
}

type ChallengeSettings struct {
	Difficulty []int `json:"difficulty"`
	Latency    []int `json:"latency"`
}

type PathsSettings struct {
	GetChallenge    string `json:"get_challenge"`
	VerifyChallenge string `json:"verify_challenge"`
}

func loadSettings() (Settings, error) {
	settingsFile, err := os.Open("settings.json")
	if err != nil {
		return Settings{}, err
	}
	defer settingsFile.Close()

	var settings Settings
	decoder := json.NewDecoder(settingsFile)
	err = decoder.Decode(&settings)
	if err != nil {
		return Settings{}, err
	}

	if settings.ApiKey == "" {
		return Settings{}, errors.New("not found api key")

	} else if settings.Port == 0 {
		return Settings{}, errors.New("not found port")

	} else if len(settings.Difficulty) != 2 {
		return Settings{}, errors.New("invalid difficulty")

	} else if len(settings.Difficulty) != 2 {
		return Settings{}, errors.New("invalid latency")

	} else if settings.GetChallenge == "" {
		return Settings{}, errors.New("not found get challenge path")

	} else if settings.VerifyChallenge == "" {
		return Settings{}, errors.New("not found verify challenge path")

	}

	return settings, nil
}
