package main

import (
	"auth-api/config"
	"auth-api/utils"
	"log"
	"os"
)

func main() {
	log.Println("Starting AUTH API")
	configPath := utils.GetConfigPath(os.Getenv("config"))
	cfg, err := config.GetConfig(configPath)
	if err != nil {
		log.Fatalf("Loading config: %v", err)
	}
}
