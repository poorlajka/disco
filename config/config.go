package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var (
	Token               string
	BotPrefix           string
	config              *configStruct
	SpotifyClientID     string
	SpotifyClientSecret string
)

type configStruct struct {
	Token               string `json:"Token"`
	BotPrefix           string `json:"BotPrefix"`
	SpotifyClientID     string `json:"SpotifyClientID"`
	SpotifyClientSecret string `json:"SpotifyClientSecret"`
}

func ReadConfig() error {
	fmt.Println("Reading config file...")

	file, err := ioutil.ReadFile("./config.json")

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	fmt.Println(string(file))

	err = json.Unmarshal(file, &config)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	Token = config.Token
	BotPrefix = config.BotPrefix
	SpotifyClientID = config.SpotifyClientID
	SpotifyClientSecret = config.SpotifyClientSecret

	return nil
}
