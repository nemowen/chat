package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/nemowen/trace"
	"github.com/stretchr/objx"
)

type room struct {
	// 转发消息给其它它客户端的通道
	forward chan *message
	// 要客户端加入到room的通道
	join chan *client
	// 要退出的客户端通道
	leave chan *client
	// 所有客户端的map集合
	clients map[*client]bool
	// 跟踪消息对象
	tracer trace.Tracer
	// 关于用户的头像
	avatar Avatar
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
		forward: make(chan *message),
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
			r.tracer.Trace("received message: ", msg.Name, msg.When, msg.Message)
			// 转发消息到所有用户
			for client := range r.clients {
				select {
				case client.send <- msg:
					// 发送消息成功
					r.tracer.Trace("-- send message to client")
				default:
					// 发送失败,删除用户
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
	authCookie, err := r.Cookie("auth")
	if err != nil {
		log.Fatal("Failed to get auth cookie:", err)
		return
	}
	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize),
		room:     room,
		userData: objx.MustFromBase64(authCookie.Value),
	}

	room.join <- client
	defer func() { room.leave <- client }()
	go client.write()
	client.read()
}
