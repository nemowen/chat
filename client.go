package main

import (
	"github.com/gorilla/websocket"
	"time"
)

// client 代表一个聊天的用户.
type client struct {
	// 一个网页socket连接.
	socket *websocket.Conn
	// 发送消息的.
	send chan *message
	// 用户报在的房间.
	room *room
	// 用户数据
	userData map[string]interface{}
}

func (c *client) read() {
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil {
			msg.Name = c.userData["name"].(string)
			msg.When = time.Now()
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()

}
