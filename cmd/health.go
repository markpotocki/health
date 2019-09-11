package main

import (
	"github.com/markpotocki/health/internal/server"
	"github.com/markpotocki/health/internal/server/store"
)

func main() {
	cs := store.MakeClientStore()
	ss := store.MakeStatusStore()

	srv := server.MakeServer(cs, ss)

	srv.Start()
}
