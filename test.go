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

func main() { // args channel, token
	var response Data
	c := Channel{Name: "blabla", Id: ""}
	token := "00000-00000"
	if curl("https://slack.com/api/rtm.start?token="+token, &response) != nil || !response.Ok {
		return
	}

	for i := 0; i < len(response.Channels) && c.Id == ""; i++ {
		if response.Channels[i].Name == c.Name {
			c.Id = response.Channels[i].Id
		}
	}

	fmt.Printf("Websocket: %s\nChannel %s = %s\n", response.Url, c.Name, c.Id)

	var dialer *websocket.Dialer
	conn, _, err := dialer.Dial(response.Url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	var msg Message
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			return
		}
		if json.Unmarshal(message, &msg) == nil && msg.Type == "message" {
			fmt.Printf("%s\n", msg.Text)
		}
	}

}
