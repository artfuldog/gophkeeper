package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/artfuldog/gophkeeper/internal/server"
)

//nolint:gochecknoglobals
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	// Channel for handle Unix signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Main app context and signalling channel
	ctx, cancel := context.WithCancel(context.Background())
	statusCh := make(chan error)

	cfg, err := server.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	server, err := server.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Version: %s\nBuild date: %s\nBuild commit: %s\n",
		buildVersion, buildDate, buildCommit)

	go server.Run(ctx, statusCh)

	select {
	case sig := <-sigs:
		cancel()

		err := <-statusCh
		if err != nil {
			log.Printf("Server is terminated unproperly: %v, signal %s triggered", err, sig)
			return
		}

		log.Printf("Server is terminated, signal %s triggered", sig)

		return
	case err := <-statusCh:
		cancel()
		log.Fatal(err)
	}
}
