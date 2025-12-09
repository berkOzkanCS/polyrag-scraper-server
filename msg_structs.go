package main

type ClientMessage struct { // message coming from clinet -> server
	Name    string `json:"name"`
	State   State  `json:"state"`
	Article struct {
		Title        string `json:"title"`
		Content      string `json:"content"`
		Date         string `json:"date"`
		ComputeHours int    `json:"computeHours"`
	} `json:"article"`
}

type MasterMessage struct { // message coming from master->server or server->client
	Urls       []string `json:"urls"`
	Command    Command  `json:"command"`
	ClientName string   `json:"clientName"`
}

type UpdateMasterMessage struct { // message being sent from server -> master
	Clients    []*Client `json:"clients"`
	Metadata   *ComputeData
	QueueLenth int16
}
