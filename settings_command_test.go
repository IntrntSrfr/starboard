package starboard

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
)

func TestSettingsCommand_HandleView(t *testing.T) {
	store := &fakeGuildSettingsStore{
		getGuildFunc: func(id string) (*GuildSettings, error) {
			return &GuildSettings{
				ID:                 id,
				MinStars:           3,
				StarboardChannelID: "123",
			}, nil
		},
	}

	handler := NewSettingsCommand(store, mio.NewDiscardLogger())
	ctx := &fakeSettingsContext{
		guildID:    "guild-1",
		subcommand: "view",
	}

	handler.Handle(ctx)

	if len(ctx.messages) != 0 {
		t.Fatalf("unexpected messages: %v", ctx.messages)
	}
	if len(ctx.embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(ctx.embeds))
	}
	if ctx.embeds[0].Title != defaultSettingsTitle {
		t.Fatalf("expected embed title %q, got %q", defaultSettingsTitle, ctx.embeds[0].Title)
	}
}

func TestSettingsCommand_HandleSetSuccess(t *testing.T) {
	updated := &GuildSettings{}
	store := &fakeGuildSettingsStore{
		getGuildFunc: func(id string) (*GuildSettings, error) {
			return &GuildSettings{
				ID:                 id,
				MinStars:           2,
				StarboardChannelID: "222",
			}, nil
		},
		updateGuildFunc: func(g *GuildSettings) error {
			*updated = *g
			return nil
		},
	}

	handler := NewSettingsCommand(store, mio.NewDiscardLogger())
	ctx := &fakeSettingsContext{
		guildID:    "guild-2",
		subcommand: "set",
		stars:      5,
		starsOK:    true,
		channelID:  "456",
		channelOK:  true,
	}

	handler.Handle(ctx)

	if updated.MinStars != 5 {
		t.Fatalf("expected min stars to be updated to 5, got %d", updated.MinStars)
	}
	if updated.StarboardChannelID != "456" {
		t.Fatalf("expected channel to be updated to 456, got %s", updated.StarboardChannelID)
	}
	if len(ctx.ephemeralEmbeds) != 1 {
		t.Fatalf("expected 1 ephemeral embed, got %d", len(ctx.ephemeralEmbeds))
	}
	if ctx.ephemeralEmbeds[0].Title != updatedSettingsTitle {
		t.Fatalf("expected updated embed title %q, got %q", updatedSettingsTitle, ctx.ephemeralEmbeds[0].Title)
	}
}

func TestSettingsCommand_HandleSetMissingStars(t *testing.T) {
	store := &fakeGuildSettingsStore{
		getGuildFunc: func(id string) (*GuildSettings, error) {
			return &GuildSettings{ID: id}, nil
		},
	}

	handler := NewSettingsCommand(store, mio.NewDiscardLogger())
	ctx := &fakeSettingsContext{
		guildID:    "guild-3",
		subcommand: "set",
		starsOK:    false,
		channelID:  "456",
		channelOK:  true,
	}

	handler.Handle(ctx)

	if len(ctx.messages) != 1 {
		t.Fatalf("expected single message, got %d", len(ctx.messages))
	}
	if ctx.messages[0] != missingStarsOptionError {
		t.Fatalf("expected message %q, got %q", missingStarsOptionError, ctx.messages[0])
	}
}

func TestSettingsCommand_HandleUpdateFailure(t *testing.T) {
	store := &fakeGuildSettingsStore{
		getGuildFunc: func(id string) (*GuildSettings, error) {
			return &GuildSettings{ID: id}, nil
		},
		updateGuildFunc: func(g *GuildSettings) error {
			return errors.New("boom")
		},
	}

	handler := NewSettingsCommand(store, mio.NewDiscardLogger())
	ctx := &fakeSettingsContext{
		guildID:    "guild-4",
		subcommand: "set",
		stars:      4,
		starsOK:    true,
		channelID:  "789",
		channelOK:  true,
	}

	handler.Handle(ctx)

	if len(ctx.messages) != 1 {
		t.Fatalf("expected single message, got %d", len(ctx.messages))
	}
	if ctx.messages[0] != updateConfigErrorMessage {
		t.Fatalf("expected message %q, got %q", updateConfigErrorMessage, ctx.messages[0])
	}
}

func TestSettingsCommand_HandleUnknownSubcommand(t *testing.T) {
	store := &fakeGuildSettingsStore{
		getGuildFunc: func(id string) (*GuildSettings, error) {
			return &GuildSettings{ID: id}, nil
		},
	}

	handler := NewSettingsCommand(store, mio.NewDiscardLogger())
	ctx := &fakeSettingsContext{
		guildID:    "guild-5",
		subcommand: "delete",
	}

	handler.Handle(ctx)

	if len(ctx.messages) != 1 {
		t.Fatalf("expected single message, got %d", len(ctx.messages))
	}
	if ctx.messages[0] != unknownSubcommandError {
		t.Fatalf("expected message %q, got %q", unknownSubcommandError, ctx.messages[0])
	}
}

type fakeGuildSettingsStore struct {
	getGuildFunc    func(string) (*GuildSettings, error)
	updateGuildFunc func(*GuildSettings) error
}

func (f *fakeGuildSettingsStore) GetGuild(id string) (*GuildSettings, error) {
	if f.getGuildFunc != nil {
		return f.getGuildFunc(id)
	}
	return nil, nil
}

func (f *fakeGuildSettingsStore) UpdateGuild(settings *GuildSettings) error {
	if f.updateGuildFunc != nil {
		return f.updateGuildFunc(settings)
	}
	return nil
}

type fakeSettingsContext struct {
	guildID    string
	subcommand string
	stars      int64
	starsOK    bool
	channelID  string
	channelOK  bool

	messages        []string
	embeds          []*discordgo.MessageEmbed
	ephemeralEmbeds []*discordgo.MessageEmbed
}

func (f *fakeSettingsContext) GuildID() string {
	return f.guildID
}

func (f *fakeSettingsContext) Subcommand() string {
	return f.subcommand
}

func (f *fakeSettingsContext) StarsOption() (int64, bool) {
	return f.stars, f.starsOK
}

func (f *fakeSettingsContext) ChannelOptionID() (string, bool) {
	return f.channelID, f.channelOK
}

func (f *fakeSettingsContext) RespondMessage(msg string) {
	f.messages = append(f.messages, msg)
}

func (f *fakeSettingsContext) RespondEmbed(embed *discordgo.MessageEmbed) {
	f.embeds = append(f.embeds, embed)
}

func (f *fakeSettingsContext) RespondEphemeralEmbed(embed *discordgo.MessageEmbed) {
	f.ephemeralEmbeds = append(f.ephemeralEmbeds, embed)
}
