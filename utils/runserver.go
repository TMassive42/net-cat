package utils

func (s *Server) Run() {
	for message := range s.Broadcast {
		s.mu.Lock()
		s.History = append(s.History, message)
		s.mu.Unlock()

		for client := range s.Clients {
			select {
			case client.Messages <- message:
			default:
				s.removeClient(client)
			}
		}
	}
}