package bot

import (
	"context"
	"errors"
	"strings"
)

var ErrEmptyQuery = errors.New("query cannot be empty")

type Source interface {
	Resolve(ctx context.Context, query string) (Track, error)
}

type QuerySource struct{}

func (QuerySource) Resolve(ctx context.Context, query string) (Track, error) {
	_ = ctx

	cleaned := strings.TrimSpace(query)
	if cleaned == "" {
		return Track{}, ErrEmptyQuery
	}

	return Track{
		Title: cleaned,
		Query: cleaned,
	}, nil
}
