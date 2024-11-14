package utils

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

func NewClient(conn net.Conn, s *Server) *Client {
	reader := bufio.NewReader(conn)

	conn.Write([]byte(""))
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	client := &Client{
		Conn:     conn,
		Name:     name,
		Messages: make(chan string, 10),
	}
	return client
}

func (c *Client) SendMessages() {
	for msg := range c.Messages {
		_, err := c.Conn.Write([]byte(msg))
		if err != nil {
			break
		}
	}
}

func (c *Client) Listen(broadcast chan<- string) {
	reader := bufio.NewReader(c.Conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			broadcast <- fmt.Sprintf("%s has left our chat...\n", c.Name)
			break
		}
		message = strings.TrimSpace(message)
		if message == "" {
			continue
		}
		formattedMsg := fmt.Sprintf("[%s][%s]: %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Name,
			message)
		broadcast <- formattedMsg
	}
}
