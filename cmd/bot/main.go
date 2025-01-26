package main

import (
	"log"
	"whisp/internal/bot"
	"whisp/internal/config"
	"whisp/internal/storage"
)

func main(){
	//Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	//Initialize SQLite database
	db, err := storage.NewSQLiteDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()


	//Start the bot
	repo:= storage.NewRepository(db)
	bot.Start(cfg, repo)
}