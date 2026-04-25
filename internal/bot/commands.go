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
		"Comandos disponiveis:\n%splay <consulta> - adiciona uma faixa a fila\n%squeue - mostra a fila\n%spause - pausa a execucao\n%sresume - retoma a execucao\n%sskip - pula a faixa atual\n%sleave - limpa o estado do servidor\n%sping - verifica se o bot esta online",
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
		return "Fila vazia."
	}

	var builder strings.Builder

	if state.Current != nil {
		builder.WriteString("Tocando agora: ")
		builder.WriteString(state.Current.DisplayName())
		builder.WriteString("\n")
	}

	if len(state.Queue) > 0 {
		builder.WriteString("Proximas faixas:\n")
		for index, track := range state.Queue {
			builder.WriteString(fmt.Sprintf("%d. %s\n", index+1, track.DisplayName()))
		}
	}

	if state.Paused {
		builder.WriteString("Status: pausado.\n")
	}

	return strings.TrimSpace(builder.String())
}
