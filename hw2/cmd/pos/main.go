package main

import (
	"log"
	"os"

	"github.com/dushparra/kpo-2026/hw2/internal/app"
	"github.com/dushparra/kpo-2026/hw2/internal/console"
	"github.com/dushparra/kpo-2026/hw2/internal/usecase"
)

func main() {
	container, err := app.NewContainer()
	if err != nil {
		log.Fatal(err)
	}
	if err := container.Invoke(func(service *usecase.Service) error {
		return console.Run(service, os.Stdin, os.Stdout)
	}); err != nil {
		log.Fatal(err)
	}
}
