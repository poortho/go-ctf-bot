package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"strconv"
	"encoding/json"
//	"io"
	"io/ioutil"
	"strings"
	"github.com/bwmarrin/discordgo"
)

var categories = [7]string{"pwn", "rev", "for", "misc", "web", "ppc", "crypto"}
var config map[string]int

func main() {
	// read token from file
	data, err := ioutil.ReadFile("./token")
	if err != nil {
		fmt.Println("Error reading token.")
		fmt.Println("Error message: ", err)
		os.Exit(1)
	}
	token := strings.TrimSpace(string(data))
	fmt.Println("Token: " + token)

	// read config from file
	data, err = ioutil.ReadFile("./category_ids.json")
	if err != nil {
		fmt.Println("Error reading category ids.")
		fmt.Println("Error message: ", err)
		os.Exit(1)
	}
	err = json.Unmarshal(data, &config)

	// create discord session
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		os.Exit(2)	
	}

	dg.AddHandler(ready)
	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		os.Exit(3)
	}

	
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	s.UpdateStatus(0, "!help")
}

func messageCreate(session *discordgo.Session, msg *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if msg.Author.ID == session.State.User.ID {
		return
	}

	switch {
	case strings.HasPrefix(msg.Content, "!help"):

		// display help
		_, err := session.ChannelMessageSend(msg.ChannelID, "Commands:\n`!help` - display this message\n`!add [category (pwn, rev, misc, for, crypto, ppc, web)] [point value] [chall name (no spaces)]` - create a new channel for a challenge\n`!solve` - mark a challenge as solved and move the channel accordingly")
		if err != nil {
			fmt.Println("Error sending help message: ", err)
			os.Exit(4)
		}
	case strings.HasPrefix(msg.Content, "!add "):
		args := strings.Fields(msg.Content)

		// create new channel
		if len(args) >= 4 && stringInSlice(args[1], categories) && IsNumeric(args[2]) {
			_, err := session.GuildChannelCreateComplex(msg.GuildID, discordgo.GuildChannelCreateData{
				Name: strings.Join(args[1:], "-"),
				Type: discordgo.ChannelTypeGuildText,
				ParentID: strconv.Itoa(config[args[1]]),
			})
			if err != nil {
				fmt.Println("Error creating new channel: ", err)
				os.Exit(5)
			}
		} else {
			_, err := session.ChannelMessageSend(msg.ChannelID, "Invalid args. Usage: `!add [category (pwn, rev, misc, for, crypto, ppc, web)] [point value] [chall name]`")
			if err != nil {
				fmt.Println("Error sending add error message: ", err)
				os.Exit(6)
			}
		}
	case strings.HasPrefix(msg.Content, "!solve"):
		// move channel to "solved" category
		channel, err := session.State.Channel(msg.ChannelID)
		if err != nil {
			fmt.Println("Error getting channel in solve command: ", err)
			os.Exit(7)
		}
		if channel.ParentID == "" || channel.ParentID == strconv.Itoa(config["solved"]) {
			_, err := session.ChannelMessageSend(msg.ChannelID, "This isn't a challenge channel, or the challenge is already solved!")
			if err != nil {
				fmt.Println("Error sending solve error message: ", err)
				os.Exit(8)
			}
		} else {
			_, err := session.ChannelEditComplex(msg.ChannelID, &discordgo.ChannelEdit{
				ParentID: strconv.Itoa(config["solved"]),
			})
			if err != nil {
				fmt.Println("Error moving solved channel: ", err)
				os.Exit(9)
			}
		}

	case strings.HasPrefix(msg.Content, "!"):
		_, err := session.ChannelMessageSend(msg.ChannelID, "Unrecognized command `" + msg.Content + "`. Enter `!help` for a list of valid commands.")
		if err != nil {
			fmt.Println("Error sending unrecognized command message: ", err)
			os.Exit(10)
		}
	}
}

func stringInSlice(a string, list [7]string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func IsNumeric(s string) bool {
    _, err := strconv.ParseInt(s, 10, 64)
    return err == nil
}