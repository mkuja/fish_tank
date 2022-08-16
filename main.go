package main

import (
	"fmt"
	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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

func pError(err *error) bool {
	if err != nil {
		log.Println(err)
		return true
	}
	return false
}

func sendToClient(ws *websocket.Conn, data []byte) (int, error) {
	writer, err := ws.NextWriter(websocket.TextMessage)
	if err != nil {
		pError(&err)
		return 0, err
	}
	sent, err := writer.Write(data)
	if err != nil {
		pError(&err)
		return 0, err
	}
	writer.Close()
	return sent, nil
}

func readFromClient(ws *websocket.Conn, buf []byte) (int, error) {
	_, reader, err := ws.NextReader()
	if err != nil {
		pError(&err)
		return 0, err
	}
	read, err := reader.Read(buf)
	if err != nil {
		return 0, err
	}
	return read, nil
}

func spawnAquarium(ws *websocket.Conn) {
	fmt.Fprintf(os.Stderr, "Spawning asciiquarium\n")
	c := exec.Command("asciiquarium")
	p, err0 := pty.Start(c)
	defer c.Process.Kill()
	if err0 != nil {
		pError(&err0)
		return
	}

	for {
		buf := make([]byte, 1024)
		read, err := p.Read(buf)
		if err == io.EOF {
			return
		}
		_, err = sendToClient(ws, buf[:read])
		if err != nil {
			return
		}
		time.Sleep(time.Millisecond * 20)
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
