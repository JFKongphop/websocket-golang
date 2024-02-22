package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type subscription struct {
	conn *connection
	room string
}

type connection struct {
	// The websocket connection.
	ws *websocket.Conn
	// Buffered channel of outbound messages.
	send chan []byte
}

func (s *subscription) readPump() {
	c := s.conn
	defer func() {
		//Unregister
		H.unregister <- *s
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		//Reading incoming message...
		_, msg, err := c.ws.ReadMessage()
		fmt.Println("test msg", msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		fmt.Println("Received message from client:", string(msg))

		msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
		m := message{s.room, msg}
		H.broadcast <- m
	}
}
// func (s *subscription) writePump() {
// 	c := s.conn
// 	ticker := time.NewTicker(pingPeriod)
// 	defer func() {
// 		ticker.Stop()
// 		c.ws.Close()
// 	}()
// 	fmt.Println("send message")
// 	for {
// 		select {
// 		//Listerning message when it comes will write it into writer and then send it to the client
// 		case message, ok := <-c.send:
// 			if !ok {
// 				c.write(websocket.CloseMessage, []byte{})
// 				return
// 			}
// 			if err := c.write(websocket.TextMessage, message); err != nil {
// 				fmt.Println(websocket.TextMessage)
// 				return
// 			}
// 		case <-ticker.C:
// 			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
// 				return
// 			}
// 		}
// 	}
// }

// func (s *subscription) writePump() {
// 	c := s.conn
// 	ticker := time.NewTicker(pingPeriod)
// 	defer func() {
// 		ticker.Stop()
// 		c.ws.Close()
// 	}()
// 	fmt.Println("send message")
// 	for {
// 		select {
// 		// Listening for messages when they come and then sending them to the client
// 		case message, ok := <-c.send:
// 			if !ok {
// 				c.write(websocket.CloseMessage, []byte{})
// 				return
// 			}
			
// 			// Add timestamp to the message
			

// 			for i := 1; i < 5; i++ {
// 				messageWithTime := fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), message)
// 				if err := c.write(websocket.TextMessage, []byte(messageWithTime)); err != nil {
// 					fmt.Println(websocket.TextMessage)
// 					return
// 				}

// 				time.Sleep(time.Second * 2)
// 			}
			

// 		case <-ticker.C:
// 			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
// 				return
// 			}
// 		}
// 	}
// }

func (s *subscription) writePump() {
	c := s.conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	fmt.Println("send message")
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// Channel closed, connection is closed
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				fmt.Println(websocket.TextMessage)
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}


func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	fmt.Println(payload)
	return c.ws.WriteMessage(mt, payload)
}


// func ServeWs(w http.ResponseWriter, r *http.Request) {
// 	ws, err := upgrader.Upgrade(w, r, nil)
// 	//Get room's id from client...
// 	queryValues := r.URL.Query()
// 	roomId := queryValues.Get("roomId")
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	c := &connection{send: make(chan []byte, 256), ws: ws}
// 	s := subscription{c, roomId}
// 	log.Println("test", roomId)
// 	H.register <- s
// 	go s.writePump()
// 	go s.readPump()
// }

func ServeWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	// Get room's id from client...
	queryValues := r.URL.Query()
	roomId := queryValues.Get("roomId")
	if err != nil {
		log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws}
	s := subscription{c, roomId}
	log.Println("test", roomId)
	H.register <- s
	go s.writePump()
	go s.readPump()

	// Automatically send a message to the client after joining the room
	// go func() {
	// 	for {
	// 		message := "This is an automated message."
	// 		s.conn.send <- []byte(message)
	// 		time.Sleep(3 * time.Second)
	// 	}
	// }()

	go func() {
    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()
    for range ticker.C {
        message := "This is an automated message."
        s.conn.send <- []byte(message)
    }
	}()

}
