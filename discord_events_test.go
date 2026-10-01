package starboard

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestFixTenorURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "non-tenor url left untouched",
			input:    "https://example.com/image.png",
			expected: "https://example.com/image.png",
		},
		{
			name:     "tenor thumbnail converted",
			input:    "https://media.tenor.com/abcd/AAAAe/efgh.png",
			expected: "https://media.tenor.com/abcd/AAAAC/efgh.gif",
		},
		{
			name:     "tenor still works with jpg",
			input:    "https://media.tenor.com/xyz/AAAAe/qwerty.jpg",
			expected: "https://media.tenor.com/xyz/AAAAC/qwerty.gif",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fixTenorURL(tt.input)
			if got != tt.expected {
				t.Fatalf("expected %q got %q", tt.expected, got)
			}
		})
	}
}

func TestGetReactionCount(t *testing.T) {
	msg := &discordgo.Message{
		Reactions: []*discordgo.MessageReactions{
			{Emoji: &discordgo.Emoji{Name: "⭐"}, Count: 4},
			{Emoji: &discordgo.Emoji{Name: "👍"}, Count: 2},
		},
	}

	if got := getReactionCount(msg, "⭐"); got != 4 {
		t.Fatalf("expected star count 4 got %d", got)
	}

	if got := getReactionCount(msg, "❌"); got != 0 {
		t.Fatalf("expected missing emoji count 0 got %d", got)
	}
}

func TestExtractImageURLsAndInfo(t *testing.T) {
	msg := &discordgo.Message{
		Embeds: []*discordgo.MessageEmbed{
			{Thumbnail: &discordgo.MessageEmbedThumbnail{URL: "https://media.tenor.com/foo/AAAAe/bar.png"}},
			{Image: &discordgo.MessageEmbedImage{URL: "https://example.com/embed.png"}},
		},
		Attachments: []*discordgo.MessageAttachment{
			{Filename: "pic.gif", URL: "https://example.com/attach.gif", ContentType: "image/gif"},
			{Filename: "doc.pdf", URL: "https://example.com/doc.pdf", ContentType: "application/pdf"},
		},
	}

	urls, extra := extractImageURLsAndInfo(msg)

	if len(urls) != 3 {
		t.Fatalf("expected 3 urls got %d", len(urls))
	}

	expectedFirst := "https://media.tenor.com/foo/AAAAC/bar.gif"
	if urls[0] != expectedFirst {
		t.Fatalf("expected tenor url %q got %q", expectedFirst, urls[0])
	}

	if urls[2] != "https://example.com/attach.gif" {
		t.Fatalf("expected attachment image url, got %q", urls[2])
	}

	if want := "\n📎 [pic.gif](https://example.com/attach.gif)"; !strings.Contains(extra, want) {
		t.Fatalf("expected extra content to contain %q, got %q", want, extra)
	}

	if want := "\n📎 [doc.pdf](https://example.com/doc.pdf)"; !strings.Contains(extra, want) {
		t.Fatalf("expected extra content to contain %q, got %q", want, extra)
	}
}

func TestBuildStarboardEmbeds(t *testing.T) {
	msg := &discordgo.Message{
		ID:      "123",
		Content: "Hello world",
		Author: &discordgo.User{
			ID:            "user",
			Username:      "Tester",
			Discriminator: "0001",
		},
		Embeds: []*discordgo.MessageEmbed{
			{Image: &discordgo.MessageEmbedImage{URL: "https://example.com/base.png"}},
		},
		Attachments: []*discordgo.MessageAttachment{
			{Filename: "extra.png", URL: "https://example.com/extra.png", ContentType: "image/png"},
			{Filename: "notes.txt", URL: "https://example.com/notes.txt", ContentType: "text/plain"},
		},
	}

	guild := &discordgo.Guild{ID: "guild"}
	channel := &discordgo.Channel{ID: "chan", Name: "general"}

	embeds := buildStarboardEmbeds(msg, 5, guild, channel)

	if len(embeds) != 2 {
		t.Fatalf("expected 2 embeds got %d", len(embeds))
	}

	first := embeds[0]

	if first.Author == nil || first.Author.Name != "Tester#0001 - ⭐ 5" {
		t.Fatalf("unexpected author info: %+v", first.Author)
	}

	if first.Footer == nil || first.Footer.Text != "#general" {
		t.Fatalf("unexpected footer: %+v", first.Footer)
	}

	if first.Description == "" || !strings.Contains(first.Description, "Hello world") {
		t.Fatalf("description missing message content: %q", first.Description)
	}

	expectedJump := "https://discord.com/channels/guild/chan/123"
	if !strings.Contains(first.Description, expectedJump) {
		t.Fatalf("expected jump link in description: %q", first.Description)
	}

	if first.Image == nil || first.Image.URL != "https://example.com/base.png" {
		t.Fatalf("first embed image mismatch: %+v", first.Image)
	}

	if second := embeds[1]; second.Image == nil || second.Image.URL != "https://example.com/extra.png" {
		t.Fatalf("second embed image mismatch: %+v", second.Image)
	}

	if want := "\n📎 [notes.txt](https://example.com/notes.txt)"; !strings.Contains(first.Description, want) {
		t.Fatalf("expected attachment listing %q in description: %q", want, first.Description)
	}
}
