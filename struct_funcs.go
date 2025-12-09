package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"strings"
)

func ClientConstructor(c *websocket.Conn, name, ip string) *Client {
	return &Client{
		Name:              name,
		Ip:                ip,
		ComputeHours:      0,
		ArticlesProcessed: 0,
		State:             Idle,
		WsConnection:      c,
	}
}

func (cl *ClientList) Add(client *Client) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.Clients = append(cl.Clients, client)
}

func (cl *ClientList) FindByNameOrIp(name, ip string) *Client {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	for _, clientPtr := range cl.Clients {
		if clientPtr.Name == name { // optional:  || clientPtr.Ip == ip
			clientPtr.Name = name
			clientPtr.Ip = ip
			return clientPtr
		}
	}
	return nil
}

func (cl *ClientList) FindByName(name string) *Client {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	for _, clientPtr := range cl.Clients {
		if clientPtr.Name == name { // optional:  || clientPtr.Ip == ip
			clientPtr.Name = name
			return clientPtr
		}
	}
	return nil
}

func (cl *ClientList) AssignWorkToIdle() {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	for _, client := range cl.Clients {
		logM(fmt.Sprintf("Checking client: %s", client.Name))
		if client.State == Idle {
			cl.AssignWorkToClient(client)
		}
	}
}

func (cl *ClientList) AssignWorkToClient(client *Client) {
	logM(fmt.Sprintf("Assigning work to client: %s", client.Name))

	var urls []string
	var i uint8 = 0
	for {
		url, success := q.Dequeue()
		if success == false || i >= CHUNK_SIZE {
			break
		}
		urls = append(urls, url)
		i += 1
	}

	// logM(fmt.Sprintf("client: %s", client.Name))

	if urls == nil {
		logM(fmt.Sprintf("Queue empty, no work to assign."))
		return
	}

	logM(fmt.Sprintf("assigning urls: %s", strings.Join(urls, ", ")))

	// send urls to client
	msg := MasterMessage{
		Urls:       urls,
		Command:    Command{"process", ""},
		ClientName: client.Name, // can be used to check if message was sent to the correct client
	}

	err := SendMessage(client.WsConnection, msg)

	if err != nil {
		return
	}

	// update client state to working if success
	client.State = Working
}

func (cl *ClientList) GetMasterClient() (*Client, error) {
	for _, client := range cl.Clients {
		if client.Name == MASTER_KEY {
			return client, nil
		}
	}

	return nil, fmt.Errorf("No master found")
}

func (cl *ClientList) GetClientList() []*Client {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return cl.Clients
}

func (q *Queue) Length() int16 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return int16(len(q.queue))
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

func (cd *ComputeData) Add(computeHr, articleCnt int16) {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	cd.TotalComputeHour += computeHr
	cd.TotalArticlesProcessed += articleCnt
}
