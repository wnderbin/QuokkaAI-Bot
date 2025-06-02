package handlers

import (
	"context"
	"fmt"
	"time"
)

func (h *TelegramHandler) checkRateLimitCommand(userID int64) (allowed bool, remaining time.Duration, err error) {
	ctx := context.Background()
	key := fmt.Sprintf("com_rate_limit:%d", userID)

	set, err := h.Redis.SetNX(ctx, key, "1", h.ComDelay).Result()
	// SetNX - atomic operation "Set if Not eXists"
	// "1" - value (in this case it is not important, only the fact of the key presence is used)
	// h.ComDelay - key lifetime
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

func (h *TelegramHandler) checkRateLimitMessage(userID int64) (allowed bool, remaining time.Duration, err error) {
	ctx := context.Background()
	key := fmt.Sprintf("mes_rate_limit:%d", userID) // Create a key to track each user's limit

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
