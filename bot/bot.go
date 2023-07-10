package bot

import (
	"disco/config"
	"disco/dgvoice"
	"disco/playlist"
	"disco/spotifyClient"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

/*
TODOS:
1. Downloads should start in a separate thread immediately on bot startup, adding tracks should get downloaded as the playlist gets updated.
2. Add command to add albums.
3. Make shure ffmpg get's killed properly on skip, also rewrite skip, pause, resume.
4. Add command to move tracks in playlist
5. Add commands to show history of played tracks
6. Rewrite spotify-dl in pure go
*/

type BotState int

// TODO WHY WOULD WE EVER DO SHIT LIKE THIS THIS SUCKS REWRITE LATER
var skipChan chan string
var pauseChan chan string
var ResumeChan chan string
var stop chan bool

const (
	ListeningForCommand BotState = iota
	ListingOptions
)

type PlayerState int

const (
	Playing PlayerState = iota
	Paused
	Idle
)

type Bot struct {
	state          BotState
	id             string
	channelID      string
	session        *discordgo.Session
	playList       playlist.PlayList
	currentTrack   *playlist.Track
	voiceChannelID string
	ServerID       string
	DiscordServers map[string]config.DiscordServer
}

var playerState PlayerState

var bot Bot

func (b *Bot) sendMessage(message string) {
	bot.session.ChannelMessageSend(bot.channelID, message)
}

func help() {
	helpMessage := `> # RIPBOZO 

	`
	bot.sendMessage(helpMessage)
}

func shuffle() {
	bot.sendMessage("> ## Shuffling playlist")
	bot.playList.Shuffle()
}

func skip() {
	if playerState == Playing {
		skipChan <- "stop"
		stop <- false
		bot.sendMessage(fmt.Sprintf("> ### Skipped: %s", bot.currentTrack.Title))
		return
	} else {
		bot.sendMessage("> ### Not playing at the moment")
	}
}

func quit() {
	if playerState == Playing {
		skipChan <- "stop"
		stop <- false
		playerState = Idle
		bot.session.UpdateGameStatus(1, "")
		return
	} else {
		bot.sendMessage("> ### Not playing at the moment")
	}
}

func pause() {
	if playerState == Playing {
		bot.sendMessage(fmt.Sprintf("> ### Paused: %s", bot.currentTrack.Title))
		playerState = Paused
		pauseChan <- "pause"
	} else {
		bot.sendMessage("> ### Not currently playing")
	}

}

// TODO THIS IS SO FUCKING UGLY FUCKING FIX IT LATER SOMETIMES
func play() {
	if playerState == Paused {
		bot.sendMessage(fmt.Sprintf("> ### Resumed playing: %s", bot.currentTrack.Title))
		ResumeChan <- "play"
		playerState = Playing
		return
	}

	if playerState == Playing {
		bot.sendMessage("> ### I'm already playing you fucking idiot!")
		return
	}
	playerState = Playing

	voiceConnection, err := bot.session.ChannelVoiceJoin(bot.ServerID, bot.voiceChannelID, false, true)
	if err != nil {
		fmt.Printf("something here")
		return
	}

	for {
		if len(bot.playList.Queue) == 0 {
			bot.sendMessage("> ### Queue is currently empty, add songs to queue by typing add <songtitle>")
			playerState = Idle
			voiceConnection.Disconnect()
			bot.session.UpdateListeningStatus("")
			return
		}
		if playerState == Idle {
			bot.sendMessage("> ### Leaving voice")
			voiceConnection.Disconnect()
			return
		}

		track := bot.playList.Dequeue(0)
		if !track.IsDownloaded {

			bot.sendMessage("> ### Loading")
			i := 1
			for !track.IsDownloaded {
				time.Sleep(1000 * time.Millisecond)
				bot.sendMessage("> ### " + strings.Repeat(".", i*10))
				i++
			}
			//bot.sendMessage("> ### Ready to play now :)")
			bot.sendMessage(fmt.Sprintf("> ### Currently playing: %s :musical_note:", track.Title))
		}
		bot.session.UpdateListeningStatus(strings.Split(track.Title, " by ")[1])

		bot.currentTrack = track
		path := "./trackAudios/audios/" + track.FilePath + ".mp3"

		skipChan = make(chan string)
		pauseChan = make(chan string)
		ResumeChan = make(chan string)
		stop = make(chan bool)
		dgvoice.PlayAudioFile(voiceConnection, path, stop, skipChan, pauseChan, ResumeChan)

		//<-stop
		playlist.RemoveFile(path, "./trackAudios/"+track.FilePath)
	}
}

func list() {
	if len(bot.playList.Queue) == 0 {
		bot.sendMessage("> ### No songs currently in queue!")
		return
	}

	var stringBuilder strings.Builder
	stringBuilder.WriteString("### Current tracks in queue (currently shows max 40): \n")

	for i, track := range bot.playList.Queue {
		if i >= 40 {
			break
		}
		stringBuilder.WriteString(fmt.Sprintf("%d. %s \n", i+1, track.Title))
	}

	bot.sendMessage(stringBuilder.String())
}

func remove(argv []string) {
	if len(bot.playList.Queue) == 0 {
		bot.sendMessage("> ### No songs currently in queue!")
		return
	}
	i, err := strconv.Atoi(argv[0])
	if err != nil {
		return
	}
	song := bot.playList.Dequeue(uint(i - 1))
	bot.sendMessage(fmt.Sprintf("> ### Removed ***%s*** from queue!", song.Title))
}

func addPlaylist(argv []string) {
	if len(argv) == 0 {
		bot.sendMessage("> ### The add playlist command requeres a id")
		return
	}
	playList := spotifyClient.SearchPlaylist(strings.Split(argv[0], "playlist/")[1])

	for _, track := range playList.Tracks.Tracks {
		fullTrack := track.Track
		url := fullTrack.SimpleTrack.ExternalURLs["spotify"]
		name := fullTrack.Name
		artist := fullTrack.Artists[0].Name

		song := playlist.Track{
			Title:      fmt.Sprintf("%s by %s", name, artist),
			FilePath:   name,
			SpotifyURL: url}

		bot.playList.Enqueue(&song)
	}
	bot.sendMessage(fmt.Sprintf("> ### Added the playlist %s by user %s", playList.Name, playList.Owner.DisplayName))
}

func addAlbum(argv []string) {
	if len(argv) == 0 {
		bot.sendMessage("> ### The add playlist command requeres a id")
		return
	}

	query := strings.Join(argv, " ")
	album := spotifyClient.SearchAlbum(query)

	albumName := album.Name
	albumArtist := album.Artists[0].Name

	for _, track := range album.Tracks.Tracks {
		url := track.ExternalURLs["spotify"]
		name := track.Name

		song := playlist.Track{
			Title:      fmt.Sprintf("%s by %s", name, albumArtist),
			FilePath:   name,
			SpotifyURL: url}

		bot.playList.Enqueue(&song)
	}
	bot.sendMessage(fmt.Sprintf("> ### Added the album %s by %s", albumName, albumArtist))
}

func addTrack(argv []string) {
	if len(argv) == 0 {
		bot.sendMessage("> ### The add command requeres a search term")
		return
	}

	query := strings.Join(argv, " ")
	track := spotifyClient.SearchTrack(query)

	url := track.SimpleTrack.ExternalURLs["spotify"]
	name := track.Name
	artist := track.Artists[0].Name

	song := playlist.Track{
		Title:      fmt.Sprintf("%s by %s", name, artist),
		FilePath:   name,
		SpotifyURL: url}

	bot.playList.Enqueue(&song)
	bot.sendMessage(fmt.Sprintf("> ### Added ***%s*** by ***%s*** to the queue", name, artist))
	//bot.state = ListingOptions
}

func handleCommand(session *discordgo.Session, message *discordgo.MessageCreate) {
	bot.session = session
	content := strings.Split(message.Content, " ")
	command, argv := content[0], content[1:]

	switch command {
	case "add":
		addTrack(argv)
	case "addalbum":
		addAlbum(argv)
	case "addplaylist":
		addPlaylist(argv)
	case "remove":
		remove(argv)
	case "list":
		list()
	case "play":
		play()
	case "skip":
		skip()
	case "shuffle":
		shuffle()
	case "pause":
		pause()
	case "quit":
		quit()
	case "currentsong":
		if playerState == Playing {
			bot.sendMessage(fmt.Sprintf("> ### Currently playing: %s", bot.currentTrack.Title))
		} else {
			bot.sendMessage("> ### Not playing currently")
		}
	case "help":
		help()
	case "\"help\"":
		bot.sendMessage("> # You are very stupid")
	default:
		bot.sendMessage(fmt.Sprintf("> ### ***\"%s\"*** is not a recognized command, type ***\"help\"*** to get the full list of comamnds", command))
	}
}

func messageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == bot.id {
		return
	}

	if message.ChannelID == bot.channelID {
		switch bot.state {
		case ListeningForCommand:
			handleCommand(session, message)
		case ListingOptions:
		}
	}
}

func Start() {
	goBot, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	playerState = Idle

	user, err := goBot.User("@me")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	goBot.AddHandler(messageHandler)

	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	bot = Bot{
		state:          ListeningForCommand,
		id:             user.ID,
		channelID:      "1119642033905205291",
		voiceChannelID: "1119642033905205292",
		ServerID:       "1119642033422868530",
		//channelID: "1122024458824192080",
		//voiceChannelID: "816247222261121045",
		//ServerID: "750434166310568058",
		session:  goBot,
		playList: playlist.PlayList{Queue: []*playlist.Track{}}}
	bot.playList.StartDownloadThread()

	fmt.Println("Bot is up and running!")
	bot.session.UpdateGameStatus(1, "")
}
