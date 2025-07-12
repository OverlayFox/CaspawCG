package casperproxy

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/OverlayFox/CaspawCG/casperProxy/types"
	gTypes "github.com/OverlayFox/CaspawCG/types"
)

const (
	// Schedule bar layer configuration
	scheduleBarStartLayer = 41
	scheduleBarEndLayer   = 44
	scheduleBarMaxChars   = 35

	// Lower third layer configuration
	lowerThirdSingleLayer = 20
	lowerThirdDuoLayer1   = 21
	lowerThirdDuoLayer2   = 22

	// Bar template layer configuration
	barTemplateStartLayer = 51
	barTemplateEndLayer   = 61

	// Network timeouts
	connectionTimeout = 30 * time.Second
	readTimeout       = 5 * time.Minute
)

// layerMapping maps layers to array indices for bar templates
var layerMapping = map[int]int{
	51: 0, 52: 1, 53: 2, 54: 3, 55: 4,
	56: 5, 57: 6, 58: 7, 59: 8, 60: 9, 61: 10,
}

// Proxy represents a CasparCG AMCP proxy server
type Proxy struct {
	proxyPort  string
	serverAddr string
	serverConn net.Conn
	sheetsData gTypes.SheetsData
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewProxy creates a new CasparCG AMCP Proxy instance
func NewProxy(proxyPort, casparCGHost, casparCGPort string, sheetsData gTypes.SheetsData) (*Proxy, error) {
	if proxyPort == "" || casparCGHost == "" || casparCGPort == "" {
		return nil, errors.New("proxy port, CasparCG host, and port must be specified")
	}

	serverAddr := fmt.Sprintf("%s:%s", casparCGHost, casparCGPort)
	serverConn, err := net.DialTimeout("tcp", serverAddr, connectionTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to CasparCG server at %s: %w", serverAddr, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Proxy{
		proxyPort:  proxyPort,
		serverAddr: serverAddr,
		serverConn: serverConn,
		sheetsData: sheetsData,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Start begins listening for client connections
func (p *Proxy) Start() error {
	listener, err := net.Listen("tcp", ":"+p.proxyPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", p.proxyPort, err)
	}
	defer listener.Close()

	log.Printf("CasparCG AMCP Proxy listening on port '%s', forwarding to '%s'", p.proxyPort, p.serverAddr)

	for {
		select {
		case <-p.ctx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting client connection: %v", err)
				continue
			}
			go p.handleClient(conn)
		}
	}
}

// handleClient manages a single client connection
func (p *Proxy) handleClient(clientConn net.Conn) {
	defer clientConn.Close()

	clientAddr := clientConn.RemoteAddr().String()
	log.Printf("Client connected from '%s'", clientAddr)
	defer log.Printf("Client '%s' disconnected", clientAddr)

	if err := clientConn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		log.Printf("Failed to set read timeout for client '%s': %v", clientAddr, err)
	}

	go p.forwardServerResponses(clientConn)

	if err := p.processClientCommands(clientConn); err != nil && err != io.EOF {
		log.Printf("Error processing commands from client '%s': %v", clientAddr, err)
	}
}

// forwardServerResponses forwards all server responses to the client
func (p *Proxy) forwardServerResponses(clientConn net.Conn) {
	p.mu.RLock()
	serverConn := p.serverConn
	p.mu.RUnlock()

	if serverConn == nil {
		log.Printf("No server connection available for client '%s'", clientConn.RemoteAddr())
		return
	}

	if _, err := io.Copy(clientConn, serverConn); err != nil {
		log.Printf("Error forwarding server response to client %s: %v",
			clientConn.RemoteAddr(), err)
	}
}

// processClientCommands reads and processes commands from the client
func (p *Proxy) processClientCommands(clientConn net.Conn) error {
	reader := bufio.NewReader(clientConn)

	for {
		select {
		case <-p.ctx.Done():
			return nil
		default:
			message, err := reader.ReadString('\n') // Read AMCP command (terminated by \r\n)
			if err != nil {
				return err
			}

			commandStr := strings.TrimSpace(message)
			if commandStr == "" {
				continue
			}

			log.Printf("Received from client '%s': %s", clientConn.RemoteAddr(), commandStr)

			processedCommand, err := p.interceptCommand(commandStr)
			if err != nil {
				log.Printf("Error intercepting command '%s': %v", commandStr, err)
				processedCommand = commandStr // Use original command on error
			}

			log.Printf("Forwarding command to server: %s", processedCommand)

			if err := p.forwardToServer(processedCommand); err != nil {
				return fmt.Errorf("failed to forward command to server: %w", err)
			}
		}
	}
}

// interceptCommand parses and processes supported commands
func (p *Proxy) interceptCommand(commandStr string) (string, error) {
	command, err := types.NewCommandFromString(commandStr)
	if err != nil {
		if errors.Is(err, types.ErrUnsupportedCommandType) {
			return commandStr, nil // Not an error, just unsupported
		}
		return commandStr, fmt.Errorf("failed to parse command: %w", err)
	}

	if command.GetCommandType() == types.CommandTypeCG {
		return p.handleCGCommand(command, commandStr)
	}

	return commandStr, nil
}

// forwardToServer sends a command to the CasparCG server
func (p *Proxy) forwardToServer(command string) error {
	p.mu.RLock()
	serverConn := p.serverConn
	p.mu.RUnlock()

	if serverConn == nil {
		return errors.New("no server connection available")
	}

	// AMCP protocol requires \r\n line endings
	if _, err := serverConn.Write([]byte(command + "\r\n")); err != nil {
		return fmt.Errorf("failed to write to server: %w", err)
	}

	return nil
}

// Close cleanly shuts down the proxy
func (p *Proxy) Close() error {
	p.cancel()

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.serverConn != nil {
		err := p.serverConn.Close()
		p.serverConn = nil
		return err
	}

	return nil
}
