package main

import (
	//"encoding/json"
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

func HandleClientConnection(name string, ip string) *Client {
	for _, clientPtr := range Clients {
		if clientPtr.Name == name || clientPtr.Ip == ip {
			clientPtr.Name = name
			clientPtr.Ip = ip
			return clientPtr
		}
	}

	// user not found so create new user
	cPtr := ClientConstructor(name, ip)
	Clients = append(Clients, cPtr)
	return cPtr
}

func (wsh webSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		logM(fmt.Sprintf("Failed to get client IP: %v", err))
	}
	name := r.URL.Query().Get("name")

	// retrieve or create new client in clients list
	client := HandleClientConnection(name, ip)

	if client != nil {

	}
	// confirm websocket handshake
	c, err := wsh.upgrader.Upgrade(w, r, nil)
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
			return
		}
		if mt == websocket.BinaryMessage {
			err = c.WriteMessage(websocket.TextMessage, []byte("server does not support binary messages."))
			if err != nil {
				logM(fmt.Sprintf("Error %s when sending message to client", err))
			}
			return
		}

		// parse client message from json
		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			logM(fmt.Sprintf("bad json: %s", err))
			continue
		}

		logM(fmt.Sprintf("Recieve message: %s.", string(message)))

		if clientMsg.Article.Content != "" {
			// probably asynchroniously save to sqlite
		}

		// update client
		client.ComputeHours += clientMsg.Article.ComputeHours
		client.State = clientMsg.State

		if client.State == Idle {
			// assign work from work buffer via round robbin
		}

		//	if strings.Trim(string(message), "\n") != "start" {
		//		err = c.WriteMessage(websocket.TextMessage, []byte("You did not say the magic word!"))
		//		if err != nil {
		//			log.Printf("Error %s when sending message to client", err)
		//			return
		//		}
		//		continue
		//	}
		//	logM(fmt.Sprintln("Start responding to client..."))
		//	i := 1
		//	for {
		//		response := fmt.Sprintf("Notification %d", i)
		//		err := c.WriteMessage(websocket.TextMessage, []byte(response))
		//		if err != nil {
		//			logM(fmt.Sprintf("Error %s when sending message to client", err))
		//			return
		//		}
		//		i = i + 1
		//		time.Sleep(2 * time.Second)
		//	}
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
