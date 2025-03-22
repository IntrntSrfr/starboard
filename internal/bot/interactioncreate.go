package bot

import (
	"errors"
	"fmt"
	"strings"

	"github.com/intrntsrfr/meido/pkg/utils/builders"
	"go.uber.org/zap"

	"github.com/bwmarrin/discordgo"
)

func textResponse(content string) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         content,
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	}
}

func defaultErrorResponse() *discordgo.InteractionResponse {
	return textResponse("There was an issue, please try again!")
}

func embedResponse(embed *discordgo.MessageEmbed) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:          []*discordgo.MessageEmbed{embed},
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	}
}

func respondWithError(b *Bot, s *discordgo.Session, i *discordgo.Interaction, content string) error {
	b.logger.Info("responded with error", zap.String("error", content))
	return s.InteractionRespond(i, textResponse(content))
}

func interactionCreate(b *Bot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			data := i.ApplicationCommandData()
			switch data.Name {
			case "help":
				helpCommand(b, s, i.Interaction, &data)
			case "settings":
				settingsCommand(b, s, i.Interaction, &data)
			}
		}
	}
}

func helpCommand(b *Bot, s *discordgo.Session, i *discordgo.Interaction, data *discordgo.ApplicationCommandInteractionData) {
	text := strings.Builder{}
	text.WriteString("There are two settings you can change")
	text.WriteString("\n - Starboard channel")
	text.WriteString("\n - Minimum required stars for a post to be posted to starboard")
	text.WriteString("\n\t - The minimum amount you can set is 1")
	text.WriteString("\nTwo examples: ")
	text.WriteString("\n`/settings edit channel` - Edit the Starboard channel")
	text.WriteString("\n`/settings edit minstars 3` - Edit the minimum amount of reactions to appear on Starboard")
	_ = s.InteractionRespond(i, textResponse(text.String()))
}

func settingsCommand(b *Bot, s *discordgo.Session, i *discordgo.Interaction, data *discordgo.ApplicationCommandInteractionData) {
	b.logger.Info("new settings command")
	var (
		resp *discordgo.InteractionResponse
		err  error
	)
	if i.GuildID == "" {
		_ = respondWithError(b, s, i, "This can only be used in a server!")
		return
	}
	if len(data.Options) < 1 {
		_ = respondWithError(b, s, i, "There was an issue, please try again!")
		return
	}
	switch data.Options[0].Name {
	case "view":
		resp, err = settingsViewCommand(b, s, i, data)
	case "edit":
		resp, err = settingsEditCommand(b, s, i, data)
	}
	if err != nil {
		b.logger.Error("could not respond", zap.Error(err))
		_ = s.InteractionRespond(i, resp)
		return
	}
	_ = s.InteractionRespond(i, resp)
}

func settingsViewCommand(b *Bot, s *discordgo.Session, i *discordgo.Interaction, data *discordgo.ApplicationCommandInteractionData) (*discordgo.InteractionResponse, error) {
	b.logger.Info("new settings view command")
	gs, err := b.db.GetGuild(i.GuildID)
	if err != nil {
		return defaultErrorResponse(), err
	}
	g, err := s.State.Guild(i.GuildID)
	if err != nil {
		return defaultErrorResponse(), err
	}
	sbChannel := "Unset"
	if gs.StarboardChannelID != "" {
		sbChannel = fmt.Sprintf("<#%v>", gs.StarboardChannelID)
	}
	embed := builders.NewEmbedBuilder().
		WithTitle(fmt.Sprintf("Settings for %v", g.Name)).
		WithOkColor().
		WithThumbnail(g.IconURL("512")).
		AddField("Starboard channel", sbChannel, true).
		AddField("Minimum stars", fmt.Sprint(gs.MinStars), true)
	return embedResponse(embed.Build()), nil
}

func settingsEditCommand(b *Bot, s *discordgo.Session, i *discordgo.Interaction, data *discordgo.ApplicationCommandInteractionData) (*discordgo.InteractionResponse, error) {
	b.logger.Info("new settings edit command")
	/*
		if i.Member.Permissions&discordgo.PermissionAdministrator == 0 {
			return textResponse("You need admin perms to do this, sorry!"), nil
		}
	*/
	if len(data.Options[0].Options) < 1 || len(data.Options[0].Options[0].Options) < 1 {
		return defaultErrorResponse(), errors.New("too few options")
	}
	gs, err := b.db.GetGuild(i.GuildID)
	if err != nil {
		return defaultErrorResponse(), err
	}
	g, err := s.State.Guild(i.GuildID)
	if err != nil {
		return defaultErrorResponse(), err
	}
	embed := builders.NewEmbedBuilder().
		WithTitle(fmt.Sprintf("Settings for %v", g.Name)).
		WithOkColor().
		WithThumbnail(g.IconURL("512"))
	switch data.Options[0].Options[0].Name {
	case "channel":
		val := data.Options[0].Options[0].Options[0].StringValue()
		oldVal := gs.StarboardChannelID
		if gs.StarboardChannelID != "" {
			oldVal = fmt.Sprintf("<#%v>", gs.StarboardChannelID)
		}
		gs.StarboardChannelID = val
		err := b.db.UpdateGuild(gs)
		if err != nil {
			b.logger.Error("could not update channel", zap.Any("interaction", i))
			return defaultErrorResponse(), err
		}
		embed.AddField("Updated channel", fmt.Sprintf("%v -> <#%v>", oldVal, val), false)
	case "minimum-stars":
		val := int(data.Options[0].Options[0].Options[0].IntValue())
		oldVal := gs.MinStars
		gs.MinStars = val
		err := b.db.UpdateGuild(gs)
		if err != nil {
			b.logger.Error("could not update min-stars", zap.Any("interaction", i))
			return defaultErrorResponse(), err
		}
		embed.AddField("Updated minimum-stars", fmt.Sprintf("%v -> %v", oldVal, val), false)
	}
	return embedResponse(embed.Build()), nil
}
