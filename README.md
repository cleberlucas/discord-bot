# Discord Music Bot

A Go and Docker scaffold for a Discord music bot.

This scaffold uses a queue layer and text commands with `discordgo`. The audio layer is intentionally source-agnostic for safety and to avoid unauthorized YouTube Music extraction dependencies. You can plug in an authorized provider or direct streams you control later.

## Included

- Go project structure with `cmd/` and `internal/`
- Environment-based configuration
- Basic commands: `play`, `queue`, `pause`, `resume`, `skip`, `leave`, and `ping`
- Dockerfile and Docker Compose support for build and runtime
- A clean base for adding an authorized audio adapter later

## Requirements

- Go 1.22 or newer
- Docker, if you want to run in a container
- A Discord bot token
- Message Content intent enabled in the Discord developer portal if you use prefix commands

## Setup

1. Export the environment variables:

```bash
export DISCORD_TOKEN="your-token"
export COMMAND_PREFIX="!"
export BOT_NAME="discord-music-bot"
```

2. Run locally:

```bash
go mod tidy
go run ./cmd/bot
```

3. Build the Docker image or start it with Compose:

```bash
docker build -t discord-music-bot .
docker compose up --build
```

## Commands

- `!play <query>` adds a track to the queue
- `!queue` shows the current queue
- `!pause` pauses playback
- `!resume` resumes playback
- `!skip` skips the current track
- `!leave` clears the server state
- `!ping` replies with pong

## YouTube Music Note

This scaffold does not implement YouTube Music extraction, downloads, or protection bypass. If you need real playback, connect an authorized provider or a source of audio you control.
