package utils

import (
	"fmt"
	"net"
)

func (s *Server) HandleConnection(conn net.Conn) {
	conn.Write([]byte(s.logo))

	client := NewClient(conn, s)
	if !s.addClient(client) {
		conn.Write([]byte("Server is full. Please try agai later.\n"))
		conn.Close()
		return
	}

	s.Broadcast <- fmt.Sprintf("%s has joined our chat...\n", &client.Name)

	s.mu.RLock()
	for _, msg := range s.History {
		client.Messages <- msg
	}
	s.mu.RUnlock()

	go s.HandleClientMessages(client)
	go client.SendMessages()
}

