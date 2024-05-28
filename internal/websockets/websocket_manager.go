package websockets

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	connections = make(map[int]*websocket.Conn)
	lock        = sync.Mutex{}
)

func AddConnection(id int, conn *websocket.Conn) {
	lock.Lock()
	defer lock.Unlock()
	connections[id] = conn

	fmt.Println("connections: ", connections)
}

func RemoveConnection(id int) {
	lock.Lock()
	defer lock.Unlock()
	if _, exists := connections[id]; exists {
		delete(connections, id)
	}
}

func GetConnection(id int) (*websocket.Conn, bool) {
	lock.Lock()
	defer lock.Unlock()
	conn, exists := connections[id]
	return conn, exists
}
