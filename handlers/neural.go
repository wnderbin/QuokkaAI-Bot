package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"quokka-ai-bot/config"
	"quokka-ai-bot/models"
	"time"
)

type NeuralHandler struct {
	DeepSeekClient *models.DeepSeekClient // Client for interacting with DeepSeek API
	DB             *sql.DB                // Connecting to a database
}

func NewNeuralHandler(apiKey string, db *sql.DB) *NeuralHandler { // A constructor that creates a new instance of the handler
	return &NeuralHandler{
		DeepSeekClient: models.NewDeepSeekClient(apiKey),
		DB:             db,
	}
}

func (h *NeuralHandler) HandleMessage(ctx context.Context, userID int64, text string) (string, error) { // The main method of message processing
	err := h.saveMessage(ctx, userID, "user", text) // Saves the user's message to the database. This is necessary so that the deepsik can further understand the context of the conversation.
	if err != nil {
		return "", fmt.Errorf("failed to save user message: %w", err)
	}

	messages, err := h.getMessages(ctx, userID, 10)
	if err != nil {
		return "", fmt.Errorf("failed to get conversation history: %w", err)
	}

	request := models.DeepSeekRequest{ // Generates a request to the DeepSeek API with the message history
		Model:    config.Load().DeepSeekModel,
		Messages: messages,
	}

	response, err := h.DeepSeekClient.ChatCompeletion(ctx, request) // Sends a request
	if err != nil {
		return "", fmt.Errorf("deepseek api error: %w", err)
	}

	err = h.saveMessage(ctx, userID, "assistant", response) // We save the answer in the database. It is necessary for understanding the context of the conversation.
	if err != nil {
		return "", fmt.Errorf("failed to save assistant message: %w", err)
	}

	return response, nil
}

func (h *NeuralHandler) ResetConversation(ctx context.Context, userID int64) error { // Deletes all message history for the specified user
	_, err := h.DB.ExecContext(ctx, "DELETE FROM chat_messages WHERE user_id = $1", userID)
	return err
}

func (h *NeuralHandler) saveMessage(ctx context.Context, userID int64, role, content string) error { // Saving a message to the database
	_, err := h.DB.ExecContext(ctx,
		"INSERT INTO chat_messages (user_id, role, content, created_at) VALUES ($1, $2, $3, $4)",
		userID, role, content, time.Now())
	return err
}

func (h *NeuralHandler) getMessages(ctx context.Context, userID int64, limit int) ([]models.Message, error) { // Getting messages in the database
	rows, err := h.DB.QueryContext(ctx,
		`SELECT role, content 
		FROM chat_messages 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.Role, &msg.Content); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Changes the output order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}
