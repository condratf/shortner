package main

import (
	"github.com/condratf/shortner/internal/app"
)

func main() {
	err := app.Server()
	if err != nil {
		panic("server has crashed")
	}
}
