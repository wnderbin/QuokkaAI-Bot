package utils

import (
	"log"
	"os"
)

func NewLogger(debug bool) *log.Logger {
	return log.New(os.Stdout, "BOT: ", log.LstdFlags|log.Lshortfile)
}
