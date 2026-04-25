package bot

import "strings"

type Track struct {
	Title       string
	Query       string
	RequestedBy string
}

func (t Track) DisplayName() string {
	if strings.TrimSpace(t.Title) != "" {
		return t.Title
	}

	return t.Query
}

type PlaybackState struct {
	Current *Track
	Queue   []Track
	Paused  bool
}
