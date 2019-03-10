package handlers

import (
	structs "../structs"
	osucommands "./osu-commands"
	"github.com/bwmarrin/discordgo"
	"github.com/thehowl/go-osuapi"
)

// OsuHandle handles commands that are regarding osu!
func OsuHandle(s *discordgo.Session, m *discordgo.MessageCreate, args []string, osuAPI *osuapi.Client, playerCache []structs.PlayerData, mapCache []structs.MapData) {
	mainArg := args[1]
	switch mainArg {
	case "link":
		go osucommands.Link(s, m, args, osuAPI, playerCache)
	case "recent", "r", "rs":
		go osucommands.Recent(s, m, args, osuAPI, playerCache, "recent", mapCache)
	case "recentb", "rb", "recentbest":
		go osucommands.Recent(s, m, args, osuAPI, playerCache, "best", mapCache)
	}
}
