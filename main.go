package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
)

type webSocketHandler struct {
	upgrader websocket.Upgrader
}

var Clients []*Client
var q = &Queue{}
var Metadata = &ComputeData{}

const CHUNK_SIZE uint8 = 100
const MASTER_KEY string = "berk"

func HandleClientConnection(c *websocket.Conn, name string, ip string) *Client {
	for _, clientPtr := range Clients {
		if clientPtr.Name == name || clientPtr.Ip == ip {
			clientPtr.Name = name
			clientPtr.Ip = ip
			return clientPtr
		}
	}

	// user not found so create new user
	cPtr := ClientConstructor(c, name, ip)
	Clients = append(Clients, cPtr)
	return cPtr
}

func HandleClient(c *websocket.Conn, message []byte, client *Client) {
	// parse client message from json
	var clientMsg ClientMessage
	if err := json.Unmarshal(message, &clientMsg); err != nil {
		logM(fmt.Sprintf("bad json: %s", err))
		return
	}

	//logM(fmt.Sprintf("Recieve message: %s.", string(message)))

	if clientMsg.Article.Content != "" {
		// probably asynchroniously save to sqlite
	}

	// update client
	client.ComputeHours += clientMsg.Article.ComputeHours
	client.State = clientMsg.State

	if client.State == Idle {
		AssignWorkToClient(c, client)
	}

}

func AssignWorkToClient(c *websocket.Conn, client *Client) {
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

	// send urls to client
	msg := MasterMessage{
		Urls:       urls,
		Command:    "process",
		ClientName: client.Name, // can be used to check if message was sent to the correct client
	}

	err := SendMessage(c, msg)

	if err != nil {
		return
	}

	// update client state to working if success
	client.State = Working
}

func SendMessage[T any](c *websocket.Conn, msg T) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		logM(fmt.Sprintf("Error: Could not construct json message %s", err))
		return err
	}

	return c.WriteMessage(websocket.TextMessage, jsonData)
}

func AssignWorkToIdle(c *websocket.Conn) {
	for _, client := range Clients {
		if client.State == Idle {
			AssignWorkToClient(c, client)
		}
	}
}

func HandleMaster(c *websocket.Conn, message []byte, client *Client) {
	var masterMsg MasterMessage
	if err := json.Unmarshal(message, &masterMsg); err != nil {
		logM(fmt.Sprintf("bad json: %s", err))
		logM(fmt.Sprintf("msg: %s", message))
		return
	}

	if masterMsg.Command == "process" || masterMsg.Command == "" {
		// load article urls into queue
		for _, url := range masterMsg.Urls {
			q.Enqueue(url)
		}
		return
	}

	// send stop command to client
	clientName := masterMsg.ClientName
	msg := MasterMessage{
		Urls:       masterMsg.Urls,
		Command:    masterMsg.Command,
		ClientName: clientName, // can be used to check if message was sent to the correct client
	}

	if msg.Urls != nil {

	}
	//err := SendMessage(c, msg) // this should send a message to the target client right?
	//if err != nil {
	//	return
	//}
}

func GetMasterClient() (*Client, error) {
	for _, client := range Clients {
		if client.Name == MASTER_KEY {
			return client, nil
		}
	}

	return nil, fmt.Errorf("No master found")
}

func UpdateMasterClient() error {
	master, err := GetMasterClient()

	if err != nil {
		return err
	}

	UMM := UpdateMasterMessage{
		Clients:  Clients,
		Metadata: Metadata,
	}

	return SendMessage(master.WsConnection, UMM)
}

func (wsh webSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		logM(fmt.Sprintf("Failed to get client IP: %v", err))
	}
	name := r.URL.Query().Get("name")

	// confirm websocket handshake
	c, err := wsh.upgrader.Upgrade(w, r, nil)

	// retrieve or create new client in clients list
	client := HandleClientConnection(c, name, ip)

	if err != nil {
		logM(fmt.Sprintf("error %s when upgrading connection to websocket", err))
		return
	}
	defer c.Close()

	// main server loop
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			logM(fmt.Sprintf("Error %s when reading message from client", err))
			continue
		}
		if mt == websocket.BinaryMessage {
			err = c.WriteMessage(websocket.TextMessage, []byte("server does not support binary messages."))
			if err != nil {
				logM(fmt.Sprintf("Error %s when sending message to client", err))
			}
			continue
		}

		if client.Name == MASTER_KEY {
			HandleMaster(c, message, client)
		} else {
			HandleClient(c, message, client)
		}
		// update master client
		err = UpdateMasterClient()
		if err != nil {
			logM(fmt.Sprintf("Error: %s", err))
		}

		// check for any idleness
		AssignWorkToIdle(c)

	}

}

func main() {
	webSocketHandler := webSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	http.Handle("/", webSocketHandler)
	logM(fmt.Sprint("Starting server..."))
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))

}
