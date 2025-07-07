package casperproxy

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"strings"

	"github.com/OverlayFox/CaspawCG/casperProxy/types"
)

type Proxy struct {
	ProxyPort string // Port your proxy will listen on for clients

	serverConn net.Conn
}

// NewProxy initializes a new CasparCG AMCP Proxy instance.
//
// ProxyPort is the port the proxy will listen on for incoming client connections.
// casparCGHost is the IP address of the actual CasparCG Server.
// casparCGPort is the port of the actual CasparCG Server.
func NewProxy(proxyPort, casparCGHost, casparCGPort string) (*Proxy, error) {
	serverConn, err := net.Dial("tcp", casparCGHost+":"+casparCGPort)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		ProxyPort:  proxyPort,
		serverConn: serverConn,
	}, nil
}

func (p *Proxy) Start() error {
	ln, err := net.Listen("tcp", ":"+p.ProxyPort)
	if err != nil {
		return err
	}
	defer ln.Close()

	log.Printf("CasparCG AMCP Proxy listening on port %s, forwarding to %s", p.ProxyPort, p.serverConn.RemoteAddr().String())

	for {
		clientConn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting client connection: %v", err)
			continue
		}
		go p.handleClient(clientConn)
	}
}

func (p *Proxy) handleClient(clientConn net.Conn) {
	defer clientConn.Close()
	log.Printf("Client connected from %s", clientConn.RemoteAddr())

	// Goroutine to forward server responses to client
	go func() {
		_, err := io.Copy(clientConn, p.serverConn)
		if err != nil {
			log.Printf("Error forwarding server response to client %s: %v", clientConn.RemoteAddr(), err)
		}
	}()

	reader := bufio.NewReader(clientConn)
	err := p.interceptCommands(reader)
	if err != nil {
		if err == io.EOF {
			log.Printf("Client %s disconnected.", clientConn.RemoteAddr())
		} else {
			log.Printf("Error reading from client %s: %v", clientConn.RemoteAddr(), err)
		}
	}
}

func (p *Proxy) interceptCommands(reader *bufio.Reader) error {
	for {
		message, err := reader.ReadString('\n') // AMCP commands are terminated by \r\n
		if err != nil {
			return err
		}
		commandStr := strings.TrimSpace(message)
		command, err := types.NewCommand(commandStr)
		if err != nil {
			if !errors.Is(err, types.ErrUnsupportedCommandType) {
				log.Printf("Error parsing command from client: '%s', error: %v", commandStr, err)
			}
			err = p.writeToServer(commandStr)
			if err != nil {
				return err
			}
			continue
		}
		log.Printf("Received from client: %s", commandStr)

		switch command.GetCommandType() {
		case types.CommandTypeCG:
			cgCommand, ok := command.(*types.CommandCG)
			if !ok {
				log.Printf("Received non-CG command from client: %s", commandStr)
				break
			}

			if *cgCommand.Call != types.CommandCallADD {
				break
			}

			if cgCommand.TemplatePath == nil {
				log.Printf("Received CG command without template path: %s", commandStr)
				break
			}

			switch *cgCommand.TemplatePath {
			case types.TemplatePathCountdown:
				countdown, ok := cgCommand.JsonData.(*types.Countdown)
				if !ok {
					log.Printf("Received CG command with invalid JSON data for Countdown: %s", commandStr)
					break
				}
				log.Printf("Received Countdown command: Title=%s, TimerMinutes=%s, TimerSeconds=%s",
					countdown.Title, countdown.TimerMinutes, countdown.TimerSeconds)
			case types.TemplatePathTitle:
				title, ok := cgCommand.JsonData.(*types.Title)
				if !ok {
					log.Printf("Received CG command with invalid JSON data for Title: %s", commandStr)
					break
				}
				log.Printf("Received Title command: Title=%s", title.Title)
			case types.TemplatePathBar:
				bar, ok := cgCommand.JsonData.(*types.Bar)
				if !ok {
					log.Printf("Received CG command with invalid JSON data for Bar: %s", commandStr)
					break
				}
				log.Printf("Received Bar command: Number=%s, Title=%s", bar.Number, bar.Title)
			default:
				log.Printf("Received CG command with unsupported template path: %s", *cgCommand.TemplatePath)
			}
		}

		err = p.writeToServer(commandStr)
		if err != nil {
			return err
		}
	}
}

func (p *Proxy) writeToServer(command string) (err error) {
	_, err = p.serverConn.Write([]byte(command + "\r\n")) // AMCP requires \r\n
	return err
}

func (p *Proxy) Close() {
	if p.serverConn != nil {
		p.serverConn.Close()
	}
}
