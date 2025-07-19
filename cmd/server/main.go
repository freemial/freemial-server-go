package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/freemial/freemial-server-go/internal/api"
	"github.com/freemial/freemial-server-go/internal/websocket"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()
	hub := websocket.NewHub()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})
	http.HandleFunc("/login", api.Login)
	http.HandleFunc("/device/bindings", func(w http.ResponseWriter, r *http.Request) {
		api.GetDeviceBindings(hub, w, r)
	})

	log.Println("Going to listen on: " + *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
