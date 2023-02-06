package main

import (
	"context"
	"log"

	"github.com/artfuldog/gophkeeper/internal/client"
)

func main() {
	flags := client.ReadFlags()

	client, err := client.NewClient(flags)
	if err != nil {
		log.Fatal(err)
	}
	if client == nil {
		return
	}
	if err := client.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}
