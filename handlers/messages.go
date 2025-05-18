package handlers

import (
	"context"
	"strings"
	"time"

	"gopkg.in/telebot.v4"
)

func (h *TelegramHandler) messageRules() string {
	return "<b>❗ Правила использования бота | Дикслеймер</b>\n\nЭтот бот предназначен только для легальных целей. Нарушение правил может привести к блокировке и юридическим последствиям в сторону пользователя. Разработчик (@wnderbin) не несет ответственности за неправомерные и незаконные действия пользователей.\n\n<b>Вы автоматически соглашаетесь с диклеймером, при использовании бота.</b>\n\n<a href=\"https://github.com/wnderbin/QuokkaAI-Bot/tree/main/rules\">Подробнее</a>"
}

func (h *TelegramHandler) messagePolicy() string {
	return "<b>📄 Политика конфиденциальности</b>\n\nЗаявление, в котором указано, как бот собирает о вас данные, как долго и в каком виде он их хранит и использует.\n\n❗<b>Необходимо ознакомиться перед использованием бота!</b>\n\n<a href=\"https://github.com/wnderbin/QuokkaAI-Bot/tree/main/privacy\">Ссылка</a>"
}

func (h *TelegramHandler) messageAbout() string {
	return "🚀 <b>Quokka-Bot - Телеграм бот с интеграцией DeepSeekAPI.</b>\n\n<b>В этом боте используется гибкая модель DeepSeek V3 0324.</b>\n\n<b>Ключевые достоинства модели:</b>\n<b>1.</b> Глубокое понимание контекста.\n<b>2.</b> Лучшая структурированность ответов.\n<b>3.</b> API с низкой задержкой - это значит, что модель 'думает' и отвечает на запросы быстрее.\n<b>4.</b> Минимальный \"hallucination\". (меньше выдуманных фактов)\n\nРазработчик: @wnderbin"
}

func (h *TelegramHandler) messageHelp() string {
	return "<b>❓ Помощь:</b>\n<a href=\"https://github.com/wnderbin/QuokkaAI-Bot/tree/main/docs\">Документация</a> - имеются русская и английская версии. В ней изложена вся работа с ботом для пользователей и разработчиков.\n\nЕсли у вас возникла ошибка - это нормально, вероятно перегружены сервера. \n\n❗ Но если вы сталкиваетесь с одной и той же ошибкой несколько раз подряд, пожалуйста обратитесь с проблемой мне в личку - @wnderbin"
}

func (h *TelegramHandler) messageReset(user *telebot.User) string {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := h.Neural.ResetConversation(ctx, user.ID); err != nil {
		h.Logger.Printf("[ ERROR ] Reset error for user %d %s: %v", user.ID, user.Username, err)
		return "⚠️ Не удалось сбросить историю диалога"
	}
	return "✅ История диалога успешно сброшена"
}

func (h *TelegramHandler) messageStart() string {
	return "<b>👋 Приветствую!</b> Я бот с интеграцией DeepSeek AI (DeepSeek V3 0324)\n\nПросто напиши мне любой интересующий тебя запрос, а я на него отвечу при помощи нейросети :)\n\n❗Перед использованием обязательно ознакомьтесь с политикой конфиденциальности\n\n<b>Команды:</b>\n/rules - Дисклеймер, обязателен к ознакомлению. Вы автоматически соглашаетесь с ним при использовании бота.\n/policy - Политика конфиденциальности. Обязательна к ознакомлению. Вы автоматически соглашаетесь с ней при использовании бота.\n/reset - Сбросить историю диалога\n/help - Помощь\n/about - О боте"
}

func (h *TelegramHandler) processMessage(c telebot.Context) error {
	// Text message processing logic
	startTime := time.Now()
	user := c.Sender()
	text := c.Text()

	h.Logger.Printf("Message from %d %s: %.100s...", user.ID, user.Username, text)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if err := c.Notify(telebot.Typing); err != nil {
		h.Logger.Printf("[ ERROR ] Failed to send typing action %d %s: %v", user.ID, user.Username, err)
	}

	response, err := h.Neural.HandleMessage(ctx, user.ID, text) // Neural network response to user
	if err != nil {
		h.Logger.Printf("[ ERROR ] Error from Neural for user %d: %v", user.ID, err)
		return c.Send("⚠️ Произошла ошибка при обработке запроса. Пожалуйста, попробуйте позже.")
	}

	if response == "" { // Due to internal errors or other conditions, the neural network may send an empty response to the user
		h.Logger.Printf("[ ERROR ] Empty response from Neural for user %d", user.ID)
		return c.Send("🤷 Не получилось сформировать ответ. Возможно, сервера перегружены. Можете попробовать задать вопрос иначе.")
	}

	cleanText := strings.TrimSpace(response) // The maximum character limit for messages in Telegram is 4000. Cut off the neural network's response if it is too long.
	if len(cleanText) > 4000 {
		cleanText = cleanText[:4000]
	}

	if err := safeSend(c, cleanText); err != nil { // Secure messaging feature
		h.Logger.Printf("[ ERROR ] Failed to send message to %d: %v", user.ID, err)
		return c.Send("⚠️ Не удалось отправить ответ. Пожалуйста, попробуйте еще раз.")
	}

	h.Logger.Printf("Successfully responded to %d in %v", user.ID, time.Since(startTime))
	return nil
}
