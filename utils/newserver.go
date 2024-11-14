package utils

import "fmt"

func NewServer() (*Server, error) {
	logo, err := LoadLogo()
	fmt.Println(logo)
	if err != nil {
		return nil, err
	}
	return &Server{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan string),
		History:    make([]string, 0, 100),
		maxClients: 10,
		logo:       logo,
	}, nil
}
