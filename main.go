package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"
)

type Request struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Content string   `json:"content"`
	source  net.Addr
}

func handleConn(c net.Conn, out chan<- Request) { //,out chan<- string
	defer c.Close()
	for {
		var message Request
		err := json.NewDecoder(c).Decode(&message)
		if err == io.EOF {
			return
		} else if err == nil {
			message.source = c.RemoteAddr()
			out <- message
		}
	}
}

func runQueue(in <-chan Request, out chan<- Request) {
	var inter interface{}
	var msg Request
	var toQueue Request
	q := NewQueue()

	for {
		msg = <-in

		for ok := true; ok; {
			select {
			case out <- msg:
				inter, ok = q.deQueue()
				if ok {
					msg = inter.(Request)
				}
			case toQueue = <-in:
				q.queue(toQueue)
			}
		}
	}
}

func talkToMe(in <-chan Request) {
	var lastSource string
	var source string
	for {
		msg := <-in
		if msg.source != nil {
			source = msg.source.String()[:strings.LastIndex(msg.source.String(), ":")]
		} else {
			source = "Unknown"
		}
		fmt.Printf("%s\t%s\n", source, msg.Content)
		var cmd *exec.Cmd
		switch msg.Command {
		case "say":
			cmd = exec.Command("say", msg.Args...)
			cmd.Stdin = strings.NewReader(msg.Content)
		case "who":
			cmd = exec.Command("say")
			cmd.Stdin = strings.NewReader(fmt.Sprintf("Dernier contact avec %s", lastSource))
		case "volume":
			cmd = exec.Command("osascript", "-e", fmt.Sprintf("set volume output volume %s", msg.Args[0]))
		default:
			cmd = exec.Command("say")
			cmd.Stdin = strings.NewReader("Erreur command non reconnue !")
		}
		cmd.Run()
		lastSource = source
	}
}

func main() {
	var channel, token string
	flag.StringVar(&token, "token", "", "usage: --token XXXX-XXXX")
	flag.StringVar(&channel, "chan", "", "usage: --channel general")
	flag.Parse()
	fmt.Printf("token: %s\nchannel: %s\n", token, channel)
	listener, err := net.Listen("tcp", "0.0.0.0:8000")
	toQueue := make(chan Request)
	toSay := make(chan Request)
	if err != nil {
		log.Fatal(err)
	}
	go runQueue(toQueue, toSay)
	go talkToMe(toSay)
	if channel != "" && token != "" {
		go SlackListenner(token, channel, toQueue)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn, toQueue)
	}
}
