package main

import (
	"disco/bot"
	"disco/config"
	"disco/playlist"
	"disco/spotifyClient"
	"fmt"
)

func main() {
	err := config.ReadConfig()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	playlist.ClearDownloads()
	spotifyClient.SetEnvVars()
	spotifyClient.Authenticate()
	bot.Start()

	<-make(chan struct{})
	return

}
