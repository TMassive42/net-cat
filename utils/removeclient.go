package utils

func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.Clients[client]; ok {
		delete(s.Clients, client)
		close(client.Messages)
		client.Conn.Close()
	}
}
