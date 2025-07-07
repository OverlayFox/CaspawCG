package casperproxy

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/OverlayFox/CaspawCG/casperProxy/types"
	gTypes "github.com/OverlayFox/CaspawCG/types"
)

type Proxy struct {
	proxyPort  string // Port the proxy listens on for clients
	serverConn net.Conn
	sheetsData gTypes.SheetsData
	mu         sync.RWMutex // Protects server connection
}

// NewProxy creates a new CasparCG AMCP Proxy instance
func NewProxy(proxyPort, casparCGHost, casparCGPort string, sheetsData gTypes.SheetsData) (*Proxy, error) {
	serverAddr := fmt.Sprintf("%s:%s", casparCGHost, casparCGPort)
	serverConn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to CasparCG server at %s: %w", serverAddr, err)
	}

	return &Proxy{
		proxyPort:  proxyPort,
		serverConn: serverConn,
		sheetsData: sheetsData,
	}, nil
}

// Start begins listening for client connections and handles them
func (p *Proxy) Start() error {
	addr := fmt.Sprintf(":%s", p.proxyPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", p.proxyPort, err)
	}
	defer listener.Close()

	log.Printf("CasparCG AMCP Proxy listening on port %s, forwarding to %s",
		p.proxyPort, p.serverConn.RemoteAddr().String())

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting client connection: %v", err)
			continue
		}

		go p.handleClient(clientConn)
	}
}

// handleClient manages a single client connection
func (p *Proxy) handleClient(clientConn net.Conn) {
	defer clientConn.Close()

	clientAddr := clientConn.RemoteAddr().String()
	log.Printf("Client connected from %s", clientAddr)
	defer log.Printf("Client %s disconnected", clientAddr)

	go p.forwardServerResponses(clientConn)

	if err := p.processClientCommands(clientConn); err != nil {
		if err != io.EOF {
			log.Printf("Error processing commands from client %s: %v", clientAddr, err)
		}
	}
}

// forwardServerResponses forwards all server responses to the client
func (p *Proxy) forwardServerResponses(clientConn net.Conn) {
	p.mu.RLock()
	serverConn := p.serverConn
	p.mu.RUnlock()

	if serverConn == nil {
		log.Printf("No server connection available for client %s", clientConn.RemoteAddr())
		return
	}

	_, err := io.Copy(clientConn, serverConn)
	if err != nil {
		log.Printf("Error forwarding server response to client %s: %v",
			clientConn.RemoteAddr(), err)
	}
}

// processClientCommands reads and processes commands from the client
func (p *Proxy) processClientCommands(clientConn net.Conn) error {
	reader := bufio.NewReader(clientConn)

	for {
		// Read AMCP command (terminated by \r\n)
		message, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		commandStr := strings.TrimSpace(message)
		if commandStr == "" {
			continue
		}

		log.Printf("Received from client %s: %s", clientConn.RemoteAddr(), commandStr)

		newCommandStr, err := p.interceptCommand(commandStr)
		if err != nil {
			log.Printf("Error intercepting command '%s': %v", commandStr, err)
		} else {
			commandStr = newCommandStr
		}

		if err := p.forwardToServer(commandStr); err != nil {
			return fmt.Errorf("failed to forward command to server: %w", err)
		}
	}
}

// interceptCommand parses and processes supported commands
func (p *Proxy) interceptCommand(commandStr string) (string, error) {
	command, err := types.NewCommand(commandStr)
	if err != nil {
		if errors.Is(err, types.ErrUnsupportedCommandType) {
			// Not an error - just an unsupported command type
			return commandStr, nil
		}
		return commandStr, fmt.Errorf("failed to parse command: %w", err)
	}

	switch command.GetCommandType() {
	case types.CommandTypeCG:
		return p.handleCGCommand(command, commandStr)
	default:
		// Other command types can be added here
		return commandStr, nil
	}
}

// handleCGCommand processes CG (Character Generator) commands
func (p *Proxy) handleCGCommand(command types.Command, originalCommand string) (string, error) {
	cgCommand, ok := command.(*types.CommandCG)
	if !ok {
		return originalCommand, fmt.Errorf("expected CG command but got %T", command)
	}

	// Only process ADD calls
	if cgCommand.Call == nil || *cgCommand.Call != types.CommandCallADD {
		return originalCommand, nil
	}

	if cgCommand.TemplatePath == nil {
		log.Printf("CG command missing template path: %s", originalCommand)
		return originalCommand, nil
	}

	return p.processCGTemplate(cgCommand, originalCommand)
}

// processCGTemplate handles different CG template types
func (p *Proxy) processCGTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	switch *cgCommand.TemplatePath {
	case types.TemplatePathCountdown:
		return p.processCountdownTemplate(cgCommand, originalCommand)
	case types.TemplatePathTitle:
		return p.processTitleTemplate(cgCommand, originalCommand)
	case types.TemplatePathBar:
		return p.processBarTemplate(cgCommand, originalCommand)
	default:
		log.Printf("Unsupported CG template path: %s", *cgCommand.TemplatePath)
		return originalCommand, nil
	}
}

// processCountdownTemplate handles countdown template commands
func (p *Proxy) processCountdownTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	countdown, ok := cgCommand.JsonData.(*types.Countdown)
	if !ok {
		return originalCommand, fmt.Errorf("invalid JSON data for Countdown template: %s", originalCommand)
	}

	countdownData := p.sheetsData.GetCountdown()

	countdown.Title = countdownData.Title
	countdown.TimerHours = countdownData.CountdownTime

	log.Printf("Processing Countdown: Title=%s, TimerMinutes=%s, TimerSeconds=%s",
		countdown.Title, countdown.TimerMinutes, countdown.TimerHours)

	return countdown.Command()
}

// processTitleTemplate handles title template commands
func (p *Proxy) processTitleTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	title, ok := cgCommand.JsonData.(*types.Title)
	if !ok {
		return originalCommand, fmt.Errorf("invalid JSON data for Title template: %s", originalCommand)
	}

	log.Printf("Processing Title: Title=%s", title.Title)

	// Additional title processing logic can be added here
	return title.Command()
}

// processBarTemplate handles bar template commands
func (p *Proxy) processBarTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	bar, ok := cgCommand.JsonData.(*types.Bar)
	if !ok {
		return originalCommand, fmt.Errorf("invalid JSON data for Bar template: %s", originalCommand)
	}

	log.Printf("Processing Bar: Number=%s, Title=%s", bar.Number, bar.Title)

	// Additional bar processing logic can be added here
	return bar.Command()
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
	_, err := serverConn.Write([]byte(command + "\r\n"))
	if err != nil {
		return fmt.Errorf("failed to write to server: %w", err)
	}

	return nil
}

// Close cleanly shuts down the proxy
func (p *Proxy) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.serverConn != nil {
		err := p.serverConn.Close()
		p.serverConn = nil
		return err
	}

	return nil
}
