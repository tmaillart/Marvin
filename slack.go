package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/websocket"
)

type Data struct {
	Url      string    `json:"url"`
	Channels []Channel `json:"channels"`
	Ok       bool      `json:"ok"`
}

type Channel struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type Message struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

func curl(url string, i interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, i)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func SlackListenner(token string, chaname string, out chan<- Request) { // args channel, token
	var response Data
	c := Channel{Name: chaname, Id: ""}
	//token := "00000-00000"
	if curl("https://slack.com/api/rtm.start?token="+token, &response) != nil || !response.Ok {
		fmt.Printf("Slack connection error: bad token\n")
		return
	}

	for i := 0; i < len(response.Channels) && c.Id == ""; i++ {
		if response.Channels[i].Name == c.Name {
			c.Id = response.Channels[i].Id
		}
	}
	if c.Id == "" {
		fmt.Printf("Slack connection error: Unrecognized channel\n")
		return
	}

	fmt.Printf("Websocket: %s\nChannel %s = %s\n", response.Url, c.Name, c.Id)

	var dialer *websocket.Dialer
	conn, _, err := dialer.Dial(response.Url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	var msg Message
	var req Request
	req.Command = "say"
	req.source = nil
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			return
		}
		if json.Unmarshal(message, &msg) == nil && msg.Type == "message" && msg.Channel == c.Id {
			req.Content = msg.Text
			out <- req
			fmt.Printf("Slack\t%s\n", msg.Text)
		}
	}

}

