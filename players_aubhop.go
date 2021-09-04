package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
	"github.com/rumblefrog/go-a2s"
)

//   main 864832817749819452
//  wl 864834961508007946
//  trikz 864834981649317899

// bot-log channel 516854208113147914

type server struct {
	name string
	addr string
}

type serverWithChannel struct {
	name string
	addr string
	id   string
}

var lastSentID = ""

var servers = []server{
	{"🍺 Pub: ", "144.48.37.114:27015"},
	{"🤍 WL: ", "144.48.37.118:27015"},
	{"🛹 Trikz: ", "144.48.37.119:27015"},
	{"🦘 Kanga: ", "146.185.214.33:27015"},
	{"🌌 Solitude: ", "51.161.131.99:27015"},
}

var serversWithChannels = []serverWithChannel{
	{"🍺 Pub: ", "144.48.37.114:27015", "864832817749819452"},
	{"🤍 WL: ", "144.48.37.118:27015", "864834961508007946"},
	{"🛹 Trikz: ", "144.48.37.119:27015", "864834981649317899"},
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

func msgReceived(s *discordgo.Session, msg *discordgo.MessageCreate) {
	if strings.Index(msg.Content, ".on") == 0 {
		if err := s.ChannelMessageDelete(msg.ChannelID, msg.ID); err != nil {
			println(err)
		}
		if lastSentID != "" {
			if err := s.ChannelMessageDelete(msg.ChannelID, lastSentID); err != nil {
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
							strings.Index(player.Name, "Bonus") == -1 {
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
			lastSentID = msg.ID
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

func main() {
	discord, err := discordgo.New("Bot " + "")

	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	discord.AddHandler(ready)

	discord.AddHandler(msgReceived)

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
