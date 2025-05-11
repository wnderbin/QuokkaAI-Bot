package main

import (
	"database/sql"
	"log"
	"quokka-ai-bot/config"
	"quokka-ai-bot/handlers"
	"quokka-ai-bot/migrator"
	"quokka-ai-bot/utils"
	"time"

	"github.com/redis/go-redis/v9"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/telebot.v4"
)

func main() {
	cfg := config.Load() // loading the configuration
	logger := utils.NewLogger(cfg.Debug)
	logger.SetOutput(&lumberjack.Logger{
		Filename:   "quokkabot.log", // Log file
		MaxSize:    100,             // MB
		MaxBackups: 10,              // Maximum files for storage
		MaxAge:     10,              // Maximum storage time
		Compress:   true,            // Compression of old logs
	}) // logger initialization
	db, err := sql.Open("postgres", "postgres://wnd:123@localhost:5432/wnd?sslmode=disable") // database connection
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	migrator.ApplyMigrations(db) // apply migrations to the database
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	neuralHandler := handlers.NewNeuralHandler(cfg.DeepSeekToken, db) // install neural network handler
	botSettings := telebot.Settings{                                  // telebot settings
		Token: cfg.TgToken,
		Poller: &telebot.LongPoller{
			Timeout: 10 * time.Second,
		},
	}
	bot, err := telebot.NewBot(botSettings) // telebot init
	if err != nil {
		logger.Fatalf("Failed to create bot: %v", err)
	}
	tgHandler := handlers.NewTelegramhandler(bot, neuralHandler, logger, redisClient)
	tgHandler.RegisterHandlers()

	logger.Println("Starting bot...")
	bot.Start()
}
