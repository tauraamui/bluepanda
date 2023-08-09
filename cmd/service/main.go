package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tauraamui/redpanda/internal/service"
)

func main() {
	svr, err := service.New()
	if err != nil {
		log.Fatal().Msgf("error: %s", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
		log.Info().Msgf("defaulting to port %s", port)
	}

	go func() {
		if err := svr.Listen(":" + port); err != nil {
			log.Fatal().Msgf("error: %s", err)
		}
	}()

	log.Info().Msg("redpanda started, waiting for interrupt...")

	<-interrupt

	log.Info().Msg("shutting down gracefully...")
	if err := svr.Cleanup(); err != nil {
		log.Fatal().Msgf("error: %s", err)
	}

	if err := svr.ShutdownWithTimeout(60 * time.Second); err != nil {
		log.Fatal().Msgf("error: %s", err)
	}

	log.Info().Msg("shut down... done")
}
