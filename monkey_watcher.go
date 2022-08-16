package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gosnmp/gosnmp"
)

var (
	botUserToken = flag.String("botusertoken", "", "Slack Bot User Token, xoxb-...")
	channel      = flag.String("channelid", "C03Q9LEEKDF", "Channel ID, C...")
)

type SlackApiChatPostMessageResponse struct {
	Ok      bool   `json:"ok,omitempty"`
	Error   string `json:"error,omitempty"`
	Channel string `json:"channel,omitempty"`
	Ts      string `json:"ts,omitempty"`
	Message struct {
		Text        string `json:"text,omitempty"`
		Username    string `json:"username,omitempty"`
		BotID       string `json:"bot_id,omitempty"`
		Attachments []struct {
			Text     string `json:"text,omitempty"`
			ID       int    `json:"id,omitempty"`
			Fallback string `json:"fallback,omitempty"`
		} `json:"attachments,omitempty"`
		Type    string `json:"type,omitempty"`
		Subtype string `json:"subtype,omitempty"`
		Ts      string `json:"ts,omitempty"`
	} `json:"message,omitempty"`
}

func SlackApiChatPostMessage(message string, channel string, color string) {
	client := &http.Client{}
	formdata := url.Values{}
	formdata.Set("channel", channel)
	formdata.Set("attachments", `[{"text": "`+message+`", "color": "`+color+`"}]`)

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", strings.NewReader(formdata.Encode()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+*botUserToken)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respDecoded := json.NewDecoder(resp.Body)

	var slackApiChatPostMessageResponse SlackApiChatPostMessageResponse

	err = respDecoded.Decode(&slackApiChatPostMessageResponse)
	if err != nil {
		panic(err)
	}
	if slackApiChatPostMessageResponse.Ok != true {
		//log.Println(resp.Body)
		log.Println(slackApiChatPostMessageResponse)
	}
	return

}

func main() {
	flag.Parse()
	tl := gosnmp.NewTrapListener()
	tl.OnNewTrap = myTrapHandler
	tl.Params = gosnmp.Default
	//	tl.Params.Logger = gosnmp.NewLogger(log.New(os.Stdout, "", 0))

	err := tl.Listen("0.0.0.0:9162")
	if err != nil {
		log.Panicf("error in listen: %s", err)
	}
}

func myTrapHandler(packet *gosnmp.SnmpPacket, addr *net.UDPAddr) {
	var status string
	var ap string
	var color string
	for _, v := range packet.Variables {
		switch v.Name {
		case ".1.3.6.1.6.3.1.1.4.1.0":
			switch v.Value {
			case ".1.3.6.1.4.1.9.9.513.0.4":
				status = "✨ AP UP ✨\n"
				color = "good"
			case ".1.3.6.1.4.1.14179.2.6.3.8":
				status = "💀 AP DOWN 💀\n"
				color = "danger"
			default:
			}

		case ".1.3.6.1.4.1.14179.2.2.1.1.3.0":
			// offline
			ap = string(v.Value.([]uint8))

		default:
			if strings.HasPrefix(v.Name, ".1.3.6.1.4.1.9.9.513.1.1.1.1.5.") {
				// online
				ap = string(v.Value.([]uint8))
			}
			log.Printf("trap: %+v\n", v)
		}
	}
	if status != "" {
		SlackApiChatPostMessage(status+ap, *channel, color)
	}
}
