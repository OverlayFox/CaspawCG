package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	cp "github.com/OverlayFox/CaspawCG/casperProxy"
	googleworkspace "github.com/OverlayFox/CaspawCG/googleWorkspace"
)

const (
	proxyPort    = "5251"
	casparCGHost = "127.0.0.1"
	casparCGPort = "5250"
)

func main() {
	gw, err := googleworkspace.NewHandler()
	if err != nil {
		log.Fatalf("Failed to create Google Workspace handler: %v", err)
	}

	gw.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Println("Received interrupt signal, shutting down...")
	return

	proxy, err := cp.NewProxy(proxyPort, casparCGHost, casparCGPort, gw)
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}

	log.Printf("Starting CasparCG AMCP Proxy on port %s, forwarding to %s:%s", proxyPort, casparCGHost, casparCGPort)

	if err := proxy.Start(); err != nil {
		log.Fatalf("Error starting proxy: %v", err)
	}
}
