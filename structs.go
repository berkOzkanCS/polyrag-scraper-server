package main

import (
	"github.com/gorilla/websocket"
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
	WsConnection      *websocket.Conn
}

type ComputeData struct {
	mu                     sync.Mutex
	TotalComputeHour       int16 `json:"totalComputeHour"`
	TotalArticlesProcessed int16 `json:"totalArticlesProcessed"`
}

type ClientList struct {
	mu      sync.Mutex
	Clients []*Client
}

type Queue struct {
	mu    sync.Mutex
	queue []string
}

type Command struct {
	Cmd    string `json:"cmd"`
	Params string `json:"parameters"`
}

type InsertJob struct {
	Title   string
	Content string
	Date    string
	Done    chan error
}
