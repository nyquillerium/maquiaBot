package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	structs "../structs"
	tools "../tools"
	admincommands "./admin-commands"
	botcreatorcommands "./bot-creator-commands"
	gencommands "./general-commands"
	osucommands "./osu-commands"
	pokemoncommands "./pokemon-commands"
	"github.com/bwmarrin/discordgo"
)

// MessageHandler handles any incoming messages
func MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	m.Content = strings.ToLower(m.Content)
	if strings.Contains(m.Content, "@everyone") {
		m.Content = strings.TrimSpace(strings.ReplaceAll(m.Content, "@everyone", ""))
	}
	if strings.Contains(m.Content, "@here") {
		m.Content = strings.TrimSpace(strings.ReplaceAll(m.Content, "@here", ""))
	}

	emojiRegex, _ := regexp.Compile(`<(:.+:)\d+>`)
	noEmoji := m.Content
	if emojiRegex.MatchString(m.Content) {
		noEmoji = emojiRegex.ReplaceAllString(m.Content, emojiRegex.FindStringSubmatch(m.Content)[1])
	}

	// Obtain profile cache data
	var profileCache []structs.PlayerData
	f, err := ioutil.ReadFile("./data/osuData/profileCache.json")
	tools.ErrRead(err)
	_ = json.Unmarshal(f, &profileCache)

	// Obtain mapper data
	var mapperData []structs.MapperData
	f, err = ioutil.ReadFile("./data/osuData/mapperData.json")
	tools.ErrRead(err)
	_ = json.Unmarshal(f, &mapperData)

	// Obtain server data
	server, err := s.Guild(m.GuildID)
	if err != nil {
		server = &discordgo.Guild{}
	}
	serverData := tools.GetServer(*server)
	serverPrefix := serverData.Prefix

	// CRAB RAVE
	if serverData.Crab && (strings.Contains(m.Content, "crab") || strings.Contains(m.Content, "rave")) && !strings.HasPrefix(m.Content, serverPrefix+"crab") {
		go gencommands.Crab(s, m)
		go tools.CommandLog(s, m, "crab rave")
	}

	// LATE
	if serverData.Late && (strings.Contains(m.Content, "late") || (strings.Contains(m.Content, "old") && !strings.Contains(m.Content, "old.ppy.sh")) || strings.Contains(m.Content, "ancient")) && !strings.HasPrefix(m.Content, serverPrefix+"late") {
		go gencommands.Late(s, m)
		go tools.CommandLog(s, m, "late")
	}

	// CHEERS
	if serverData.Cheers && (strings.Contains(m.Content, "🍻") || strings.Contains(m.Content, "🍺") || strings.Contains(m.Content, "🦐") || strings.Contains(m.Content, "cheer")) && !strings.HasPrefix(m.Content, serverPrefix+"cheers") {
		go gencommands.Cheers(s, m)
		go tools.CommandLog(s, m, "cheers")
	}

	// Generate regexes for message parsing
	profileRegex, _ := regexp.Compile(`(osu|old)\.ppy\.sh\/(u|users)\/(\S+)`)
	beatmapRegex, _ := regexp.Compile(`(osu|old)\.ppy\.sh\/(s|b|beatmaps|beatmapsets)\/(\d+)(#(osu|taiko|fruits|mania)\/(\d+))?`)
	linkRegex, _ := regexp.Compile(`https?:\/\/\S*`)
	timestampRegex, _ := regexp.Compile(`(\d+):(\d{2}):(\d{3})\s*(\(((\d\,?)+)\))?`)
	ideaRegex, _ := regexp.Compile(`(n+i+c+e*|g+o+d+|g+u+d+|c+o+l+)\s*i+d+e+a+`)
	overRegex, _ := regexp.Compile(`ove*r\s+it+`)

	// NICE IDEA
	if ideaRegex.MatchString(m.Content) && serverData.NiceIdea && (!strings.HasPrefix(m.Content, serverPrefix+"idea") || !strings.HasPrefix(m.Content, serverPrefix+"niceidea")) {
		go s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=aAxjVu3iZps")
		go tools.CommandLog(s, m, "nice idea")
	}

	// OVER IT
	if overRegex.MatchString(m.Content) && serverData.OverIt && !strings.HasPrefix(m.Content, serverPrefix+"over") {
		go gencommands.OverIt(s, m)
		go tools.CommandLog(s, m, "over it")
	}

	// Timestamp conversions
	if timestampRegex.MatchString(noEmoji) && serverData.OsuToggle {
		go osucommands.TimestampMessage(s, m, timestampRegex)
	}

	// Vibe check (1/100000 chance if vibe is on in the server)
	if serverData.Vibe {
		roll, _ := rand.Int(rand.Reader, big.NewInt(100000))
		number := roll.Int64()
		if number == 0 {
			go gencommands.Vibe(s, m, "")
		}
	}

	// Command checks
	if strings.HasPrefix(m.Content, "maquiaprefix") {
		go admincommands.Prefix(s, m)
		go tools.CommandLog(s, m, "maquiaprefix")
		return
	} else if strings.HasPrefix(m.Content, "maquiacleanf") || strings.Contains(m.Content, "maquiacleanfarm") {
		go botcreatorcommands.CleanFarm(s, m, profileCache)
		return
	} else if strings.HasPrefix(m.Content, "maquiaclean") {
		go botcreatorcommands.Clean(s, m, profileCache)
		return
	} else if strings.HasPrefix(m.Content, serverPrefix) {
		args := strings.Split(m.Content, " ")
		switch args[0] {
		// Commands without functions
		case serverPrefix + "complain":
			go s.ChannelMessageSend(m.ChannelID, "Shut up hoe")
		case serverPrefix + "dubs", serverPrefix + "doubles", serverPrefix + "trips", serverPrefix + "triples", serverPrefix + "quads", serverPrefix + "quadruples", serverPrefix + "quints", serverPrefix + "quintuples":
			go s.ChannelMessageSend(m.ChannelID, "Ur retarded")
		case serverPrefix + "k", serverPrefix + "key":
			go s.ChannelMessageSend(m.ChannelID, "``` Default AES encryption key: Nxb]^NSc;L*qn3K(/tN{6N7%4n32fF#@```\n This key is given out publicly and I use it for all of my encryption tools, so please do not use me for sensitive data.\n To use your own key, make sure you add a `-k` flag!")
		case serverPrefix + "noncesize", serverPrefix + "nsize":
			key := []byte("Nxb]^NSc;L*qn3K(/tN{6N7%4n32fF#@")
			block, _ := aes.NewCipher(key)
			gcm, _ := cipher.NewGCM(block)
			go s.ChannelMessageSend(m.ChannelID, "The nonce size using the default AES encryption key is "+strconv.Itoa(gcm.NonceSize()))
		case serverPrefix + "src", serverPrefix + "source":
			go s.ChannelMessageSend(m.ChannelID, "https://github.com/VINXIS/maquiaBot")

		// Bot owner commands
		case serverPrefix + "announce":
			go botcreatorcommands.Announce(s, m)
		case serverPrefix + "clean":
			go botcreatorcommands.Clean(s, m, profileCache)
		case serverPrefix + "cleane", serverPrefix + "cleanempty":
			go botcreatorcommands.CleanEmpty(s, m, profileCache)
		case serverPrefix + "cleanf", serverPrefix + "cleanfarm":
			go botcreatorcommands.CleanFarm(s, m, profileCache)
		case serverPrefix + "up", serverPrefix + "update":
			go botcreatorcommands.Update(s, m)
		case serverPrefix + "updatestatus":
			go botcreatorcommands.UpdateStatus(s, m)

		// Sub-handles for pokemon and osu!
		case serverPrefix + "h", serverPrefix + "help":
			go HelpHandle(s, m, serverPrefix)
		case serverPrefix + "o", serverPrefix + "osu":
			go OsuHandle(s, m, args, profileCache, mapperData)
		case serverPrefix + "pokemon":
			go PokemonHandle(s, m, args, serverPrefix)

		// Admin commands'
		case serverPrefix + "prefix", serverPrefix + "newprefix":
			go admincommands.Prefix(s, m)
		case serverPrefix + "purge":
			go admincommands.Purge(s, m)
		case serverPrefix + "toggle":
			go admincommands.Toggle(s, m)
		case serverPrefix + "tr", serverPrefix + "track":
			go admincommands.Track(s, m)
		case serverPrefix + "tt", serverPrefix + "trackt", serverPrefix + "ttoggle", serverPrefix + "tracktoggle":
			go admincommands.TrackToggle(s, m)

		// General commands
		case serverPrefix + "adj", serverPrefix + "adjective", serverPrefix + "adjectives":
			go gencommands.Adjectives(s, m)
		case serverPrefix + "avatar", serverPrefix + "ava", serverPrefix + "a":
			go gencommands.Avatar(s, m)
		case serverPrefix + "cc", serverPrefix + "cp", serverPrefix + "comparec", serverPrefix + "comparep", serverPrefix + "comparecock", serverPrefix + "comparepenis":
			if serverData.Daily {
				go gencommands.PenisCompare(s, m)
			}
		case serverPrefix + "cv", serverPrefix + "comparev", serverPrefix + "comparevagina":
			if serverData.Daily {
				go gencommands.VaginaCompare(s, m)
			}
		case serverPrefix + "ch", serverPrefix + "choose":
			go gencommands.Choose(s, m)
		case serverPrefix + "cheers":
			go gencommands.Cheers(s, m)
		case serverPrefix + "crab":
			go gencommands.Crab(s, m)
		case serverPrefix + "decrypt":
			go gencommands.Decrypt(s, m)
		case serverPrefix + "e", serverPrefix + "emoji", serverPrefix + "emote":
			go gencommands.Emoji(s, m)
		case serverPrefix + "encrypt":
			go gencommands.Encrypt(s, m)
		case serverPrefix + "face":
			go gencommands.Face(s, m)
		case serverPrefix + "history":
			if serverData.Daily {
				go gencommands.History(s, m)
			}
		case serverPrefix + "idea", serverPrefix + "niceidea":
			go s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=aAxjVu3iZps")
		case serverPrefix + "info":
			go gencommands.Info(s, m, profileCache)
		case serverPrefix + "kanye":
			go gencommands.Kanye(s, m)
		case serverPrefix + "leven", serverPrefix + "levenshtein":
			go gencommands.Levenshtein(s, m)
		case serverPrefix + "late", serverPrefix + "old", serverPrefix + "ancient":
			go gencommands.Late(s, m)
		case serverPrefix + "meme":
			go gencommands.Meme(s, m)
		case serverPrefix + "noun", serverPrefix + "nouns":
			go gencommands.Nouns(s, m)
		case serverPrefix + "ocr":
			go gencommands.OCR(s, m)
		case serverPrefix + "over":
			go gencommands.OverIt(s, m)
		case serverPrefix + "p", serverPrefix + "per", serverPrefix + "percent", serverPrefix + "percentage":
			go gencommands.Percentage(s, m)
		case serverPrefix + "parse":
			go gencommands.Parse(s, m)
		case serverPrefix + "penis", serverPrefix + "cock":
			if serverData.Daily {
				go gencommands.Penis(s, m)
			}
		case serverPrefix + "ping":
			go gencommands.Ping(s, m)
		case serverPrefix + "q", serverPrefix + "quote":
			go gencommands.Quote(s, m)
		case serverPrefix + "qa", serverPrefix + "qadd", serverPrefix + "quotea", serverPrefix + "quoteadd":
			go gencommands.QuoteAdd(s, m)
		case serverPrefix + "qd", serverPrefix + "qr", serverPrefix + "qdelete", serverPrefix + "qremove", serverPrefix + "quotedelete", serverPrefix + "quoteremove":
			go gencommands.QuoteRemove(s, m)
		case serverPrefix + "qs", serverPrefix + "quotes":
			go gencommands.Quotes(s, m)
		case serverPrefix + "rc", serverPrefix + "rp", serverPrefix + "rankc", serverPrefix + "rankp", serverPrefix + "rankcock", serverPrefix + "rankpenis":
			if serverData.Daily {
				go gencommands.PenisRank(s, m)
			}
		case serverPrefix + "remind", serverPrefix + "reminder":
			go gencommands.Remind(s, m)
		case serverPrefix + "reminders":
			go gencommands.Reminders(s, m)
		case serverPrefix + "remindremove", serverPrefix + "rremove":
			go gencommands.RemoveReminder(s, m)
		case serverPrefix + "rinfo", serverPrefix + "roleinfo":
			go gencommands.RoleInfo(s, m)
		case serverPrefix + "roll":
			go gencommands.Roll(s, m)
		case serverPrefix + "rv", serverPrefix + "rankv", serverPrefix + "rankvagina":
			if serverData.Daily {
				go gencommands.VaginaRank(s, m)
			}
		case serverPrefix + "sinfo", serverPrefix + "serverinfo":
			go gencommands.ServerInfo(s, m)
		case serverPrefix + "skill", serverPrefix + "skills":
			go gencommands.Skills(s, m)
		case serverPrefix + "stats", serverPrefix + "class":
			go gencommands.Stats(s, m)
		case serverPrefix + "twitch", serverPrefix + "twitchdl":
			go gencommands.Twitch(s, m)
		case serverPrefix + "twitter", serverPrefix + "twitterdl":
			go gencommands.Twitter(s, m)
		case serverPrefix + "vagina":
			if serverData.Daily {
				go gencommands.Vagina(s, m)
			}
		case serverPrefix + "vibe", serverPrefix + "vibec", serverPrefix + "vibecheck":
			go gencommands.Vibe(s, m, "notRandom")

		// osu! commands
		case serverPrefix + "bfarm", serverPrefix + "bottomfarm":
			go osucommands.BottomFarm(s, m, profileCache)
		case serverPrefix + "bpm":
			if serverData.Daily {
				go osucommands.BPM(s, m, profileCache)
			}
		case serverPrefix + "c", serverPrefix + "compare":
			go osucommands.Compare(s, m, profileCache)
		case serverPrefix + "farm":
			go osucommands.Farm(s, m, profileCache)
		case serverPrefix + "l", serverPrefix + "leader", serverPrefix + "leaderboard":
			go osucommands.Leaderboard(s, m, beatmapRegex, profileCache)
		case serverPrefix + "link", serverPrefix + "set":
			go osucommands.Link(s, m, args, profileCache)
		case serverPrefix + "m", serverPrefix + "map":
			go osucommands.BeatmapMessage(s, m, beatmapRegex)
		case serverPrefix + "mt", serverPrefix + "mtrack", serverPrefix + "maptrack", serverPrefix + "mappertrack":
			go osucommands.TrackMapper(s, m, mapperData)
		case serverPrefix + "mti", serverPrefix + "mtinfo", serverPrefix + "mtrackinfo", serverPrefix + "maptracking", serverPrefix + "mappertracking", serverPrefix + "mappertrackinfo":
			go osucommands.TrackMapperInfo(s, m, mapperData)
		case serverPrefix + "osutop", serverPrefix + "osudetail":
			go osucommands.ProfileMessage(s, m, profileRegex, profileCache)
		case serverPrefix + "ppadd":
			go osucommands.PPAdd(s, m, profileCache)
		case serverPrefix + "profile":
			go osucommands.ProfileMessage(s, m, profileRegex, profileCache)
		case serverPrefix + "r", serverPrefix + "rs", serverPrefix + "recent":
			go osucommands.Recent(s, m, "recent", profileCache)
		case serverPrefix + "rb", serverPrefix + "recentb", serverPrefix + "recentbest":
			go osucommands.Recent(s, m, "best", profileCache)
		case serverPrefix + "s", serverPrefix + "sc", serverPrefix + "scorepost":
			go osucommands.ScorePost(s, m, profileCache, "scorePost")
		case serverPrefix + "t", serverPrefix + "top":
			go osucommands.Top(s, m, profileCache)
		case serverPrefix + "tfarm", serverPrefix + "topfarm":
			go osucommands.TopFarm(s, m, profileCache)
		case serverPrefix + "ti", serverPrefix + "tinfo", serverPrefix + "tracking", serverPrefix + "trackinfo":
			go osucommands.TrackInfo(s, m)

		// Pokemon commands
		case serverPrefix + "b", serverPrefix + "berry":
			go pokemoncommands.Berry(s, m)
		}
		go tools.CommandLog(s, m, args[0])
		return
	} else if beatmapRegex.MatchString(m.Content) && serverData.OsuToggle { // If a beatmap was linked
		go osucommands.BeatmapMessage(s, m, beatmapRegex)
		return
	} else if profileRegex.MatchString(m.Content) && serverData.OsuToggle { // If a profile was linked
		go osucommands.ProfileMessage(s, m, profileRegex, profileCache)
		return
	}

	// Dont mention me mate. Ill fuck u up
	if len(m.Mentions) > 0 {
		for _, mention := range m.Mentions {
			if mention.ID == s.State.User.ID {
				roll, _ := rand.Int(rand.Reader, big.NewInt(100))
				number := roll.Int64()
				if number%11 == 0 {
					s.ChannelMessageSend(m.ChannelID, "lol dude i bet u havent even watched the Maquia movie stop pinging me .")
					break
				}
				s.ChannelMessageSend(m.ChannelID, "what do u want dude lol")
				break
			}
		}
	}

	// Check if an image was linked
	if len(m.Attachments) > 0 || linkRegex.MatchString(m.Content) || (len(m.Embeds) > 0 && m.Embeds[0].Image != nil) {
		go osucommands.OsuImageParse(s, m, linkRegex)
		go osucommands.ReplayMessage(s, m, linkRegex, profileCache)
	}
}
