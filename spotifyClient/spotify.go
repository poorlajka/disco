package spotifyClient

import (
	"context"
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
	err := os.Setenv("SPOTIFY_ID", "b69148d285ce4b21a21be46f55b43ef8")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("SPOTIFY_SECRET", "4a3531a7337d4b50b4ddd466ecb7f27a")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("SPOTIPY_CLIENT_ID", "b69148d285ce4b21a21be46f55b43ef8")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("SPOTIPY_CLIENT_SECRET", "4a3531a7337d4b50b4ddd466ecb7f27a")
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

func Search(query string, searchType spotify.SearchType) spotify.FullTrack {
	result, err := spotifyClient.client.Search(spotifyClient.ctx, query, searchType)
	if err != nil {
		fmt.Printf("couldn't search: %v", err)
	}

	//tracks := result.Tracks.Tracks

	/*
		for i := 0; i < 10; i++ {
			fmt.Printf("%d %s by ***%s*** \n", i+1, tracks[i].SimpleTrack.Name, tracks[i].Artists[0].Name)
		}
	*/

	return result.Tracks.Tracks[0]

	//.SimpleTrack.ExternalURLs["spotify"]

}

func SearchPlaylist(id string) spotify.FullPlaylist {
	playList, err := spotifyClient.client.GetPlaylist(spotifyClient.ctx, spotify.ID(id))
	if err != nil {
		fmt.Printf("couldn't search: %v", err)
	}
	return *playList
}
