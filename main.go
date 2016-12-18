package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
)

type Request struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Content string   `json:"content"`
}

type Queue struct {
	head *Link
}

type Link struct {
	value    Request
	next     *Link
	previous *Link
}

func NewQueue() *Queue {
	q := new(Queue)
	q.head = nil
	return q
}

func (q *Queue) queue(str Request) {
	if q.head == nil {
		q.head = new(Link)
		q.head.value = str
		q.head.next = q.head
		q.head.previous = q.head
		return
	}
	link := new(Link)
	link.value = str

	link.next = q.head
	link.previous = q.head.previous
	link.previous.next = link
	q.head.previous = link
}

func (q *Queue) deQueue() (Request, bool) {
	var str Request

	if q.head == nil {
		return str, false
	}

	if q.head.next == q.head || q.head.previous == q.head {
		str = q.head.value
		q.head = nil
		return str, true
	}
	str = q.head.value
	//toRemove:=q.previous
	//q.previous.previous.next=q
	//q.previous=q.previous.previous
	//delete(toRemove)
	q.head.previous.next = q.head.next
	q.head.next.previous = q.head.previous
	q.head = q.head.next
	return str, true
}

func handleConn(c net.Conn, out chan<- Request) { //,out chan<- string
	defer c.Close()
	var message Request
	for {
		err := json.NewDecoder(c).Decode(&message)
		if err != nil {
			return
		}
		fmt.Printf("%s\n", message.Command)
		out <- message
	}
}

func runQueue(in <-chan Request, out chan<- Request) {
	var msg Request
	var toQueue Request
	var ok bool
	q := NewQueue()
	for {
		q.queue(<-in)
		msg, ok = q.deQueue()
		for ok {
			select {
			case out <- msg:
				msg, ok = q.deQueue()
			case toQueue = <-in:
				q.queue(toQueue)
			}
		}
	}
}

func talkToMe(in <-chan Request) {
	for {
		msg := <-in
		fmt.Printf("%s\n", msg.Content)
		var cmd *exec.Cmd
		switch msg.Command {
		case "say":
			cmd = exec.Command("say", msg.Args...)
			cmd.Stdin = strings.NewReader(msg.Content)
		default:
			cmd = exec.Command("say")
			cmd.Stdin = strings.NewReader("Erreur command non reconnue !")
		}
		cmd.Run()
	}
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:8000")
	toQueue := make(chan Request)
	toSay := make(chan Request)
	if err != nil {
		log.Fatal(err)
	}
	go runQueue(toQueue, toSay)
	go talkToMe(toSay)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn, toQueue)
	}
}
