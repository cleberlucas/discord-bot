package bot

import (
	"fmt"
	"sync"
)

type Player interface {
	Enqueue(guildID string, track Track) PlaybackState
	Pause(guildID string) (PlaybackState, error)
	Resume(guildID string) (PlaybackState, error)
	Skip(guildID string) (PlaybackState, error)
	Leave(guildID string)
	Snapshot(guildID string) PlaybackState
}

type MemoryPlayer struct {
	mu     sync.Mutex
	guilds map[string]*guildPlayback
}

type guildPlayback struct {
	current *Track
	queue   []Track
	paused  bool
}

func NewMemoryPlayer() *MemoryPlayer {
	return &MemoryPlayer{
		guilds: make(map[string]*guildPlayback),
	}
}

func (p *MemoryPlayer) Enqueue(guildID string, track Track) PlaybackState {
	p.mu.Lock()
	defer p.mu.Unlock()

	guild := p.ensureGuild(guildID)
	copyTrack := track

	if guild.current == nil {
		guild.current = &copyTrack
	} else {
		guild.queue = append(guild.queue, copyTrack)
	}

	return guild.snapshot()
}

func (p *MemoryPlayer) Pause(guildID string) (PlaybackState, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	guild, ok := p.guilds[guildID]
	if !ok || guild.current == nil {
		return PlaybackState{}, fmt.Errorf("no active track in guild %s", guildID)
	}

	guild.paused = true
	return guild.snapshot(), nil
}

func (p *MemoryPlayer) Resume(guildID string) (PlaybackState, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	guild, ok := p.guilds[guildID]
	if !ok || guild.current == nil {
		return PlaybackState{}, fmt.Errorf("no active track in guild %s", guildID)
	}

	guild.paused = false
	return guild.snapshot(), nil
}

func (p *MemoryPlayer) Skip(guildID string) (PlaybackState, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	guild, ok := p.guilds[guildID]
	if !ok || (guild.current == nil && len(guild.queue) == 0) {
		return PlaybackState{}, fmt.Errorf("queue is empty for guild %s", guildID)
	}

	if len(guild.queue) == 0 {
		guild.current = nil
		guild.paused = false
		return guild.snapshot(), nil
	}

	next := guild.queue[0]
	guild.queue = append([]Track(nil), guild.queue[1:]...)
	guild.current = &next
	guild.paused = false
	return guild.snapshot(), nil
}

func (p *MemoryPlayer) Leave(guildID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.guilds, guildID)
}

func (p *MemoryPlayer) Snapshot(guildID string) PlaybackState {
	p.mu.Lock()
	defer p.mu.Unlock()

	guild, ok := p.guilds[guildID]
	if !ok {
		return PlaybackState{}
	}

	return guild.snapshot()
}

func (p *MemoryPlayer) ensureGuild(guildID string) *guildPlayback {
	guild, ok := p.guilds[guildID]
	if !ok {
		guild = &guildPlayback{}
		p.guilds[guildID] = guild
	}

	return guild
}

func (g *guildPlayback) snapshot() PlaybackState {
	state := PlaybackState{
		Paused: g.paused,
	}

	if g.current != nil {
		current := *g.current
		state.Current = &current
	}

	if len(g.queue) > 0 {
		state.Queue = append([]Track(nil), g.queue...)
	}

	return state
}
