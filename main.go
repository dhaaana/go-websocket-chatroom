package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dhaaana/go-websocket-chatroom/config"
)

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		fmt.Println(err)
		return
	}
	http.HandleFunc("/", serveLanding)
	http.HandleFunc("/chat", serveChatRoom)

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
