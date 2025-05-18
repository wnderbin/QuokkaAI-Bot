package handlers

import (
	"fmt"

	"gopkg.in/telebot.v4"
)

func (h *TelegramHandler) HandleText(c telebot.Context) error {
	user := c.Sender()

	allowed, waitTime, err := h.checkRateLimitMessage(user.ID)
	if err != nil {
		h.Logger.Printf("[ ERROR ] Redis error for user %d %s: %v", user.ID, user.Username, err)
		// In case of a Redis error, we skip the check so as not to block users
		return h.processMessage(c)
	}

	if !allowed {
		h.Logger.Printf("Rate limit for user %d %s (wait %.1fs)", user.ID, user.Username, waitTime.Seconds())
		return c.Send(fmt.Sprintf("⏳ Пожалуйста, подождите %.0f секунд перед следующим запросом", waitTime.Seconds()))
	}

	return h.processMessage(c)
}

func (h *TelegramHandler) HandleStart(c telebot.Context) error {
	user := c.Sender()
	h.Logger.Printf("Start message from user %d %s", user.ID, user.Username)
	allowed, waitTime, err := h.checkRateLimitCommand(user.ID)
	if err != nil {
		h.Logger.Printf("[ ERROR ] Redis error for user %d %s: %v", user.ID, user.Username, err)
		// In case of a Redis error, we skip the check so as not to block users
		return c.Send(h.messageStart(), telebot.ModeHTML)
	}

	if !allowed {
		h.Logger.Printf("Rate limit for user %d %s (wait %.1fs)", user.ID, user.Username, waitTime.Seconds())
		return c.Send(fmt.Sprintf("⏳ Пожалуйста, подождите %.0f секунд перед следующей командой", waitTime.Seconds()))
	}

	return c.Send(h.messageStart(), telebot.ModeHTML)
}

func (h *TelegramHandler) HandleReset(c telebot.Context) error { // Clearing history
	user := c.Sender()
	h.Logger.Printf("Reset message from user %d %s", user.ID, user.Username)
	allowed, waitTime, err := h.checkRateLimitCommand(user.ID)
	if err != nil {
		h.Logger.Printf("[ ERROR ] Redis error for user %d %s: %v", user.ID, user.Username, err)
		return c.Send(h.messageReset(user))
	}
	if !allowed {
		h.Logger.Printf("Rate limit for user %d %s (wait %.1fs)", user.ID, user.Username, waitTime.Seconds())
		return c.Send(fmt.Sprintf("⏳ Пожалуйста, подождите %.0f секунд перед следующей командой", waitTime.Seconds()))
	}
	return c.Send(h.messageReset(user))
}

func (h *TelegramHandler) HandleHelp(c telebot.Context) error {
	user := c.Sender()
	h.Logger.Printf("Help message from user %d %s", user.ID, user.Username)
	allowed, waitTime, err := h.checkRateLimitCommand(user.ID)
	if err != nil {
		h.Logger.Printf("[ ERROR ] Redis error for user %d %s: %v", user.ID, user.Username, err)
		c.Send(h.messageHelp(), telebot.ModeHTML)
	}
	if !allowed {
		h.Logger.Printf("Rate limit for user %d %s (wait %.1fs)", user.ID, user.Username, waitTime.Seconds())
		return c.Send(fmt.Sprintf("⏳ Пожалуйста, подождите %.0f секунд перед следующей командой", waitTime.Seconds()))
	}
	return c.Send(h.messageHelp(), telebot.ModeHTML)
}

func (h *TelegramHandler) HandleAbout(c telebot.Context) error {
	user := c.Sender()
	h.Logger.Printf("About message from user %d %s", user.ID, user.Username)
	allowed, waitTime, err := h.checkRateLimitCommand(user.ID)
	if err != nil {
		h.Logger.Printf("[ ERROR ] Redis error for user %d %s: %v", user.ID, user.Username, err)
		return c.Send(h.messageAbout(), telebot.ModeHTML)
	}
	if !allowed {
		h.Logger.Printf("Rate limit for user %d %s (wait %.1fs)", user.ID, user.Username, waitTime.Seconds())
		return c.Send(fmt.Sprintf("⏳ Пожалуйста, подождите %.0f секунд перед следующей командой", waitTime.Seconds()))
	}
	return c.Send(h.messageAbout(), telebot.ModeHTML)
}

func (h *TelegramHandler) HandlePolicy(c telebot.Context) error {
	user := c.Sender()
	h.Logger.Printf("Policy message from user %d %s", user.ID, user.Username)
	allowed, waitTime, err := h.checkRateLimitCommand(user.ID)
	if err != nil {
		h.Logger.Printf("[ ERROR ] Redis error for user %d %s: %v", user.ID, user.Username, err)
		return c.Send(h.messagePolicy(), telebot.ModeHTML)
	}
	if !allowed {
		h.Logger.Printf("Rate limit for user %d %s (wait %.1fs)", user.ID, user.Username, waitTime.Seconds())
		return c.Send(fmt.Sprintf("⏳ Пожалуйста, подождите %.0f секунд перед следующей командой", waitTime.Seconds()))
	}
	return c.Send(h.messagePolicy(), telebot.ModeHTML)
}

func (h *TelegramHandler) HandleRules(c telebot.Context) error {
	user := c.Sender()
	h.Logger.Printf("Rules message from user %d %s", user.ID, user.Username)
	allowed, waitTime, err := h.checkRateLimitCommand(user.ID)
	if err != nil {
		h.Logger.Printf("[ ERROR ] Redis error for user %d %s: %v", user.ID, user.Username, err)
		return c.Send(h.messageRules(), telebot.ModeHTML)
	}
	if !allowed {
		h.Logger.Printf("Rate limit for user %d %s (wait %.1fs)", user.ID, user.Username, waitTime.Seconds())
		return c.Send(fmt.Sprintf("⏳ Пожалуйста, подождите %.0f секунд перед следующей командой", waitTime.Seconds()))
	}
	return c.Send(h.messageRules(), telebot.ModeHTML)
}
