package main

import(
  "log"
  "net"
  "bufio"
  "fmt"
  "os/exec"
  "strings"
)

type Queue struct{
  value string
  next *Queue
  previous *Queue
}

func (q *Queue) init() {//int
  //q.value=""
  q.next=nil
  q.previous=nil
}

func (q *Queue) isEmpty() bool{//int
  return q.next==nil || q.previous==nil
}

func (q *Queue) queue(str string) {//int
  if q.isEmpty() {
    q.value=str
    q.next=q
    q.previous=q
    return
  }
  link:=new(Queue)
  link.value=str

  link.next=q
  link.previous=q.previous
  link.previous.next=link
  q.previous=link
  //return 0
}

func (q *Queue) deQueue() (string,*Queue,bool){
  if q.isEmpty() {
    return "",q,false
  }
  if q.next==q || q.previous==q {
    str:=q.value
    q.init()
    return str,q,true
  }
  str:=q.value
  //toRemove:=q.previous
  //q.previous.previous.next=q
  //q.previous=q.previous.previous
  //delete(toRemove)
  q.previous.next=q.next
  q.next.previous=q.previous
  q=q.next
  return str,q,true
}

func handleConn(c net.Conn,out chan<- string){//,out chan<- string
  defer c.Close()
  reader := bufio.NewReader(c)
  for{
    message, err := reader.ReadString('\n')
    if err!=nil {
      return
    }
    out <- message
    //fmt.Printf("loop\n")
  }
}

func runQueue(in <-chan string,out chan<- string){
  var msg string
  var toQueue string
  var ok bool
  var q=new(Queue)
  q.init()
  for{
    q.queue(<-in)
    msg,q,ok=q.deQueue()
    for ok {
      select{
      case out <- msg:
        msg,q,ok=q.deQueue()
      case toQueue = <-in:
        q.queue(toQueue)
      }
    }
  }
}

func talkToMe(in <-chan string){
  for{
    msg:=<-in
    fmt.Printf("%s",msg)

    // Extract parts from input (format: message | options)
    parts := strings.Split(msg, "|")

    // Message
    message := parts[0]

    // Optional arguments
    arguments := strings.Fields(strings.Join(parts[1:], ""))

    // Build command arguments
    args := append([]string{message}, arguments...)

    // Execute command
    cmd:=exec.Command("say", args...)
    cmd.Run()
  }
}

func main(){
  listener,err:=net.Listen("tcp","0.0.0.0:8000")
  toQueue := make(chan string)
  toSay := make(chan string)
  if err!=nil{
    log.Fatal(err)
  }
  go runQueue(toQueue,toSay)
  go talkToMe(toSay)
  for{
    conn, err:=listener.Accept()
    if err != nil{
      log.Print(err)
      continue
    }
    go handleConn(conn,toQueue)
  }
}
