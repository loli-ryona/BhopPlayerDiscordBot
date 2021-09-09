package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cavaliercoder/grab"
	"github.com/davecgh/go-spew/spew"
	"github.com/rumblefrog/go-a2s"
)

//original players_aubhop.go was made by tanami
//i just forked it a built off it

/* TO DO
- Make bot messages embeded instead of normal chat
- Somehow get gamebanana api working on .dl */

type server struct {
	name string
	addr string
}

type serverWithChannel struct {
	name string
	addr string
	id   string
}

type discordAuthID struct {
	AuthID string `json:"AuthID"`
}

var lastSentPlayersCommand = ""
var lastSentHelpCommand = ""

var servers = []server{
	{"🍺 Pub: ", "144.48.37.114:27015"},
	{"🤍 WL: ", "144.48.37.118:27015"},
	{"🛹 Trikz: ", "144.48.37.119:27015"},
	{"🦘 Kanga: ", "146.185.214.33:27015"},
	{"🌌 Solitude: ", "51.161.131.99:27015"},
	{"🚸 IMK Easy: ", "139.99.209.158:27016"},
	{"🚸 IMK Hard: ", "139.99.209.158:27017"},
	{"🏳️‍🌈 Gay Tradies: ", "203.28.238.134:27015"},
	{"☭ Luchshe Veteranov: ", "46.174.52.164:27015"},
}

//main 883671130286739476
//wl 883671190521147412
//kanga 883671212784488498

var serversWithChannels = []serverWithChannel{
	{"🍺 Pub: ", "144.48.37.114:27015", "883671130286739476"},
	{"🤍 WL: ", "144.48.37.118:27015", "883671190521147412"},
	{"🦘 Kanga: ", "146.185.214.33:27015", "883671212784488498"},
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	go func() {
		ticker := time.NewTicker(time.Second * 60)
		for ; true; <-ticker.C {
			for _, serverWithChannel := range serversWithChannels {
				process(s, serverWithChannel.name, serverWithChannel.addr, serverWithChannel.id)
			}
		}
	}()
}

//Merz coded this part so idfk
//Prints players online in channel
func cmdPlayersOnline(s *discordgo.Session, msg *discordgo.MessageCreate) {
	if strings.Index(msg.Content, ".players") == 0 {
		if err := s.ChannelMessageDelete(msg.ChannelID, msg.ID); err != nil {
			println(err)
		}
		if lastSentPlayersCommand != "" {
			if err := s.ChannelMessageDelete(msg.ChannelID, lastSentPlayersCommand); err != nil {
				println(err)
			}
		}
		// output := "*Players online:*\n"
		output := ""
		for _, server := range servers {
			client, err := a2s.NewClient(server.addr)
			var realPlayers []string
			if err != nil {
				// don't care
			} else {
				defer client.Close()
				players, err := client.QueryPlayer()

				if err != nil {
					// don't care
				} else {
					for _, player := range players.Players {
						if strings.Index(player.Name, "!replay") == -1 &&
							strings.Index(player.Name, "WR") == -1 &&
							strings.Index(player.Name, "Main") == -1 &&
							strings.Index(player.Name, "Bonus") == -1 &&
							strings.Index(player.Name, "GOTV") == -1 {
							realPlayers = append(realPlayers, player.Name)
						}
					}

					if len(realPlayers) > 0 {
						output += fmt.Sprintf("%s**%s**\n", server.name, strings.Join(realPlayers, ", "))
					}
				}
			}
		}
		msg, err := s.ChannelMessageSend(msg.ChannelID, output)
		if err != nil {
			println(err)
		} else {
			lastSentPlayersCommand = msg.ID
		}
	}
}

//Prints help to channel
func cmdHelp(s *discordgo.Session, msg *discordgo.MessageCreate) {
	if strings.Index(msg.Content, ".birbhelp") == 0 {
		if err := s.ChannelMessageDelete(msg.ChannelID, msg.ID); err != nil {
			println(err)
		}
		if lastSentHelpCommand != "" {
			if err := s.ChannelMessageDelete(msg.ChannelID, lastSentHelpCommand); err != nil {
				println(err)
			}
		}

		output := "*List of commands:*\n.players = List players on CS:S/CS:GO Bhop Servers\n.dl = Downloads map to the server\n.birbhelp = Prints this message"

		msg, err := s.ChannelMessageSend(msg.ChannelID, output)
		if err != nil {
			println(err)
		} else {
			lastSentHelpCommand = msg.ID
		}
	}
}

//Downloads map to location
//need this for later https://gamebanana.com/mods/
func cmdDownloadMap(s *discordgo.Session, msg *discordgo.MessageCreate) {
	role := msg.GuildID
	println(role)
	if strings.Index(msg.Content, ".dl") == 0 {
		splitInput := strings.Split(msg.Content, " ")
		if strings.Index(splitInput[1], "https://gamebanana.com/mods/") != 0 {
			bhopMap := splitInput[1]
			println(bhopMap)

			//create client
			filepath := "C:/shubek2/maps"
			url := "http://sojourner.me/fastdl/maps/"
			client := grab.NewClient()
			req, _ := grab.NewRequest(filepath, url+bhopMap+".bsp.bz2")

			//start download
			fmt.Printf("Downloading %v...\n", req.URL())
			resp := client.Do(req)
			fmt.Printf("  %v\n", resp.HTTPResponse.Status)

			//start ui loop
			t := time.NewTicker(500 * time.Millisecond)
			defer t.Stop()

		Loop:
			for {
				select {
				case <-t.C:
					fmt.Printf("  Transferred %v bytes - ETA: %v\n", byteString(resp.BytesComplete()), resp.ETA())
				case <-resp.Done:
					//download finished
					break Loop
				}
			}

			//err check
			if err := resp.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
				errorMsg := err.Error()
				msg, err := s.ChannelMessageSend(msg.ChannelID, errorMsg)
				if err != nil {
					println(err)
				} else {
					println(msg.ID)
				}
			}

			//fin
			fmt.Printf("Download saved to ./%v \n", resp.Filename)
			msg, err := s.ChannelMessageSend(msg.ChannelID, "Finished Downloading: "+bhopMap)
			if err != nil {
				println(err)
			} else {
				println(msg.ID)
			}
		} else {
			//now is the aids part which is using the gamebanana api
			msg, err := s.ChannelMessageSend(msg.ChannelID, "Gamebanana api implementation currently wip")
			if err != nil {
				println(err)
			} else {
				println(msg.ID)
			}
		}
	}
}

func process(s *discordgo.Session, name string, addr string, id string) {
	go func() {
		fmt.Println("working on " + name)
		client, err := a2s.NewClient(addr)

		if err != nil {
			// don't care
			fmt.Println("ruh roh")
		} else {
			defer client.Close()
			info, err := client.QueryInfo()

			if err != nil {
				// todo: do something here
				fmt.Printf("%s crapped itself\n", name)

			} else {
				real := float64(info.Players - info.Bots)
				status := fmt.Sprintf("%s%0.f", name, real)
				fmt.Println(status)

				// todo: do not update same value to save rate limits
				// update the discord bot

				ch, err := s.ChannelEdit(id, status)
				if err != nil {
					spew.Dump(ch)
					fmt.Println(err)
				}
			}
		}
	}()
}

func byteString(n int64) string {
	if n < 1<<10 {
		return fmt.Sprintf("%dB", n)
	}
	if n < 1<<20 {
		return fmt.Sprintf("%dKB", n>>10)
	}
	if n < 1<<30 {
		return fmt.Sprintf("%dMB", n>>20)
	}
	if n < 1<<40 {
		return fmt.Sprintf("%dGB", n>>30)
	}
	return fmt.Sprintf("%dTB", n>>40)
}

func main() {
	//opening the authid.json file
	jsonAuthFile, err := os.Open("authid.json")
	if err != nil {
		println(err)
	} else {
		println("Successfullying opened authid.json")
	}
	defer jsonAuthFile.Close()

	//read jsonAuthFile as byte array
	bV, _ := ioutil.ReadAll(jsonAuthFile)
	var dAuthID discordAuthID
	json.Unmarshal(bV, &dAuthID)
	if dAuthID.AuthID != "" {
		dAuthenticationToken := dAuthID.AuthID
		discord, err := discordgo.New("Bot " + dAuthenticationToken)
		if err != nil {
			fmt.Println("Error creating Discord session: ", err)
			return
		}

		discord.AddHandler(ready)
		discord.AddHandler(cmdPlayersOnline)
		discord.AddHandler(cmdHelp)
		discord.AddHandler(cmdDownloadMap)

		err = discord.Open()

		if err != nil {
			fmt.Println("Error opening Discord session: ", err)
			return
		}

		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc

		discord.Close()
	}

	println("Expected auth id but instead found blank.")
	println("Please put your auth id in authid.json")
}
