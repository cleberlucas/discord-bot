package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	musicbot "example.com/discord-music-bot/internal/bot"
	"example.com/discord-music-bot/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	bot, err := musicbot.New(cfg, logger)
	if err != nil {
		log.Fatalf("create bot: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := bot.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Fatalf("run bot: %v", err)
	}
}
