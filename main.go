package main

import (
	"fmt"
	"log"
	"nc/utils"
	"net"
	"os"
)

func main() {
	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	port := "8989"
	if len(os.Args) == 2 {
		port = os.Args[1]
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	srv, err := utils.NewServer()
	if err != nil {
		log.Fatalf("Error creating server: %v", err)
	}

	fmt.Printf("Listening on the port :%s\n", port)

	go srv.Run()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go srv.HandleConnection(conn)
	}
}
