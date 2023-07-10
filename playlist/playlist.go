package playlist

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"sync"
)

type PlayList struct {
	Queue     []*Track
	Playing   bool
	IsUpdated bool
	Mutex     sync.Mutex
}

type Track struct {
	Title        string
	FilePath     string
	SpotifyURL   string
	IsDownloaded bool
}

func (pl *PlayList) Enqueue(song *Track) {
	pl.IsUpdated = false
	/*
		pl.Mutex.Lock()
		defer pl.Mutex.Unlock()
	*/

	pl.Queue = append(pl.Queue, song)
}

func (pl *PlayList) Dequeue(i uint) *Track {
	pl.IsUpdated = false
	//pl.Mutex.Lock()
	//defer pl.Mutex.Unlock()

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
	pl.IsUpdated = false
	pl.Mutex.Lock()
	defer pl.Mutex.Unlock()

	//Fisher--Yates algorithm
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

func (pl *PlayList) swap(index1 int, index2 int) {
	pl.IsUpdated = false
	pl.Mutex.Lock()
	defer pl.Mutex.Unlock()

	pl.Queue[index1] = pl.Queue[index2]
}

func SpotifyDownload(url string, fileName string) {
	fmt.Println(fileName)
	cmd := exec.Command(
		"python",
		"./spotify-dl-master/spotify_dl/spotify_dl.py",
		"--url",
		url,
		"--output",
		"./trackAudios",
		"--name",
		fileName)

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

func (playlist *PlayList) StartDownloadThread() {
	go playlist.Download()
}

func (playlist *PlayList) Download() {
	for {
		playlist.DownloadTracks()
	}
}

func (playlist *PlayList) DownloadTracks() {
	playlist.Mutex.Lock()
	defer playlist.Mutex.Unlock()

	playlist.IsUpdated = true
	for _, track := range playlist.Queue {
		if !playlist.IsUpdated {
			return
		}
		track.Download()
	}
}

func (track *Track) Download() {
	if !track.IsDownloaded {
		os.Remove("./trackAudios/download_list.log")
		SpotifyDownload(track.SpotifyURL, "./audios/"+track.FilePath)
		track.IsDownloaded = true
		os.Remove("./trackAudios/" + track.FilePath)
		os.Remove("./trackAudios/download_list.log")
	}
}
