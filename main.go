package main

import (
	"log"
	"time"
	"os"
	"fmt"
	"flag"
	"github.com/davecgh/go-spew/spew"
	"github.com/bwmarrin/discordgo"
	"github.com/celestialstats/chatlog"
)

var clientId = flag.String("client-id", "", "Discord Client ID")
var clientSecret = flag.String("client-secret", "", "Discord Client Secret")
var botToken = flag.String("bot-token", "", "Discord Bot Token")
var logDir = flag.String("log-dir", "", "Chat Log Directory")

func main() {
	log.Println("--- Celestial Stats Discord Client ---")
	flag.Parse()
	if *clientId == "" { *clientId = os.Getenv("DISCORD_CLIENTID") }
	if *clientSecret == "" { *clientSecret = os.Getenv("DISCORD_CLIENTSECRET") }
	if *botToken == "" { *botToken = os.Getenv("DISCORD_BOTTOKEN") }
	if *logDir == "" { *logDir = os.Getenv("LOGDIR") }
	log.Println("Launch Parameters:")
	log.Println("\tDISCORD_CLIENTID:", *clientId)
	log.Println("\tDISCORD_CLIENTSECRET:", *clientSecret)
	log.Println("\tDISCORD_BOTTOKEN:", *botToken)
	log.Println("\tLOGDIR:", *logDir)
	testLog := chatlog.NewChatLog(*logDir, "DISCORD", "test")
	log.Println(testLog.ComputeFilename())
	testLog.OpenLog()
	return
	dg, err := discordgo.New("Bot " + *botToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	// Get the account information.
	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}
	spew.Dump(u)
	// Get the account information.
	uc, err := dg.UserGuilds()
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}
	spew.Dump(uc)
	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return

	log.Fatal("LUL")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Print message to stdout.
	fmt.Printf("%20s %20s %20s > %s\n", m.ChannelID, time.Now().Format(time.Stamp), m.Author.Username, m.Content)
}
