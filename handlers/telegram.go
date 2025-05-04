package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
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
}

func NewTelegramhandler(bot *telebot.Bot, neural *NeuralHandler, logger *log.Logger, rdb *redis.Client) *TelegramHandler { // Constructor that initializes the Telegram handler
	go func() {
		for {
			if err := neural.cleanUpOldMessages(context.Background(), 5*time.Minute); err != nil {
				logger.Printf("cleanUp error: %v", err)
			}
			time.Sleep(1 * time.Minute)
		}
	}() // Automatic database cleaning when messages are stored for more than X hours specified in the cleanUpMessages function
	return &TelegramHandler{
		Bot:      bot,
		Neural:   neural,
		Logger:   logger,
		Redis:    rdb,
		MsgDelay: 1 * time.Minute,
	}
}

func (h *TelegramHandler) RegisterHandlers() { // Registers command and message handlers
	h.Bot.Handle("/start", h.HandleStart)
	h.Bot.Handle("/reset", h.HandleReset)
	h.Bot.Handle("/help", h.HandleHelp)
	h.Bot.Handle("/about", h.HandleAbout)

	h.Bot.Handle(telebot.OnText, h.HandleText)
}

func (h *TelegramHandler) HandleStart(c telebot.Context) error { // Start bot
	user := c.Sender()
	h.Logger.Printf("New user: %d %s", user.ID, user.Username)
	return c.Send(`
		<b>Приветствую!</b> Я бот с интеграцией DeepSeek AI (DeepSeek V3 0324).

		Просто напиши мне сообщение, а я отвечу с помощью нейросети.

		<b>Команды:</b>
		/reset - сбросить историю диалога
		/help - помощь
		/about - о боте
	`, telebot.ModeHTML)
}

func (h *TelegramHandler) HandleHelp(c telebot.Context) error {
	return c.Send("Помощи пока не будет, но ты потерпи")
}

func (h *TelegramHandler) HandleAbout(c telebot.Context) error {
	return c.Send("Тут будет инфа о боте")
}

func (h *TelegramHandler) HandleReset(c telebot.Context) error { // Clearing history
	userID := c.Sender().ID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.Neural.ResetConversation(ctx, userID); err != nil {
		h.Logger.Printf("Reset error for user %d: %v", userID, err)
		return c.Send("⚠️ Не удалось сбросить историю диалога")
	}

	return c.Send("✅ История диалога успешно сброшена")
}

func (h *TelegramHandler) HandleText(c telebot.Context) error {
	user := c.Sender()

	allowed, waitTime, err := h.checkRateLimit(user.ID)
	if err != nil {
		h.Logger.Printf("Redis error for user %d: %v", user.ID, err)
		// In case of a Redis error, we skip the check so as not to block users
		return h.processMessage(c)
	}

	if !allowed {
		h.Logger.Printf("Rate limit for user %d (wait %.1fs)", user.ID, waitTime.Seconds())
		return c.Send(fmt.Sprintf("⏳ Пожалуйста, подождите %.0f секунд перед следующим сообщением.", waitTime.Seconds()))
	}

	return h.processMessage(c)
}

func (h *TelegramHandler) processMessage(c telebot.Context) error {
	// Text message processing logic
	startTime := time.Now()
	user := c.Sender()
	text := c.Text()

	h.Logger.Printf("Message from %d (%s): %.100s...", user.ID, user.Username, text)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := c.Notify(telebot.Typing); err != nil {
		h.Logger.Printf("Failed to send typing action: %v", err)
	}

	response, err := h.Neural.HandleMessage(ctx, user.ID, text) // Neural network response to user
	if err != nil {
		h.Logger.Printf("Error from Neural for user %d: %v", user.ID, err)
		return c.Send("⚠️ Произошла ошибка при обработке запроса. Пожалуйста, попробуйте позже.")
	}

	if response == "" { // Due to internal errors or other conditions, the neural network may send an empty response to the user
		h.Logger.Printf("Empty response from Neural for user %d", user.ID)
		return c.Send("🤷 Не получилось сформировать ответ. Попробуйте задать вопрос иначе.")
	}

	cleanText := strings.TrimSpace(response) // The maximum character limit for messages in Telegram is 4000. Cut off the neural network's response if it is too long.
	if len(cleanText) > 4000 {
		cleanText = cleanText[:4000]
	}

	if err := safeSend(c, cleanText); err != nil { // Secure messaging feature
		h.Logger.Printf("Failed to send message to %d: %v", user.ID, err)
		return c.Send("⚠️ Не удалось отправить ответ. Пожалуйста, попробуйте еще раз.")
	}

	h.Logger.Printf("Successfully responded to %d in %v", user.ID, time.Since(startTime))
	return nil
}

func (h *TelegramHandler) checkRateLimit(userID int64) (allowed bool, remaining time.Duration, err error) {
	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:%d", userID) // Create a key to track each user's limit

	set, err := h.Redis.SetNX(ctx, key, "1", h.MsgDelay).Result() // Attempting to install key
	// SetNX - atomic operation "Set if Not eXists"
	// "1" - value (in this case it is not important, only the fact of the key presence is used)
	// h.MsgDelay - key lifetime
	if err != nil { // If there is an error, we allow the request, otherwise, if there is an error, the user will simply be blocked
		return true, 0, err
	}

	if set { // If the key was installed, the limit is not exceeded.
		return true, 0, nil
	}

	ttl, err := h.Redis.TTL(ctx, key).Result() // TTL - we get the remaining lifetime of the key
	// Positive value - how many seconds are left before the key is deleted
	if err != nil {
		return true, 0, err
	}

	if ttl < 0 {
		return true, 0, nil
	}

	return false, ttl, nil // We will only get here if the TTL is positive.
}

func (n *NeuralHandler) cleanUpOldMessages(ctx context.Context, olderThan time.Duration) error {
	interval := fmt.Sprintf("%d minutes", int(olderThan.Minutes()))
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
