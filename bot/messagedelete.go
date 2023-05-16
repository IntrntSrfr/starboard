package bot

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func (b *Bot) messageDeleteHandler(s *discordgo.Session, m *discordgo.MessageDelete) {
	star, err := b.db.GetStar(m.ID)
	if err != nil {
		b.logger.Error("could not get star", zap.Error(err), zap.String("event", "message delete"), zap.Any("message", m))
		return
	}

	s.ChannelMessageDelete(star.StarboardChannelID, star.StarboardMsgID)
	err = b.db.DeleteStar(m.ID)
	if err != nil {
		b.logger.Error("could not delete star", zap.Error(err), zap.String("event", "message delete"), zap.Any("message", m))
		return
	}
}
