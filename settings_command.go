package starboard

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
)

const (
	minStarsFloor             = 1.0
	defaultSettingsTitle      = "Settings"
	updatedSettingsTitle      = "Updated settings"
	configErrorMessage        = "Couldn't get server config"
	updateConfigErrorMessage  = "Couldn't update server config"
	channelLookupError        = "Couldn't find that channel"
	unknownSubcommandError    = "Unrecognized settings subcommand"
	missingStarsOptionError   = "Missing stars option"
	missingChannelOptionError = "Missing channel option"
)

type GuildSettingsStore interface {
	GetGuild(string) (*GuildSettings, error)
	UpdateGuild(*GuildSettings) error
}

type SettingsContext interface {
	GuildID() string
	Subcommand() string
	StarsOption() (int64, bool)
	ChannelOptionID() (string, bool)
	RespondMessage(string)
	RespondEmbed(*discordgo.MessageEmbed)
	RespondEphemeralEmbed(*discordgo.MessageEmbed)
}

type SettingsCommand struct {
	store  GuildSettingsStore
	logger mio.Logger
}

func NewSettingsCommand(store GuildSettingsStore, logger mio.Logger) *SettingsCommand {
	return &SettingsCommand{
		store:  store,
		logger: logger,
	}
}

func (c *SettingsCommand) Handle(ctx SettingsContext) {
	settings, err := c.store.GetGuild(ctx.GuildID())
	if err != nil {
		ctx.RespondMessage(configErrorMessage)
		c.logger.Error(configErrorMessage, "error", err)
		return
	}

	switch ctx.Subcommand() {
	case "set":
		c.handleSet(ctx, settings)
	case "view", "":
		ctx.RespondEmbed(c.buildSettingsEmbed(defaultSettingsTitle, settings))
	default:
		ctx.RespondMessage(unknownSubcommandError)
	}
}

func (c *SettingsCommand) handleSet(ctx SettingsContext, settings *GuildSettings) {
	stars, ok := ctx.StarsOption()
	if !ok {
		ctx.RespondMessage(missingStarsOptionError)
		return
	}

	channelID, ok := ctx.ChannelOptionID()
	if !ok {
		ctx.RespondMessage(missingChannelOptionError)
		return
	}

	if channelID == "" {
		ctx.RespondMessage(channelLookupError)
		return
	}

	settings.MinStars = int(stars)
	settings.StarboardChannelID = channelID

	if err := c.store.UpdateGuild(settings); err != nil {
		ctx.RespondMessage(updateConfigErrorMessage)
		c.logger.Error(updateConfigErrorMessage, "error", err)
		return
	}

	ctx.RespondEphemeralEmbed(c.buildSettingsEmbed(updatedSettingsTitle, settings))
}

func (c *SettingsCommand) buildSettingsEmbed(title string, settings *GuildSettings) *discordgo.MessageEmbed {
	channelFieldStr := "Not set"
	if settings.StarboardChannelID != "" {
		channelFieldStr = fmt.Sprintf("<#%v>", settings.StarboardChannelID)
	}

	return builders.NewEmbedBuilder().
		WithTitle(title).
		WithOkColor().
		AddField("Stars required", fmt.Sprint(settings.MinStars), true).
		AddField("Starboard channel", channelFieldStr, true).
		Build()
}

type discordSettingsContext struct {
	cmd *discord.DiscordApplicationCommand
}

func newDiscordSettingsContext(cmd *discord.DiscordApplicationCommand) SettingsContext {
	return &discordSettingsContext{cmd: cmd}
}

func (c *discordSettingsContext) GuildID() string {
	return c.cmd.GuildID()
}

func (c *discordSettingsContext) Subcommand() string {
	if len(c.cmd.Data.Options) == 0 {
		return ""
	}
	return c.cmd.Data.Options[0].Name
}

func (c *discordSettingsContext) StarsOption() (int64, bool) {
	opt, ok := c.cmd.Options("set:stars")
	if !ok {
		return 0, false
	}
	return opt.IntValue(), true
}

func (c *discordSettingsContext) ChannelOptionID() (string, bool) {
	opt, ok := c.cmd.Options("set:channel")
	if !ok {
		return "", false
	}
	channel := opt.ChannelValue(nil)
	if channel == nil {
		return "", true
	}
	return channel.ID, true
}

func (c *discordSettingsContext) RespondMessage(msg string) {
	_ = c.cmd.Respond(msg)
}

func (c *discordSettingsContext) RespondEmbed(embed *discordgo.MessageEmbed) {
	_ = c.cmd.RespondEmbed(embed)
}

func (c *discordSettingsContext) RespondEphemeralEmbed(embed *discordgo.MessageEmbed) {
	resp := &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
		Flags:  discordgo.MessageFlagsEphemeral,
	}
	_ = c.cmd.RespondComplex(resp, discordgo.InteractionResponseChannelMessageWithSource)
}
