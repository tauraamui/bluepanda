package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/tauraamui/bluepanda/internal/logging"
	"github.com/tauraamui/bluepanda/internal/service"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log := logging.New()
	svr, err := service.New(log)
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

	log.Info().Msg("bluepanda started, waiting for interrupt...")

	<-interrupt

	log.Info().Msg("shutting down gracefully...")
	if err := svr.Cleanup(log); err != nil {
		log.Fatal().Msgf("error: %s", err)
	}

	if err := svr.ShutdownWithTimeout(60 * time.Second); err != nil {
		log.Fatal().Msgf("error: %s", err)
	}

	log.Info().Msg("shut down... done")
}
