package bot

import "github.com/bwmarrin/discordgo"

func interactionCreate(b *Bot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 0,
			Data: &discordgo.InteractionResponseData{
				TTS:        false,
				Content:    "",
				Components: []discordgo.MessageComponent{},
				Embeds:     []*discordgo.MessageEmbed{},
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse:       []discordgo.AllowedMentionType{},
					Roles:       []string{},
					Users:       []string{},
					RepliedUser: false,
				},
				Files:    []*discordgo.File{},
				Flags:    0,
				Choices:  []*discordgo.ApplicationCommandOptionChoice{},
				CustomID: "",
				Title:    "",
			},
		})

	}
}
