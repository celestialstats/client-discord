package main

import (
	"log"
	"os"
)

func main() {
	log.Print("Loading Environmental Variables...");
	var clientId = os.Getenv("DISCORD_CLIENTID");
	var clientSecret = os.Getenv("DISCORD_CLIENTSECRET");
	log.Println("\tDISCORD_CLIENTID:", clientId);
	log.Println("\tDISCORD_CLIENTSECRET:", clientSecret);
	log.Fatal("LUL")
}
