package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type DiscordServer struct {
	ServerID      string
	TextChannels  map[string]string
	VoiceChannels map[string]string
}

var (
	Token               string
	BotPrefix           string
	config              *configStruct
	SpotifyClientID     string
	SpotifyClientSecret string
	DiscordServers      map[string]DiscordServer
)

type configStruct struct {
	Token               string                   `json:"Token"`
	BotPrefix           string                   `json:"BotPrefix"`
	SpotifyClientID     string                   `json:"SpotifyClientID"`
	SpotifyClientSecret string                   `json:"SpotifyClientSecret"`
	DiscordServers      map[string]DiscordServer `json:"DiscordServers"`
}

//var dat map[string]interface

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
	DiscordServers = config.DiscordServers

	return nil
}
