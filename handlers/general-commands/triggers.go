package gencommands

import (
	"regexp"
	"strconv"
	"strings"

	tools "maquiaBot/tools"

	"github.com/bwmarrin/discordgo"
)

// Triggers list out the triggers enabled in the server
func Triggers(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Get server
	server, err := s.Guild(m.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "This is not a server!")
		return
	}
	serverImg := "https://cdn.discordapp.com/icons/" + server.ID + "/" + server.Icon
	if strings.Contains(server.Icon, "a_") {
		serverImg += ".gif"
	} else {
		serverImg += ".png"
	}

	serverData := tools.GetServer(*server, s)
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    server.Name,
			IconURL: serverImg,
		},
	}

	if len(serverData.Triggers) == 0 {
		s.ChannelMessageSend(m.ChannelID, "There are no triggers configuered for this server currently! Admins can see `help trigger` for details on how to add triggers.")
		return
	}

	for _, trigger := range serverData.Triggers {
		trigger.Cause = `(?i)` + trigger.Cause
		regex := false
		_, err := regexp.Compile(trigger.Cause)
		if err == nil {
			regex = true
		}
		valueText := "Trigger: " + trigger.Cause + "\nResult: " + trigger.Result + "\nRegex compatible: " + strconv.FormatBool(regex)
		if len(valueText) > 1024 {
			valueText = valueText[:1021] + "..."
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  strconv.FormatInt(trigger.ID, 10),
			Value: valueText,
		})

		if len(embed.Fields) == 25 {
			if len(serverData.Triggers) > 25 {
				embed.Footer = &discordgo.MessageEmbedFooter{
					Text: "Page 1",
				}
			}
			break
		}
	}
	msg, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: "Admins can use `trigger -d` to delete a trigger! If there are more than 25 triggers, please use the reactions to go through pages!",
		Embed:   embed,
	})
	if err != nil {
		return
	}
	if len(embed.Fields) == 25 && len(serverData.Triggers) > 25 {
		_ = s.MessageReactionAdd(m.ChannelID, msg.ID, "➡️")
	}
	return
}
