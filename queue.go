package main

type Queue struct {
	head *Link
	tail *Link
}

type Link struct {
	value interface{}
	next  *Link
}

func NewQueue() *Queue {
	q := new(Queue)
	q.head = nil
	q.tail = nil
	return q
}

func (q *Queue) queue(str interface{}) {
	if q.head == nil {
		q.head = new(Link)
		q.tail = q.head
		q.head.value = str
		q.head.next = nil
		return
	}
	link := new(Link)
	link.value = str
	link.next = nil
	q.head.next = link
	q.head = link
}

func (q *Queue) deQueue() (interface{}, bool) {
	var str interface{}

	if q.head == nil {
		return str, false
	}

	str = q.tail.value

	if q.head == q.tail {
		q.head = nil
		q.tail = nil
	} else {
		q.tail = q.tail.next
	}
	return str, true
}
