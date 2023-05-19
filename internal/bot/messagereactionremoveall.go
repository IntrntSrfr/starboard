package bot

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func messageReactionRemoveAllHandler(b *Bot) func(s *discordgo.Session, r *discordgo.MessageReactionRemoveAll) {
	return func(s *discordgo.Session, r *discordgo.MessageReactionRemoveAll) {
		star, err := b.db.GetStar(r.MessageID)
		if err != nil {
			if err != sql.ErrNoRows {
				b.logger.Error("could not get star", zap.Error(err), zap.Any("react remove all", r))
			}
			return
		}
		_ = s.ChannelMessageDelete(star.StarboardChannelID, star.StarboardMsgID)
		if err = b.db.DeleteStar(r.MessageID); err != nil {
			b.logger.Error("could not delete star", zap.Error(err), zap.Any("react remove all", r))
		}
	}
}
