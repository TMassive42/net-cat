package utils

import (
	"bufio"
	"net"
	"sync"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		Name    string
		want    *Server
		wantErr bool
	}{
		{
			Name: "Successful server creation",
			want: &Server{
				Clients:    make(map[*Client]bool),
				History:    make([]string, 0, 100),
				maxClients: 10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := NewServer()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == nil {
				t.Fatal("expected non-nil Server instance")
			}

			// Validate individual fields
			if len(got.Clients) != len(tt.want.Clients) {
				t.Errorf("NewServer().clients = %v, want %v", got.Clients, tt.want.Clients)
			}

			if cap(got.History) != cap(tt.want.History) {
				t.Errorf("NewServer().History capacity = %d, want %d", cap(got.History), cap(tt.want.History))
			}

			if got.maxClients != tt.want.maxClients {
				t.Errorf("NewServer().maxClients = %d, want %d", got.maxClients, tt.want.maxClients)
			}

			if got.Broadcast == nil {
				t.Error("NewServer().Broadcast channel is nil, expected initialized channel")
			}
		})
	}
}

func TestServer_removeClient(t *testing.T) {
	type fields struct {
		Clients    map[*Client]bool
		Broadcast  chan string
		History    []string
		mu         sync.RWMutex
		maxClients int
		logo       string
	}
	type args struct {
		client *Client
	}
	tests := []struct {
		Name   string
		fields fields
		args   args
	}{
		{
			Name: "Remove existing client",
			fields: fields{
				Clients: map[*Client]bool{
					&Client{Name: "Client1"}: true,
				},
				Broadcast:  make(chan string),
				History:    []string{},
				maxClients: 10,
				logo:       "Test Logo",
			},
			args: args{
				client: &Client{Name: "Client1"},
			},
		},
		{
			Name: "Remove non-existing client",
			fields: fields{
				Clients: map[*Client]bool{
					&Client{Name: "Client1"}: true,
				},
				Broadcast:  make(chan string),
				History:    []string{},
				maxClients: 10,
				logo:       "Test Logo",
			},
			args: args{
				client: &Client{Name: "NonExistingClient"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			s := &Server{
				Clients:    tt.fields.Clients,
				Broadcast:  tt.fields.Broadcast,
				History:    tt.fields.History,
				mu:         tt.fields.mu,
				maxClients: tt.fields.maxClients,
				logo:       tt.fields.logo,
			}
			s.removeClient(tt.args.client)

			// Verify client was removed for "Remove existing client" case
			if tt.Name == "Remove existing client" {
				if _, exists := s.Clients[tt.args.client]; exists {
					t.Errorf("Client was not removed from Clients map")
				}
			}

			// Verify no panic or errors for "Remove non-existing client" case
			if tt.Name == "Remove non-existing client" {
				// Just ensure no errors or panics occurred during the run
			}
		})
	}
}

func TestServer_Run(t *testing.T) {
	// Set up channels and server fields
	broadcast := make(chan string, 1)
	client1 := &Client{Messages: make(chan string, 1)}
	client2 := &Client{Messages: make(chan string, 1)}         // simulate a responsive client
	unresponsiveClient := &Client{Messages: make(chan string)} // simulate an unresponsive client

	clients := map[*Client]bool{
		client1:            true,
		client2:            true,
		unresponsiveClient: true,
	}

	server := &Server{
		Clients:   clients,
		Broadcast: broadcast,
		History:   []string{},
		mu:        sync.RWMutex{},
	}

	// Run the server in a goroutine
	go server.Run()

	// Test appending to history and broadcasting to clients
	messages := "Test messages"
	broadcast <- messages

	// Allow time for the server to process
	time.Sleep(100 * time.Millisecond)

	// Check messages was added to history
	server.mu.RLock()
	if len(server.History) == 0 || server.History[0] != messages {
		t.Errorf("Expected messages %q in history, but got %v", messages, server.History)
	}
	server.mu.RUnlock()

	// Check responsive clients received the messages
	select {
	case msg := <-client1.Messages:
		if msg != messages {
			t.Errorf("Expected messages %q for client1, got %q", messages, msg)
		}
	default:
		t.Errorf("client1 did not receive the messages")
	}

	select {
	case msg := <-client2.Messages:
		if msg != messages {
			t.Errorf("Expected messages %q for client2, got %q", messages, msg)
		}
	default:
		t.Errorf("client2 did not receive the messages")
	}

	// Check that the unresponsive client was removed
	server.mu.RLock()
	_, exists := server.Clients[unresponsiveClient]
	server.mu.RUnlock()
	if exists {
		t.Errorf("Expected unresponsive client to be removed, but it was still in Clients map")
	}
}

func TestServer_addClient(t *testing.T) {
	type fields struct {
		Clients    map[*Client]bool
		Broadcast  chan string
		History    []string
		mu         sync.RWMutex
		maxClients int
		logo       string
	}
	type args struct {
		client *Client
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Add client with unique name within capacity",
			fields: fields{
				Clients:    make(map[*Client]bool),
				Broadcast:  make(chan string),
				History:    []string{},
				maxClients: 2,
			},
			args: args{
				client: &Client{Name: "Client1"},
			},
			want: true,
		},
		{
			name: "Fail to add client when max capacity is reached",
			fields: fields{
				Clients: map[*Client]bool{
					&Client{Name: "Client1"}: true,
					&Client{Name: "Client2"}: true,
				},
				Broadcast:  make(chan string),
				History:    []string{},
				maxClients: 2,
			},
			args: args{
				client: &Client{Name: "Client3"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				Clients:    tt.fields.Clients,
				Broadcast:  tt.fields.Broadcast,
				History:    tt.fields.History,
				mu:         tt.fields.mu,
				maxClients: tt.fields.maxClients,
				logo:       tt.fields.logo,
			}
			if got := s.addClient(tt.args.client); got != tt.want {
				t.Errorf("Server.addClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_HandleConnection(t *testing.T) {
	type fields struct {
		Clients    map[*Client]bool
		Broadcast  chan string
		History    []string
		mu         sync.RWMutex
		maxClients int
		logo       string
	}
	type args struct {
		conn net.Conn
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantMsg string
		wantAdd bool
	}{
		{
			name: "Normal Connection",
			fields: fields{
				Clients:    make(map[*Client]bool),
				Broadcast:  make(chan string, 10),
				History:    []string{"Welcome to the chat!", "Chat started..."},
				maxClients: 2,
				logo:       "Welcome to the Server!\n",
			},
			wantMsg: "Welcome to the Server!\n",
			wantAdd: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize server based on test case fields
			s := &Server{
				Clients:    tt.fields.Clients,
				Broadcast:  tt.fields.Broadcast,
				History:    tt.fields.History,
				maxClients: tt.fields.maxClients,
				logo:       tt.fields.logo,
			}

			// Create an in-memory connection using net.Pipe
			serverConn, clientConn := net.Pipe()
			defer serverConn.Close()
			defer clientConn.Close()

			// Run HandleConnection in a goroutine
			go s.HandleConnection(serverConn)

			// Give some time for goroutine to process
			time.Sleep(100 * time.Millisecond)

			// Read messages sent to client
			clientReader := bufio.NewReader(clientConn)
			messages, _ := clientReader.ReadString('\n')

			// Validate expected messages
			if messages != tt.wantMsg {
				t.Errorf("expected messages %q, got %q", tt.wantMsg, messages)
			}
		})
	}
}

func TestServer_HandleClientMessagess(t *testing.T) {
	type fields struct {
		Clients    map[*Client]bool
		Broadcast  chan string
		History    []string
		mu         sync.RWMutex
		maxClients int
		logo       string
	}
	type args struct {
		client *Client
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				Clients:    tt.fields.Clients,
				Broadcast:  tt.fields.Broadcast,
				History:    tt.fields.History,
				mu:         tt.fields.mu,
				maxClients: tt.fields.maxClients,
				logo:       tt.fields.logo,
			}
			s.HandleClientMessages(tt.args.client)
		})
	}
}
