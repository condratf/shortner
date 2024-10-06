package main

import (
	"fmt"

	"github.com/condratf/shortner/internal/app"
)

func main() {
	fmt.Println("start")

	if app.Server() != nil {
		panic("server has crashed")
	}
}
