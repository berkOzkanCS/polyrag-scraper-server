package main

import (
	"sync"
)

type State int

const (
	Idle         State = iota // 0
	Working                   // 1
	Sleep                     // 2
	Disconnected              // 3
)

type Client struct {
	Name              string
	Ip                string
	ComputeHours      int
	ArticlesProcessed int
	State             State
}

func ClientConstructor(name, ip string) *Client {
	return &Client{
		Name:              name,
		Ip:                ip,
		ComputeHours:      0,
		ArticlesProcessed: 0,
		State:             Idle,
	}
}

type ClientMessage struct {
	Name    string `json:"name"`
	State   State  `json:"state"`
	Article struct {
		Title        string `json:"title"`
		Content      string `json:"content"`
		Date         string `json:"date"`
		ComputeHours int    `json:"computeHours"`
	} `json:"article"`
}

type MasterMessage struct {
	Urls       []string `json:"urls"`
	Command    string   `json:"cmd"`
	ClientName string   `json:"clientName"`
}

type Queue struct {
	mu    sync.Mutex
	queue []string
}

func (q *Queue) Enqueue(url string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue = append(q.queue, url)
	//logM(fmt.Sprintf("Enqueueing %s len(q)", url, len(q.queue)))
}

func (q *Queue) Dequeue() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.queue) <= 0 {
		return "", false
	}
	val := q.queue[0]
	q.queue = q.queue[1:]
	return val, true
}
