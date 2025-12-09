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

var cl = &ClientList{}
var q = &Queue{}
var Metadata = &ComputeData{}

const CHUNK_SIZE uint8 = 100
const MASTER_KEY string = "berk"

func HandleClientConnection(c *websocket.Conn, name string, ip string) *Client {
	client := cl.FindByNameOrIp(name, ip)

	if client != nil {
		return client
	}

	// user not found so create new user
	cPtr := ClientConstructor(c, name, ip)
	cl.Add(cPtr)
	return cPtr
}

func HandleClient(message []byte, client *Client) {
	// parse client message from json
	var clientMsg ClientMessage
	if err := json.Unmarshal(message, &clientMsg); err != nil {
		logM(fmt.Sprintf("bad json: %s", err))
		return
	}

	logM(fmt.Sprintf("Recieve message: %s.", string(message)))

	if clientMsg.Article.Content != "" {
		// probably asynchroniously save to sqlite
		logM("Article Saved!")
	}

	// update client
	client.ComputeHours += clientMsg.Article.ComputeHours
	client.State = clientMsg.State

	Metadata.Add(int16(clientMsg.Article.ComputeHours), 1)

	if client.State == Idle {
		cl.AssignWorkToClient(client)
	}

}

func SendMessage[T any](c *websocket.Conn, msg T) error {
	logM("Sending message to cclient.")
	jsonData, err := json.Marshal(msg)
	if err != nil {
		logM(fmt.Sprintf("Error: Could not construct json message %s", err))
		return err
	}

	return c.WriteMessage(websocket.TextMessage, jsonData)
}

func HandleMaster(message []byte) {
	var masterMsg MasterMessage
	if err := json.Unmarshal(message, &masterMsg); err != nil {
		logM(fmt.Sprintf("bad json: %s", err))
		logM(fmt.Sprintf("msg: %s", message))
		return
	}

	if masterMsg.Command.Cmd == "process" || masterMsg.Command.Cmd == "" {
		// load article urls into queue
		for _, url := range masterMsg.Urls {
			q.Enqueue(url)
			logM(fmt.Sprintf("Successfuly queued %s", url))
		}
		logM(fmt.Sprintf("Exitin Master Msg Handler"))
		return
	} else {
		// send command to client
		clientName := masterMsg.ClientName
		msg := MasterMessage{
			Urls:       nil,
			Command:    masterMsg.Command,
			ClientName: clientName, // can be used to check if message was sent to the correct client
		}
		clientToSendMsg := cl.FindByName(clientName)
		if clientToSendMsg != nil {
			err := SendMessage(clientToSendMsg.WsConnection, msg)
			if err != nil {
				logM(fmt.Sprintf("%s", err))
			}
		}
		logM(fmt.Sprintf("Could not find clinet by the name %s", clientName))
	}
}

func UpdateMasterClient() error {
	master, err := cl.GetMasterClient()

	if err != nil {
		return err
	}

	UMM := UpdateMasterMessage{
		Clients:    cl.GetClientList(),
		Metadata:   Metadata,
		QueueLenth: q.Length(),
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

	if client.Name == MASTER_KEY {
		msg := fmt.Sprintf("Master Client %s Connected!", MASTER_KEY)
		err = c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			logM(fmt.Sprintf("Error %s when sending message to client", err))
		}
	}

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
			HandleMaster(message)
		} else {
			HandleClient(message, client)
		}

		// update master
		err = UpdateMasterClient()
		logM("Master client updated")
		if err != nil {
			logM(fmt.Sprintf("Error: %s", err))
		}

		// check for any idleness
		logM("Checking for idle workers")
		cl.AssignWorkToIdle()

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
