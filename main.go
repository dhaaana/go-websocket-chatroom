package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dhaaana/go-websocket-chatroom/config"
	"github.com/gorilla/websocket"
)

type SocketPayload struct {
	Message string
}

type SocketResponse struct {
	From    string
	Type    string
	Message string
}

type WebSocketConnection struct {
	*websocket.Conn
	Username string
}

const (
	MESSAGE_NEW_USER = "New User"
	MESSAGE_CHAT     = "Chat"
	MESSAGE_LEAVE    = "Leave"
)

var connections = make([]*WebSocketConnection, 0)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		fmt.Println(err)
		return
	}
	http.HandleFunc("/", serveLanding)
	http.HandleFunc("/chat", serveChatRoom)
	http.HandleFunc("/ws", initializeConnection)

	fmt.Println("Server running at port " + config.Port)
	http.ListenAndServe(":"+config.Port, nil)
}

func serveLanding(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile("public/landing.html")
	if err != nil {
		http.Error(w, "Could not open requested file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", content)
}

func serveChatRoom(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile("public/chat.html")
	if err != nil {
		http.Error(w, "Could not open requested file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", content)
}

func initializeConnection(w http.ResponseWriter, r *http.Request) {
	websocketConn, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	username := r.URL.Query().Get("username")
	currentConn := WebSocketConnection{Conn: websocketConn, Username: username}
	connections = append(connections, &currentConn)

	go handleIO(&currentConn, connections)
}

func handleIO(currentConn *WebSocketConnection, connections []*WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR", fmt.Sprintf("%v", r))
		}
	}()

	broadcastMessage(currentConn, MESSAGE_NEW_USER, "")

	for {
		payload := SocketPayload{}
		err := currentConn.ReadJSON(&payload)
		if err != nil {
			// If the connection is closed, we can safely ignore the error
			if strings.Contains(err.Error(), "websocket: close") {
				broadcastMessage(currentConn, MESSAGE_LEAVE, "")
				ejectConnection(currentConn)
				return
			}

			// Otherwise, we should log the error
			log.Println("ERROR", err.Error())
			continue
		}

		broadcastMessage(currentConn, MESSAGE_CHAT, payload.Message)
	}
}

func ejectConnection(currentConn *WebSocketConnection) {
	filtered := []*WebSocketConnection{}
	for _, conn := range connections {
		if conn != currentConn {
			filtered = append(filtered, conn)
		}
	}
	connections = filtered
}

func broadcastMessage(currentConn *WebSocketConnection, kind, message string) {
	for _, eachConn := range connections {
		if eachConn == currentConn {
			continue
		}

		eachConn.WriteJSON(SocketResponse{
			From:    currentConn.Username,
			Type:    kind,
			Message: message,
		})
	}
}
