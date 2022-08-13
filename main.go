package main

import (
	"bytes"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os/exec"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1920,
	WriteBufferSize: 1920,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func setupRoutes() {
	log.Println("Setting up routes..")
	fs := http.FileServer(http.Dir("./site"))
	http.Handle("/", fs)
	http.HandleFunc("/aquarium", aquarium)
}

func aquarium(writer http.ResponseWriter, request *http.Request) {
	log.Println("Entered aquarium handler.")
	log.Println("Upgrading connection.")
	ws, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Fatalln(err)
		return
	}
	go spawnAquarium(ws)
}

func spawnAquarium(ws *websocket.Conn) {
	log.Println("Spawning aquarium..")
	cmd := exec.Command("/usr/bin/asciiquarium")
	buf := bytes.Buffer{}
	cmd.Stdout = &buf

	err := cmd.Start()
	if err != nil {
		log.Fatalln(err)
		return
	}
	for {
		time.Sleep(50 * time.Millisecond)
		tx, txErr := ws.NextWriter(websocket.TextMessage)
		// Open handle for writer
		if txErr != nil {
			// Connection must have been closed.
			log.Println(txErr)
			log.Println("Waiting to asciiquarium to shut down...")
			cmd.Wait()
			log.Println("asciiquarium shut down.")
			return
		}
		tx.Write(buf.Bytes())
		tx.Close()
	}
}

func main() {
	setupRoutes()
	log.Println("Serving at :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Println(err)
		return
	}
}
::