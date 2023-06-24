package bot

import (
	"disco/config"
	"disco/dgvoice"
	"disco/playlist"
	"disco/spotifyClient"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify"
)

type BotState int

var playerChan chan string
var stop chan bool

const (
	ListeningForCommand BotState = iota
	ListingOptions
)

type PlayerState int

const (
	Playing PlayerState = iota
	Skipping
	Idle
)

type Bot struct {
	state        BotState
	id           string
	channelID    string
	session      *discordgo.Session
	playList     playlist.PlayList
	currentTrack playlist.Track
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
	bot.playList.Shuffle()
	list()
}

func skip() {
	if playerState == Playing {
		playerChan <- "stop"
		stop <- true
		bot.sendMessage(fmt.Sprintf("> ### Skipped: %s", bot.currentTrack.Title))
		return
	} else {
		bot.sendMessage("> ### Not playing at the moment")
	}
}

// TODO THIS IS UGLY FUCKING FIX IT LATER SOMETIMES
func play() {
	if playerState == Playing {
		bot.sendMessage("> ### I'm already playing you fucking idiot!")
		return
	}
	playerState = Playing
	go bot.playList.Download()
	voiceConnection, err := bot.session.ChannelVoiceJoin("1119642033422868530", "1119642033905205292", false, true)
	if err != nil {
		return
	}

	for i := 0; i < 5; i++ {
		bot.sendMessage(fmt.Sprintf("> ### Beginning playing in %d seconds!", 5-i))
		time.Sleep(1 * time.Second)
	}

	for {
		if len(bot.playList.Queue) == 0 {
			bot.sendMessage("> ### Queue is currently empty, add songs to queue by typing add <songtitle>")
			playerState = Idle
			voiceConnection.Disconnect()
			bot.session.UpdateListeningStatus("")
			return
		}

		track := bot.playList.Dequeue(0)
		bot.session.UpdateListeningStatus(strings.Split(track.Title, " by ")[1])

		bot.currentTrack = track
		path := "./trackAudios/audios/" + track.FilePath + ".mp3"

		bot.sendMessage(fmt.Sprintf("> ### Currently playing: %s", track.Title))

		playerChan = make(chan string)
		stop = make(chan bool)
		dgvoice.PlayAudioFile(voiceConnection, path, stop, playerChan)
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
	song := bot.playList.Dequeue(0)
	bot.sendMessage(fmt.Sprintf("> ### Removed ***%s*** from queue!", song.Title))
}

// TODO THIS IS ABIT DIRTY COMBINE WITH ADD MBY SOMEHOW
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
		song := playlist.Track{Title: fmt.Sprintf("%s by %s", name, artist), FilePath: name, SpotifyURL: url}
		bot.playList.Enqueue(song)
	}
	bot.sendMessage(fmt.Sprintf("> ### Added the playlist %s by user %s", playList.Name, playList.Owner.DisplayName))
}

func add(argv []string) {
	if len(argv) == 0 {
		bot.sendMessage("> ### The add command requeres a search term")
		return
	}

	query := strings.Join(argv, " ")
	track := spotifyClient.Search(query, spotify.SearchTypeTrack)
	url := track.SimpleTrack.ExternalURLs["spotify"]
	name := track.Name
	artist := track.Artists[0].Name
	song := playlist.Track{Title: fmt.Sprintf("%s by %s", name, artist), FilePath: name, SpotifyURL: url}

	bot.playList.Enqueue(song)
	bot.sendMessage(fmt.Sprintf("> ### Added ***%s*** by ***%s*** to the queue", name, artist))
	//bot.state = ListingOptions
}

func handleCommand(session *discordgo.Session, message *discordgo.MessageCreate) {
	bot.session = session
	content := strings.Split(message.Content, " ")
	command, argv := content[0], content[1:]

	switch command {
	case "add":
		add(argv)
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
	case "quit":
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
		state:     ListeningForCommand,
		id:        user.ID,
		channelID: "1119642033905205291",
		session:   goBot,
		playList:  playlist.PlayList{Queue: []playlist.Track{}}}

	fmt.Println("Bot is up and running!")
}
