package bot

import (
	"fmt"
	"strings"
)

type Command struct {
	Name string
	Args string
}

func ParseCommand(prefix, content string) (Command, bool) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" || prefix == "" || !strings.HasPrefix(trimmed, prefix) {
		return Command{}, false
	}

	body := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
	if body == "" {
		return Command{}, false
	}

	fields := strings.Fields(body)
	name := strings.ToLower(fields[0])
	args := strings.TrimSpace(body[len(fields[0]):])

	return Command{Name: name, Args: strings.TrimSpace(args)}, true
}

func HelpMessage(prefix string) string {
	return fmt.Sprintf(
		"Available commands:\n%splay <query> - add a track to the queue\n%squeue - show the current queue\n%spause - pause playback\n%sresume - resume playback\n%sskip - skip the current track\n%sleave - clear the server state\n%sping - check whether the bot is online",
		prefix,
		prefix,
		prefix,
		prefix,
		prefix,
		prefix,
		prefix,
	)
}

func FormatQueue(state PlaybackState) string {
	if state.Current == nil && len(state.Queue) == 0 {
		return "Queue is empty."
	}

	var builder strings.Builder

	if state.Current != nil {
		builder.WriteString("Now playing: ")
		builder.WriteString(state.Current.DisplayName())
		builder.WriteString("\n")
	}

	if len(state.Queue) > 0 {
		builder.WriteString("Next tracks:\n")
		for index, track := range state.Queue {
			builder.WriteString(fmt.Sprintf("%d. %s\n", index+1, track.DisplayName()))
		}
	}

	if state.Paused {
		builder.WriteString("Status: paused.\n")
	}

	return strings.TrimSpace(builder.String())
}
