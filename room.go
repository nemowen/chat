package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/nemowen/trace"
)

type room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients.
	forward chan []byte
	// join is a channel for clients wishing to join the room.
	join chan *client
	// leave is a channel for clients wishing to leave the room.
	leave chan *client
	// clients holds all current clients in this room.
	clients map[*client]bool
	// tracer will receive trace information of activity
	// in the room.
	tracer trace.Tracer
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  socketBufferSize,
		WriteBufferSize: socketBufferSize,
	}
)

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// joining
			r.clients[client] = true
			r.tracer.Trace("new client joined")
		case client := <-r.leave:
			// leaving
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("a client left")
		case msg := <-r.forward:
			r.tracer.Trace("received message: ", string(msg))
			// forward message to all clients
			for client := range r.clients {
				select {
				case client.send <- msg:
					// send the message
					r.tracer.Trace("-- send message to client")
				default:
					// failed to send
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace("-- failed to send, cleared up client")
				}
			}
		}
	}
}

func (room *room) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("websocket.upgrader:", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   room,
	}

	room.join <- client
	defer func() { room.leave <- client }()
	go client.write()
	client.read()
}
