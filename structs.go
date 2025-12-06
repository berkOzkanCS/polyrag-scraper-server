package main

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
