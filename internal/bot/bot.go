package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"example.com/discord-music-bot/internal/config"
)

type Bot struct {
	cfg     config.Config
	logger  *log.Logger
	session *discordgo.Session
	player  Player
	source  Source
}

func New(cfg config.Config, logger *log.Logger) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsMessageContent

	bot := &Bot{
		cfg:     cfg,
		logger:  logger,
		session: session,
		player:  NewMemoryPlayer(),
		source:  QuerySource{},
	}

	session.AddHandler(bot.onReady)
	session.AddHandler(bot.onMessageCreate)

	return bot, nil
}

func (b *Bot) Run(ctx context.Context) error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("open discord session: %w", err)
	}
	defer b.session.Close()

	b.logger.Printf("bot %s iniciado com prefixo %q", b.cfg.Name, b.cfg.Prefix)

	<-ctx.Done()
	return nil
}

func (b *Bot) onReady(session *discordgo.Session, ready *discordgo.Ready) {
	if ready == nil || ready.User == nil {
		b.logger.Println("discord ready")
		return
	}

	b.logger.Printf("conectado como %s#%s", ready.User.Username, ready.User.Discriminator)
}

func (b *Bot) onMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message == nil || message.Author == nil || message.Author.Bot {
		return
	}

	command, ok := ParseCommand(b.cfg.Prefix, message.Content)
	if !ok {
		return
	}

	if message.GuildID == "" {
		b.reply(message.ChannelID, "Use este bot em um servidor do Discord.")
		return
	}

	switch command.Name {
	case "ping":
		b.reply(message.ChannelID, "pong")
	case "help":
		b.reply(message.ChannelID, HelpMessage(b.cfg.Prefix))
	case "play":
		b.handlePlay(message, command)
	case "queue":
		state := b.player.Snapshot(message.GuildID)
		b.reply(message.ChannelID, FormatQueue(state))
	case "pause":
		state, err := b.player.Pause(message.GuildID)
		if err != nil {
			b.reply(message.ChannelID, err.Error())
			return
		}
		b.reply(message.ChannelID, fmt.Sprintf("Execucao pausada.\n%s", FormatQueue(state)))
	case "resume":
		state, err := b.player.Resume(message.GuildID)
		if err != nil {
			b.reply(message.ChannelID, err.Error())
			return
		}
		b.reply(message.ChannelID, fmt.Sprintf("Execucao retomada.\n%s", FormatQueue(state)))
	case "skip":
		state, err := b.player.Skip(message.GuildID)
		if err != nil {
			b.reply(message.ChannelID, err.Error())
			return
		}
		b.reply(message.ChannelID, fmt.Sprintf("Faixa pulada.\n%s", FormatQueue(state)))
	case "leave":
		b.player.Leave(message.GuildID)
		b.reply(message.ChannelID, "Estado do servidor removido.")
	default:
		b.reply(message.ChannelID, fmt.Sprintf("Comando desconhecido.\n%s", HelpMessage(b.cfg.Prefix)))
	}
}

func (b *Bot) handlePlay(message *discordgo.MessageCreate, command Command) {
	query := strings.TrimSpace(command.Args)
	if query == "" {
		b.reply(message.ChannelID, fmt.Sprintf("Uso: %splay <consulta>", b.cfg.Prefix))
		return
	}

	track, err := b.source.Resolve(context.Background(), query)
	if err != nil {
		b.reply(message.ChannelID, err.Error())
		return
	}

	track.RequestedBy = message.Author.Username
	state := b.player.Enqueue(message.GuildID, track)

	if state.Current != nil && state.Current.DisplayName() == track.DisplayName() && len(state.Queue) == 0 {
		b.reply(message.ChannelID, fmt.Sprintf("Faixa selecionada: %s\nA integracao de audio ainda esta no modo base para adaptadores autorizados.", track.DisplayName()))
		return
	}

	b.reply(message.ChannelID, fmt.Sprintf("Adicionado a fila: %s\n%s", track.DisplayName(), FormatQueue(state)))
}

func (b *Bot) reply(channelID, content string) {
	if strings.TrimSpace(content) == "" {
		return
	}

	if _, err := b.session.ChannelMessageSend(channelID, content); err != nil && b.logger != nil {
		b.logger.Printf("send message failed: %v", err)
	}
}
