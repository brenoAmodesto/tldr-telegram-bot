package main

import (
	"log"
	"time"

	"tldr-telegram-bot/internal/config"
	"tldr-telegram-bot/internal/db"
	"tldr-telegram-bot/internal/telegram"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration validation error: %v", err)
	}

	// Initialize database
	db.InitDB()

	// Load timezone location
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Fatalf("Failed to load timezone location: %v", err)
	}

	// Schedule daily summary at 03:00 BRT
	c := cron.New(cron.WithLocation(loc))
	_, err = c.AddFunc("06 02 * * *", func() {
		log.Println("🕖 Running daily summary at 03:00 BRT")
		telegram.RunDailySummary()
	})
	if err != nil {
		log.Fatalf("Failed to schedule daily summary: %v", err)
	}
	c.Start()

	// Start the Telegram bot
	bot, err := telegram.NewBot()
	if err != nil {
		log.Fatalf("Error initializing Telegram bot: %v", err)
	}

	log.Println("Bot started and listening for messages...")
	bot.Start()
}
