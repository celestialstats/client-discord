package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/celestialstats/chatlog"
	"github.com/celestialstats/metacache"
	"github.com/davecgh/go-spew/spew"
	"log"
	"os"
	"strconv"
	"time"
)

var quit = make(chan struct{})
var clientId = flag.String("client-id", "", "Discord Client ID")
var clientSecret = flag.String("client-secret", "", "Discord Client Secret")
var botToken = flag.String("bot-token", "", "Discord Bot Token")
var logDir = flag.String("log-dir", "", "Chat Log Directory")
var logger *chatlog.ChatLog
var metaUsers *metacache.MetaCache
var metaGuilds *metacache.MetaCache
var metaChannels *metacache.MetaCache

func main() {
	log.Println("--- Celestial Stats Discord Client ---")
	flag.Parse()
	if *clientId == "" {
		*clientId = os.Getenv("DISCORD_CLIENTID")
	}
	if *clientSecret == "" {
		*clientSecret = os.Getenv("DISCORD_CLIENTSECRET")
	}
	if *botToken == "" {
		*botToken = os.Getenv("DISCORD_BOTTOKEN")
	}
	if *logDir == "" {
		*logDir = os.Getenv("LOGDIR")
	}
	log.Println("Launch Parameters:")
	log.Println("\tDISCORD_CLIENTID:", *clientId)
	log.Println("\tDISCORD_CLIENTSECRET:", *clientSecret)
	log.Println("\tDISCORD_BOTTOKEN:", *botToken)
	log.Println("\tLOGDIR:", *logDir)
	logger = chatlog.NewChatLog(*logDir, "DISCORD", 1000)
	metaUsers = metacache.NewMetaCache(1, 100)
	metaGuilds = metacache.NewMetaCache(1, 100)
	metaChannels = metacache.NewMetaCache(1, 100)

	go StartClient()

	log.Println("Bot is now running.  Press CTRL-C to exit.")

	<-quit

	log.Println("Exiting...")
	return
}

func StartClient() {
	dg, err := discordgo.New("Bot " + *botToken)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
		return
	}

	u, err := dg.User("@me")
	spew.Dump(u)
	if err != nil {
		log.Fatal("Error obtaining account details:", err)
	}

	ug, err := dg.UserGuilds()
	spew.Dump(ug)
	if err != nil {
		log.Fatal("Error obtaining guild details:", err)
	}

	/*
		uc, err := dg.Channel("298236553350873090")
		spew.Dump(uc)
		if err != nil {
			log.Fatal("Error obtaining guild details:", err)
		}
	*/

	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening connection:", err)
		return
	}
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Print message to stdout.
	spew.Dump(m)
	metaChannels.CheckAndUpdate(m.ChannelID, metacache.MetaLookup{Parameters: map[string]string{}, Function: getChannelData})
	metaChannels.PrintData()
	fmt.Printf("%20s %20s %20s > %s\n", m.ChannelID, time.Now().Format(time.Stamp), m.Author.Username, m.Content)
	curTime, _ := m.Timestamp.Parse()
	logger.AddEntry(map[string]string{
		"Timestamp":      strconv.FormatInt(curTime.UnixNano()/int64(time.Millisecond), 36),
		"Server":         m.ChannelID,
		"AuthorID":       m.Author.ID,
		"AuthorUsername": m.Author.Username,
		"Content":        m.Content,
	})
}

func getChannelData(lookupData map[string]string) map[string]string {
	return map[string]string{
		"Attr One":   "Value One",
		"Attr Two":   "Value Two",
		"Attr Three": "Value Three",
		"Attr Four":  "Value Four",
	}
}
