package spotifyClient

import (
	"context"
	"disco/config"
	"fmt"
	"os"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

type SpotifyClient struct {
	ctx    context.Context
	client *spotify.Client
}

var spotifyClient SpotifyClient

func SetEnvVars() {
	err := os.Setenv("SPOTIFY_ID", config.SpotifyClientID)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("SPOTIFY_SECRET", config.SpotifyClientSecret)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("SPOTIPY_CLIENT_ID", config.SpotifyClientID)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("SPOTIPY_CLIENT_SECRET", config.SpotifyClientSecret)
	if err != nil {
		fmt.Println(err)
	}
}

func Authenticate() {
	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		fmt.Printf("couldn't get token: %v", err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)

	spotifyClient.client, spotifyClient.ctx = client, ctx
}

func SearchTrack(query string) spotify.FullTrack {
	result, err := spotifyClient.client.Search(spotifyClient.ctx, query, spotify.SearchTypeTrack)
	if err != nil {
		fmt.Printf("couldn't search: %v", err)
	}
	return result.Tracks.Tracks[0]
}

func SearchAlbum(query string) spotify.FullAlbum {
	result, err := spotifyClient.client.Search(spotifyClient.ctx, query, spotify.SearchTypeAlbum)
	if err != nil {
		fmt.Printf("couldn't search: %v", err)
	}

	albumID := result.Albums.Albums[0].ID
	album, err := spotifyClient.client.GetAlbum(spotifyClient.ctx, albumID)
	if err != nil {
		fmt.Printf("couldn't search: %v", err)
	}
	return *album
}

func SearchPlaylist(id string) spotify.FullPlaylist {
	playList, err := spotifyClient.client.GetPlaylist(spotifyClient.ctx, spotify.ID(id))
	if err != nil {
		fmt.Printf("couldn't search: %v", err)
	}
	return *playList
}
