package main

import (
	"log"

	"github.com/condratf/shortner/internal/app"
)

func main() {
	err := app.Server()
	if err != nil {
		log.Fatal("server has crashed")
	}
}
