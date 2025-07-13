package main

import (
	"os"
	"os/signal"
	"runtime"
	"time"

	cp "github.com/OverlayFox/CaspawCG/casperProxy"
	googleworkspace "github.com/OverlayFox/CaspawCG/googleWorkspace"
	"github.com/rs/zerolog"
)

const (
	proxyPort    = "5251"
	casparCGHost = "127.0.0.1"
	casparCGPort = "5250"
)

func main() {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().Timestamp().Str("component", "main").Logger()
	log.Info().
		Int("pid", os.Getpid()).
		Str("os", runtime.GOOS).
		Str("arch", runtime.GOARCH).
		Str("goversion", runtime.Version()).
		Str("proxyPort", proxyPort).
		Str("casparCGHost", casparCGHost).
		Str("casparCGPort", casparCGPort).
		Msg("Starting CasparCG AMCP Proxy")

	// Initialize Google Workspace handler
	gw, err := googleworkspace.NewHandler(log.With().Str("component", "googleWorkspace").Logger())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Google Workspace handler")
	}
	gw.Start()
	defer gw.Close()

	// Initialize and start the CasparCG AMCP Proxy
	proxy, err := cp.NewProxy(proxyPort, casparCGHost, casparCGPort, gw, log.With().Str("component", "casperProxy").Logger())
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create proxy")
	}
	if err := proxy.Start(); err != nil {
		log.Fatal().Err(err).Msgf("Error starting proxy")
	}

	// Wait for interrupt signal (Ctrl+C) to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Info().Msg("Received interrupt signal, shutting down")
}
