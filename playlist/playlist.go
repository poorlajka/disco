package playlist

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
)

type Track struct {
	Title        string
	FilePath     string
	SpotifyURL   string
	IsDownloaded bool
}

type PlayList struct {
	Queue   []*Track
	Playing bool
}

func (pl *PlayList) Enqueue(song *Track) {
	pl.Queue = append(pl.Queue, song)
}

func (pl *PlayList) Dequeue(i uint) *Track {
	track := pl.Queue[i]
	if len(pl.Queue) == 1 {
		pl.Queue = []*Track{}
		return track
	}
	ret := make([]*Track, 0)
	ret = append(ret, pl.Queue[:i]...)
	new := append(ret, pl.Queue[i+1:]...)
	pl.Queue = new
	return track
}

func (pl *PlayList) Shuffle() {
	//Fisher--Yates
	var randArr []int
	for i := 0; i < len(pl.Queue); i++ {
		randArr = append(randArr, rand.Intn(len(pl.Queue)))
	}

	for i := 0; i < len(pl.Queue); i++ {
		//Swap
		temp := pl.Queue[randArr[i]]
		pl.Queue[randArr[i]] = pl.Queue[i]
		pl.Queue[i] = temp
	}
}

func SpotifyDownload(url string, fileName string) {
	fmt.Println(url)
	cmd := exec.Command("python", "./spotify-dl-master/spotify_dl/spotify_dl.py", "--url", url, "--output", "./trackAudios", "--name", fileName)
	err := cmd.Run()

	if err != nil {
		fmt.Printf("Error downloading %v", err)
	}
}

func ClearDownloads() {
	err := os.RemoveAll("./trackAudios")
	if err != nil {
		fmt.Printf("%v", err)

		//log.Fatalf("%v", err)
	}

	err = os.Mkdir("./trackAudios", os.ModePerm)
	if err != nil {
		fmt.Printf("%v", err)
		//log.Fatalf("%v", err)
	}
}

func RemoveFile(filePath string, name string) {
	err := os.Remove(filePath)
	if err != nil {
		fmt.Printf("%v", err)
	}
}

func (pl *PlayList) Download() {
	for _, track := range pl.Queue {
		if !track.IsDownloaded {
			os.Remove("./trackAudios/download_list.log")
			SpotifyDownload(track.SpotifyURL, "./audios/"+track.FilePath)
			track.IsDownloaded = true
			os.Remove("./trackAudios/" + track.FilePath)
			os.Remove("./trackAudios/download_list.log")
		}
	}
}
