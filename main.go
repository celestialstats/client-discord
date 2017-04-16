package main

import (
	"flag"
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
	"github.com/celestialstats/chatlog"
	_ "github.com/davecgh/go-spew/spew"
)

var quit = make(chan struct{})
var clientId = flag.String("client-id", "", "Discord Client ID")
var clientSecret = flag.String("client-secret", "", "Discord Client Secret")
var botToken = flag.String("bot-token", "", "Discord Bot Token")
var rmqHostname = flag.String("rmq-hostname", "", "RabbitMQ Server Hostname")
var rmqPort = flag.String("rmq-port", "", "RabbitMQ Server Port")
var rmqUsername = flag.String("rmq-username", "", "RabbitMQ Username")
var rmqPassword = flag.String("rmq-password", "", "RabbitMQ Password")
var rmqLogQueue = flag.String("rmq-receive-queue", "", "RabbitMQ Receive Queue")
var logger *chatlog.ChatLog

func main() {
	log.SetLevel(log.DebugLevel)
	log.Info("--- Celestial Stats Discord Client ---")
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
	if *rmqHostname == "" {
		*rmqHostname = os.Getenv("LOGREC_RABBITMQ_HOSTNAME")
	}
	if *rmqPort == "" {
		*rmqPort = os.Getenv("LOGREC_RABBITMQ_PORT")
	}
	if *rmqUsername == "" {
		*rmqUsername = os.Getenv("LOGREC_RABBITMQ_USERNAME")
	}
	if *rmqPassword == "" {
		*rmqPassword = os.Getenv("LOGREC_RABBITMQ_PASSWORD")
	}
	if *rmqLogQueue == "" {
		*rmqLogQueue = os.Getenv("LOGREC_RABBITMQ_RECEIVE_QUEUE")
	}
	log.Info("Launch Parameters:")
	log.Info("\tDiscord Client ID: ", *clientId)
	log.Info("\tDiscord Client Secret: ", *clientSecret)
	log.Info("\tDiscord Bot Token: ", *botToken)
	log.Info("\tRabbitMQ Hostname: ", *rmqHostname)
	log.Info("\tRabbitMQ Port: ", *rmqPort)
	log.Info("\tRabbitMQ Username: ", *rmqUsername)
	log.Info("\tRabbitMQ Password: ", *rmqPassword)
	log.Info("\tRabbitMQ Log Queue: ", *rmqLogQueue)

	logger = chatlog.NewChatLog(
		*rmqHostname,
		*rmqPort,
		*rmqUsername,
		*rmqPassword,
		*rmqLogQueue,
		"DISCORD",
		1000,
	)

	go StartClient()

	log.Info("Bot is now running.  Press CTRL-C to exit.")

	<-quit

	log.Info("Exiting...")
	return
}

func StartClient() {
	dg, err := discordgo.New("Bot " + *botToken)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
		return
	}

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
	curTime, _ := m.Timestamp.Parse()
	ts := strconv.FormatInt(curTime.UnixNano()/int64(time.Millisecond), 36)
	// Print message to stdout.
	log.Debug("[", ts, "] [", m.ChannelID, "] [", m.Author.Username, "] ", m.Content)
	logger.AddEntry(map[string]string{
		"Timestamp": ts,
		"Type":      "MESSAGE",
		"ChannelID": m.ChannelID,
		"AuthorID":  m.Author.ID,
		"Content":   m.Content,
	})
}
