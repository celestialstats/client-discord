package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
	"github.com/celestialstats/chatlog"
	"github.com/celestialstats/metacache"
	"github.com/davecgh/go-spew/spew"
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
var metaGuilds *metacache.MetaCache
var metaChannels *metacache.MetaCache
var metaUsers *metacache.MetaCache

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
	if *logDir == "" {
		*logDir = os.Getenv("LOGDIR")
	}
	log.Info("Launch Parameters:")
	log.Info("\tDISCORD_CLIENTID:", *clientId)
	log.Info("\tDISCORD_CLIENTSECRET:", *clientSecret)
	log.Info("\tDISCORD_BOTTOKEN:", *botToken)
	log.Info("\tLOGDIR:", *logDir)
	logger = chatlog.NewChatLog(*logDir, "DISCORD", 1000)
	metaGuilds = metacache.NewMetaCache(1, 100)
	metaChannels = metacache.NewMetaCache(1, 100)
	metaUsers = metacache.NewMetaCache(1, 100)

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
	// Print message to stdout.
	// CheckAndUpdate Channels also triggers Guild check.
	metaChannels.CheckAndUpdate(
		m.ChannelID,
		metacache.MetaLookup{
			Parameters: map[string]interface{}{
				"DiscordSession": s,
				"ChannelID":      m.ChannelID,
				"GuildCache":     metaGuilds,
			},
			Function: getChannelData,
		},
	)
	metaUsers.CheckAndUpdate(
		m.Author.ID,
		metacache.MetaLookup{
			Parameters: map[string]interface{}{
				"Author": m.Author,
			},
			Function: getUserData,
		},
	)
	metaGuilds.PrintData()
	metaChannels.PrintData()
	metaUsers.PrintData()
	fmt.Printf(
		"%20s %20s %20s > %s\n",
		m.ChannelID,
		time.Now().Format(time.Stamp),
		m.Author.Username,
		m.Content,
	)
	curTime, _ := m.Timestamp.Parse()
	logger.AddEntry(map[string]string{
		"Timestamp":      strconv.FormatInt(curTime.UnixNano()/int64(time.Millisecond), 36),
		"Server":         m.ChannelID,
		"AuthorID":       m.Author.ID,
		"AuthorUsername": m.Author.Username,
		"Content":        m.Content,
	})
}

func getChannelData(lookupData map[string]interface{}) map[string]string {
	dg := lookupData["DiscordSession"].(*discordgo.Session)
	chanData, err := dg.Channel(lookupData["ChannelID"].(string))
	if err != nil {
		log.Fatal("Error obtaining guild details:", err)
	}
	lookupData["GuildCache"].(*metacache.MetaCache).CheckAndUpdate(
		chanData.GuildID,
		metacache.MetaLookup{
			Parameters: map[string]interface{}{
				"DiscordSession": dg,
				"GuildID":        chanData.GuildID,
			},
			Function: getGuildData,
		},
	)
	log.Debug("\t\t\tReturning New Discord Channel Data:")
	log.Debug("\t\t\t\tGuildID: ", chanData.GuildID)
	log.Debug("\t\t\t\tName: ", chanData.Name)
	return map[string]string{
		"GuildID": chanData.GuildID,
		"Name":    chanData.Name,
	}
}

func getGuildData(lookupData map[string]interface{}) map[string]string {
	dg := lookupData["DiscordSession"].(*discordgo.Session)
	guildData, err := dg.Guild(lookupData["GuildID"].(string))
	if err != nil {
		log.Fatal("Error obtaining guild details:", err)
	}
	log.Debug("\t\t\tReturning New Discord Guild Data:")
	log.Debug("\t\t\t\tName: ", guildData.Name)
	log.Debug("\t\t\t\tRegion: ", guildData.Region)
	log.Debug("\t\t\t\tOwnerID: ", guildData.OwnerID)
	log.Debug("\t\t\t\tMemberCount: ", strconv.Itoa(guildData.MemberCount))
	return map[string]string{
		"Name":        guildData.Name,
		"Region":      guildData.Region,
		"OwnerID":     guildData.OwnerID,
		"MemberCount": strconv.Itoa(guildData.MemberCount),
	}
}

func getUserData(lookupData map[string]interface{}) map[string]string {
	userData := lookupData["Author"].(*discordgo.User)
	spew.Dump()
	return map[string]string{
		"Username":      userData.Username,
		"Discriminator": userData.Discriminator,
		"IsBot":         strconv.FormatBool(userData.Bot),
	}
}
