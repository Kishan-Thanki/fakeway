package main

import (
	"log"

	"github.com/kishan-thanki/fakeway/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
