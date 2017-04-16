package main

import (
	"flag"
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
	"github.com/celestialstats/chatlog"
	"github.com/celestialstats/metacache"
)

var quit = make(chan struct{})
var clientId = flag.String("client-id", "", "Discord Client ID")
var clientSecret = flag.String("client-secret", "", "Discord Client Secret")
var botToken = flag.String("bot-token", "", "Discord Bot Token")
var rmqHostname = flag.String("rmq-hostname", "", "RabbitMQ Server Hostname")
var rmqPort = flag.String("rmq-port", "", "RabbitMQ Server Port")
var rmqUsername = flag.String("rmq-username", "", "RabbitMQ Username")
var rmqPassword = flag.String("rmq-password", "", "RabbitMQ Password")
var rmqLogQueue = flag.String("rmq-log-queue", "", "RabbitMQ Log Queue")
var cacheExpiryMin = flag.Int("cache-expiry", -1, "Cache Expiration in Minutes")
var logger *chatlog.ChatLog
var metaChannels *metacache.MetaCache

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
		*rmqHostname = os.Getenv("DISCORD_RABBITMQ_HOSTNAME")
	}
	if *rmqPort == "" {
		*rmqPort = os.Getenv("DISCORD_RABBITMQ_PORT")
	}
	if *rmqUsername == "" {
		*rmqUsername = os.Getenv("DISCORD_RABBITMQ_USERNAME")
	}
	if *rmqPassword == "" {
		*rmqPassword = os.Getenv("DISCORD_RABBITMQ_PASSWORD")
	}
	if *rmqLogQueue == "" {
		*rmqLogQueue = os.Getenv("DISCORD_RABBITMQ_LOG_QUEUE")
	}
	if *cacheExpiryMin == -1 {
		val, err := strconv.Atoi(os.Getenv("CACHE_EXPIRY"))
		if err != nil {
			// Hmm, what to do here?
		}
		*cacheExpiryMin = val
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
	log.Info("\tCache Expiry (Min): ", *cacheExpiryMin)

	metaChannels = metacache.NewMetaCache(*cacheExpiryMin, 1000)

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
	//channelID, _ := strconv.ParseUint(m.ChannelID, 10, 64)
	//authorID, _ := strconv.ParseUint(m.Author.ID, 10, 64)

	// CheckAndUpdate Channels also triggers Guild check.
	metaChannels.CheckAndUpdate(
		m.ChannelID,
		metacache.MetaLookup{
			Parameters: map[string]interface{}{
				"DiscordSession": s,
				"ChannelID":      m.ChannelID,
			},
			Function: getChannelData,
		},
	)

	// Since we may have just ran across this channel for the first time
	// we need to wait for metaChannels to actually cache the data.
	var theMap map[string]string
	var lookupErr error
	for {
		theMap, lookupErr = metaChannels.Retrieve(m.ChannelID)
		if lookupErr == nil {
			break
		}
		log.Debug("No key ", m.ChannelID, " yet... Sleeping 10ms...")
		time.Sleep(10 * time.Millisecond)
	}

	// Print message to stdout.
	log.Debug("[", ts, "] [", m.ChannelID, "] [", m.Author.Username, "] ", m.Content)
	logger.AddEntry(map[string]string{
		"Timestamp": ts,
		"Type":      "MESSAGE",
		"GuildID":   theMap["GuildID"],
		"ChannelID": m.ChannelID,
		"AuthorID":  m.Author.ID,
		"Content":   m.Content,
	})
}

func getChannelData(lookupData map[string]interface{}) map[string]string {
	dg := lookupData["DiscordSession"].(*discordgo.Session)
	chanData, err := dg.Channel(lookupData["ChannelID"].(string))
	if err != nil {
		log.Fatal("Error obtaining guild details:", err)
	}
	log.Debug("\tReturning New Discord Channel Data:")
	log.Debug("\t\tGuildID: ", chanData.GuildID)
	log.Debug("\t\tName: ", chanData.Name)
	return map[string]string{
		"GuildID": chanData.GuildID,
		"Name":    chanData.Name,
	}
}
