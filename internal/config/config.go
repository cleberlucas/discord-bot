package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Token  string
	Prefix string
	Name   string
}

func Load() (Config, error) {
	token := strings.TrimSpace(os.Getenv("DISCORD_TOKEN"))
	if token == "" {
		return Config{}, fmt.Errorf("DISCORD_TOKEN is required")
	}

	prefix := strings.TrimSpace(os.Getenv("COMMAND_PREFIX"))
	if prefix == "" {
		prefix = "!"
	}

	name := strings.TrimSpace(os.Getenv("BOT_NAME"))
	if name == "" {
		name = "discord-music-bot"
	}

	return Config{
		Token:  token,
		Prefix: prefix,
		Name:   name,
	}, nil
}
