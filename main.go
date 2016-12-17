package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
)

type Queue struct {
	head *Link
}

type Link struct {
	value    string
	next     *Link
	previous *Link
}

func NewQueue() *Queue {
	q := new(Queue)
	q.head = nil
	return q
}

func (q *Queue) queue(str string) {
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

func (q *Queue) deQueue() (string, bool) {
	if q.head == nil {
		return "", false
	}

	if q.head.next == q.head || q.head.previous == q.head {
		str := q.head.value
		q.head = nil
		return str, true
	}
	str := q.head.value
	//toRemove:=q.previous
	//q.previous.previous.next=q
	//q.previous=q.previous.previous
	//delete(toRemove)
	q.head.previous.next = q.head.next
	q.head.next.previous = q.head.previous
	q.head = q.head.next
	return str, true
}

func handleConn(c net.Conn, out chan<- string) { //,out chan<- string
	defer c.Close()
	reader := bufio.NewReader(c)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		out <- message
		//fmt.Printf("loop\n")
	}
}

func runQueue(in <-chan string, out chan<- string) {
	var msg string
	var toQueue string
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

func talkToMe(in <-chan string) {
	for {
		msg := <-in
		fmt.Printf("%s", msg)

		// Extract parts from input (format: message | options)
		parts := strings.Split(msg, "|")

		// Message
		message := parts[0]

		// Optional arguments
		arguments := strings.Fields(strings.Join(parts[1:], ""))

		// Build command arguments
		args := append([]string{message}, arguments...)

		// Execute command
		cmd := exec.Command("say", args...)
		cmd.Run()
	}
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:8000")
	toQueue := make(chan string)
	toSay := make(chan string)
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
