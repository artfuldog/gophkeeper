package main

import (
	"context"
	"errors"
	"log"

	"github.com/artfuldog/gophkeeper/internal/client"
)

func main() {
	flags := client.ReadFlags()

	app, err := client.NewClient(flags)
	if err != nil {
		if errors.Is(err, client.ErrShowVersion) {
			return
		}

		log.Fatal(err)
	}

	if err := app.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}
