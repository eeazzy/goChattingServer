package network

type data struct {
}

func registerServer(server *Server) {
	d := &data{}

	r := NewRoom()
	go r.Run()

	engine.GET("/room", r.ServeHTTP)

	return d
}
