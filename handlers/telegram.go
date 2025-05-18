// Tasks: Sending and receiving messages, functionality for interacting with the bot.
package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gopkg.in/telebot.v4"
)

type TelegramHandler struct {
	Bot      *telebot.Bot   // Telegram bot instance
	Neural   *NeuralHandler // NeuralHandler instance for working with neural network
	Logger   *log.Logger    // Logger for recording events
	Redis    *redis.Client  // Redis database for storing message intervals
	MsgDelay time.Duration
	ComDelay time.Duration
}

func NewTelegramhandler(bot *telebot.Bot, neural *NeuralHandler, logger *log.Logger, rdb *redis.Client) *TelegramHandler { // Constructor that initializes the Telegram handler
	go func() {
		for {
			if err := neural.cleanUpOldMessages(context.Background(), 24*time.Hour); err != nil {
				logger.Printf("cleanUp error: %v", err)
			}
			time.Sleep(2 * time.Hour)
		}
	}() // Automatic database cleaning when messages are stored for more than X hours specified in the cleanUpMessages function
	return &TelegramHandler{
		Bot:      bot,
		Neural:   neural,
		Logger:   logger,
		Redis:    rdb,
		MsgDelay: 1 * time.Minute,
		ComDelay: 10 * time.Second,
	}
}

func (h *TelegramHandler) RegisterHandlers() { // Registers command and message handlers
	h.Bot.Handle("/start", h.HandleStart)
	h.Bot.Handle("/reset", h.HandleReset)
	h.Bot.Handle("/help", h.HandleHelp)
	h.Bot.Handle("/about", h.HandleAbout)
	h.Bot.Handle("/policy", h.HandlePolicy)
	h.Bot.Handle("/rules", h.HandleRules)

	h.Bot.Handle(telebot.OnText, h.HandleText)
}

func (n *NeuralHandler) cleanUpOldMessages(ctx context.Context, olderThan time.Duration) error {
	interval := fmt.Sprintf("%d hours", int(olderThan.Hours()))
	_, err := n.DB.ExecContext(ctx, `
		DELETE FROM chat_messages WHERE created_at < NOW() - $1::interval
	`, interval) // Clear messages that are stored in the chat messages for more than <olderThan> time
	return err
}

func safeSend(c telebot.Context, text string) error { // safeSend sends a message with panic handling and retries
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic while sending: %v", r)
		}
	}()

	// Try to send 3 times with a delay
	var lastErr error
	for i := range 3 {
		if err := c.Send(text); err == nil {
			return nil
		} else {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}
	return lastErr // If the error still remains, return the last one
}
