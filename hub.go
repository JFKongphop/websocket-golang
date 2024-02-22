package main

import (
	"fmt"
	// "log"
	// "sync" // Import sync package for synchronization
)

// Hub maintains the set of active clients and broadcasts messages to the
// type Hub struct {
// 	// put registered clients into the room.
// 	rooms map[string]map[*connection]bool
// 	// Inbound messages from the clients.
// 	broadcast chan message

// 	// Register requests from the clients.
// 	register chan subscription

// 	// Unregister requests from clients.
// 	unregister chan subscription

// 	activeConnection int
// }

type Hub struct {
	// Use a mutex to synchronize access to the rooms map and activeConnection count
	// mu             sync.RWMutex
	rooms          map[string]map[*connection]bool
	broadcast      chan message
	register       chan subscription
	unregister     chan subscription
	activeConns    map[string]int // Track the number of active connections across all rooms
}

type message struct {
	Room string `json:"room"`
	Data []byte `json:"data"`
}
var H = &Hub{
	broadcast:  make(chan message),
	register:   make(chan subscription),
	unregister: make(chan subscription),
	rooms:      make(map[string]map[*connection]bool),
	activeConns: make(map[string]int),
}
// func (h *Hub) Run() {
// 	fmt.Println("hub")
// 	for {
// 		select {
// 		case s := <-h.register:
// 			log.Println("register", h.rooms[s.room])
// 			connections := h.rooms[s.room]
// 			if connections == nil {
// 				connections = make(map[*connection]bool)
// 				h.rooms[s.room] = connections
// 				// h.activeConnection++
// 			}
// 			fmt.Println("register connection", connections)
// 			h.rooms[s.room][s.conn] = true
// 		case s := <-h.unregister:
// 			log.Println("unregis")
// 			connections := h.rooms[s.room]
// 			if connections != nil {
// 				if _, ok := connections[s.conn]; ok {
// 					delete(connections, s.conn)
// 					close(s.conn.send)
// 					if len(connections) == 0 {
// 						delete(h.rooms, s.room)
// 					}
// 				}
// 			}
// 			fmt.Println("unregister connection", connections)
// 		case m := <-h.broadcast:
// 			log.Println("boardcast")
// 			connections := h.rooms[m.Room]
// 			for c := range connections {
// 				select {
// 				case c.send <- m.Data:
// 				default:
// 					close(c.send)
// 					delete(connections, c)

// 					if len(connections) == 0 {
// 						delete(h.rooms, m.Room)
// 					}
// 				}
// 			}
// 			fmt.Println("boardcast connection", connections)
// 		}
// 	}
// }

func (h *Hub) Run() {
	fmt.Println("Hub running")
	for {
		select {
		case s := <-h.register:
			// h.mu.Lock()
			connections := h.rooms[s.room]
			if connections == nil {
				connections = make(map[*connection]bool)
				h.rooms[s.room] = connections
			}
			connections[s.conn] = true

			h.activeConns[s.room]++ // Increment active connections count
			// h.mu.Unlock()
			fmt.Println("connection", h.activeConns)

		case s := <-h.unregister:
			// h.mu.Lock()
			fmt.Println(s.room)
			// connections := h.rooms[s.room]
			// if connections != nil {
			// 	if _, ok := connections[s.conn]; ok {
			// 		delete(connections, s.conn)
			// 		close(s.conn.send)
			// 		h.activeConns-- // Decrement active connections count
			// 	}
			// 	if len(connections) == 0 {
			// 		delete(h.rooms, s.room)
			// 	}
			// }
			h.activeConns[s.room]--
			if h.activeConns[s.room] == 0 {
				delete(h.rooms, s.room)
				delete(h.activeConns, s.room) // Remove count for the room if it has no connections
			}
			fmt.Println("no", h.activeConns)
			// h.mu.Unlock()

		case m := <-h.broadcast:
			// h.mu.RLock()
			connections := h.rooms[m.Room]
			for c := range connections {
				select {
				case c.send <- m.Data:
				default:
					delete(connections, c)
					close(c.send)
					h.activeConns[m.Room]-- // Decrement active connections count
				}
			}
			// h.mu.RUnlock()
			
			// Check if the number of active connections is zero
			// if h.activeConns == 0 {
			// 	// h.mu.Lock()
			// 	delete(h.rooms, m.Room)
			// 	// h.mu.Unlock()
			// }
		}
	}
}