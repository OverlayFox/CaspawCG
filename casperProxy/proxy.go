package casperproxy

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/OverlayFox/CaspawCG/casperProxy/types"
	gTypes "github.com/OverlayFox/CaspawCG/types"
	"github.com/rs/zerolog"
)

const (
	// Lower third layer configuration
	lowerThirdSingleLayer = 20
	lowerThirdDuoLayer1   = 21
	lowerThirdDuoLayer2   = 22

	// Network timeouts
	connectionTimeout = 30 * time.Second
)

// Proxy represents a CasparCG AMCP proxy server
type Proxy struct {
	logger zerolog.Logger

	proxyPort  string
	serverAddr string
	serverConn net.Conn

	updateCh chan *types.CommandCG

	sheetsData      gTypes.SheetsData
	scheduleHandler *ScheduleHandler

	wg     sync.WaitGroup
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewProxy creates a new CasparCG AMCP Proxy instance
func NewProxy(proxyPort, casparCGHost, casparCGPort string, sheetsData gTypes.SheetsData, logger zerolog.Logger) (*Proxy, error) {
	if proxyPort == "" || casparCGHost == "" || casparCGPort == "" {
		return nil, errors.New("proxy port, CasparCG host, and port must be specified")
	}

	serverAddr := fmt.Sprintf("%s:%s", casparCGHost, casparCGPort)
	serverConn, err := net.DialTimeout("tcp", serverAddr, connectionTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to CasparCG server at %s: %w", serverAddr, err)
	}
	logger.Debug().Str("server_address", serverAddr).Msg("Successfully connected to the CasparCG server")

	ctx, cancel := context.WithCancel(context.Background())

	return &Proxy{
		logger:     logger,
		proxyPort:  proxyPort,
		serverAddr: serverAddr,
		serverConn: serverConn,
		updateCh:   make(chan *types.CommandCG, 1),
		sheetsData: sheetsData,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Start begins listening for client connections
func (p *Proxy) Start() error {
	p.logger.Info().Str("listen_port", p.proxyPort).Str("server_address", p.serverAddr).Msg("Starting CasparCG AMCP Proxy listener.")
	listener, err := net.Listen("tcp", ":"+p.proxyPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", p.proxyPort, err)
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer listener.Close()

		for {
			select {
			case <-p.ctx.Done():
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					p.logger.Error().Err(err).Msg("Error accepting client connection")
					continue
				}
				go p.handleClient(conn)
			}
		}
	}()

	p.wg.Add(1)
	go p.updateCommandProcessor()

	return nil
}

// handleClient manages a single client connection
func (p *Proxy) handleClient(clientConn net.Conn) {
	defer clientConn.Close()

	clientAddr := clientConn.RemoteAddr().String()
	p.logger.Info().Str("client_address", clientAddr).Msg("New client connection")
	defer p.logger.Info().Str("client_address", clientAddr).Msg("Client connection closed")

	go p.forwardServerResponses(clientConn)

	if err := p.processClientCommands(clientConn); err != nil && err != io.EOF {
		p.logger.Error().Err(err).Str("client_address", clientAddr).Msg("Error processing commands from client")
	}
}

// forwardServerResponses forwards all server responses to the client
func (p *Proxy) forwardServerResponses(clientConn net.Conn) {
	p.mu.RLock()
	serverConn := p.serverConn
	p.mu.RUnlock()

	if serverConn == nil {
		p.logger.Error().Str("client_address", clientConn.RemoteAddr().String()).Msg("No server connection available for client to forward responses")
		return
	}

	if _, err := io.Copy(clientConn, serverConn); err != nil {
		p.logger.Error().Err(err).Str("client_address", clientConn.RemoteAddr().String()).Msg("Error forwarding server response to client")
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

			p.logger.Debug().Str("client_address", clientConn.RemoteAddr().String()).Str("command", commandStr).Msg("Received AMCP command from client")

			processedCommand, err := p.interceptCommand(commandStr)
			if err != nil {
				if !errors.Is(err, ErrNotCGAddCall) {
					p.logger.Error().Err(err).Str("command", commandStr).Msg("Error intercepting command")
				}
				processedCommand = commandStr // Use original command on error
			}

			p.logger.Debug().Str("client_address", clientConn.RemoteAddr().String()).Str("processed_command", processedCommand).Msg("Processed command from client, forwarding it to server")

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

func (p *Proxy) updateCommandProcessor() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		case command, ok := <-p.updateCh:
			if !ok {
				return
			}

			p.logger.Debug().Msg("Processing schedule update command")

			p.mu.RLock()
			serverConn := p.serverConn
			p.mu.RUnlock()

			if serverConn == nil {
				return
			}

			cmd, err := command.BuildCommand()
			if err != nil {
				p.logger.Error().Err(err).Msg("Failed to build command for schedule update")
				return
			}

			p.logger.Debug().Str("processed_command", cmd).Msg("Sending update command to Server.")

			// AMCP protocol requires \r\n line endings
			if _, err := serverConn.Write([]byte(cmd + "\r\n")); err != nil {
				return
			}

		}
	}
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
