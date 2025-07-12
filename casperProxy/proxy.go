package casperproxy

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/OverlayFox/CaspawCG/casperProxy/types"
	gTypes "github.com/OverlayFox/CaspawCG/types"
)

const (
	scheduleBarStartLayer = 41 // First layer for schedule bars
	scheduleBarMaxChars   = 35 // Max characters per row in schedule bar template
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

		log.Printf("Forwarding command to server: %s", commandStr)

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

	log.Printf("Processing CG command: %s", originalCommand)

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
	case types.TemplatePathCountdownToTime:
		return p.processCountdownToTimeTemplate(cgCommand, originalCommand)
	case types.TemplatePathTitle:
		return originalCommand, nil
	case types.TemplatePathBarRed, types.TemplatePathBarBlue:
		return p.processBarTemplate(cgCommand, originalCommand)
	case types.TemplatePathSchedule:
		return p.processScheduleTemplate(cgCommand, originalCommand)
	case types.TemplatePathDanceComp:
		return p.processDetailedDanceCompTemplate(cgCommand, originalCommand)
	case types.TemplatePathLowerThird:
		return p.processLowerThirdTemplate(cgCommand, originalCommand)
	default:
		log.Printf("Unsupported CG template path: %s", *cgCommand.TemplatePath)
		return originalCommand, nil
	}
}

func (p *Proxy) processLowerThirdTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	lowerThird, ok := cgCommand.JsonData.(*types.LowerThird)
	if !ok {
		return originalCommand, fmt.Errorf("invalid JSON data for Lower Third template: %s", originalCommand)
	}

	layer := *cgCommand.Layer
	lowerThirdData := &gTypes.LowerThird{}
	if layer == 20 {
		lowerThirdData = p.sheetsData.GetLowerThirdSingle()
	} else if layer == 21 {
		lowerThirdData, _ = p.sheetsData.GetLowerThirdDuo()
	} else if layer == 22 {
		_, lowerThirdData = p.sheetsData.GetLowerThirdDuo()
	} else {
		return originalCommand, fmt.Errorf("unsupported layer %d for lower third template", layer)
	}

	if lowerThirdData == nil {
		return originalCommand, fmt.Errorf("no lower third data available")
	}

	lowerThird.Name = lowerThirdData.Row1
	lowerThird.Info = lowerThirdData.Row2

	return cgCommand.Command()
}

func (p *Proxy) processDetailedDanceCompTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	danceComp, ok := cgCommand.JsonData.(*types.DetailedDanceComp)
	if !ok {
		return originalCommand, fmt.Errorf("invalid JSON data for Detailed Dance Competition template: %s", originalCommand)
	}

	finalist := p.sheetsData.GetDetailedDanceCompSingle()
	if finalist == nil {
		return originalCommand, fmt.Errorf("no detailed dance competition data available")
	}

	danceComp.Name = finalist.Name
	danceComp.TotalScore = finalist.TotalScore
	danceComp.AppearanceScore = finalist.Appearance
	danceComp.ProfessionalismScore = finalist.Professionalism
	danceComp.ConsistencyScore = finalist.Consistency
	danceComp.ComplexityScore = finalist.Complexity
	danceComp.DecibelsScore = finalist.Decibels
	danceComp.OriginalityScore = finalist.Originality
	danceComp.QuantumScore = finalist.Quantum

	cleanName := strings.NewReplacer(" ", "_", "&", "_").Replace(finalist.Name)
	pictureFileName := "danceComp/contestant_" + strings.ToLower(cleanName) + ".png"

	absPicturePath, err := filepath.Abs("../casparCG/template/images/" + pictureFileName)
	if err != nil {
		log.Printf("Error resolving absolute path for %s: %v", pictureFileName, err)
	}
	if _, err := os.Stat(absPicturePath); err == nil {
		danceComp.PicturePath = pictureFileName
	} else if !os.IsNotExist(err) {
		log.Printf("Error checking picture file %s: %v", absPicturePath, err)
	}

	return cgCommand.Command()
}

func (p *Proxy) getScheduleBar(schedule []*gTypes.ScheduleRow, layer int) (*types.ScheduleBar, error) {
	if layer < 41 || layer > 44 {
		return nil, fmt.Errorf("invalid layer %d for schedule bar template", layer)
	}

	if len(schedule) < 3 {
		return nil, fmt.Errorf("not enough schedule rows for layer %d", layer)
	}

	bar := &types.ScheduleBar{
		StartTime: schedule[layer-scheduleBarStartLayer].StartTime.Format("15:04"),
		EndTime:   schedule[layer-scheduleBarStartLayer].EndTime.Format("15:04"),
		Hotel:     schedule[layer-scheduleBarStartLayer].Hotel,
		Room:      schedule[layer-scheduleBarStartLayer].Room,
	}

	title := schedule[layer-scheduleBarStartLayer].Title

	if strings.Contains(title, "-") {
		parts := strings.SplitN(title, "-", 2)
		bar.Row1 = ""
		bar.Row2 = strings.TrimSpace(parts[0])
		bar.Row3 = strings.TrimSpace(parts[1])
	} else if len(title) > scheduleBarMaxChars {
		// Find the last whitespace before or at the maxChar character
		splitIdx := scheduleBarMaxChars
		for i := scheduleBarMaxChars; i >= 0; i-- {
			if i < len(title) && title[i] == ' ' {
				splitIdx = i
				break
			}
		}
		// If no whitespace found, just split at maxChar
		if splitIdx == scheduleBarMaxChars {
			bar.Row1 = ""
			bar.Row2 = title[:scheduleBarMaxChars]
			bar.Row3 = title[scheduleBarMaxChars:]
		} else {
			bar.Row1 = ""
			bar.Row2 = strings.TrimSpace(title[:splitIdx])
			bar.Row3 = strings.TrimSpace(title[splitIdx:])
		}
	} else {
		bar.Row1 = title
		bar.Row2 = ""
		bar.Row3 = ""
	}

	return bar, nil
}

func (p *Proxy) processScheduleTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	bar, ok := cgCommand.JsonData.(*types.ScheduleBar)
	if !ok {
		return originalCommand, fmt.Errorf("invalid JSON data for Schedule Bar template: %s", originalCommand)
	}

	schedule := p.sheetsData.GetCurrentSchedule()
	scheduleData, err := p.getScheduleBar(schedule, *cgCommand.Layer)
	if err != nil {
		return originalCommand, fmt.Errorf("failed to get schedule data: %w", err)
	}

	bar.Row1 = scheduleData.Row1
	bar.Row2 = scheduleData.Row2
	bar.Row3 = scheduleData.Row3
	bar.StartTime = scheduleData.StartTime
	bar.EndTime = scheduleData.EndTime
	bar.Hotel = scheduleData.Hotel
	bar.Room = scheduleData.Room

	return cgCommand.Command()

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

	log.Printf("Processing Countdown: Title=%s, TimerMinutes=%s, TimerHours=%s",
		countdown.Title, countdown.TimerMinutes, countdown.TimerHours)

	return cgCommand.Command()
}

// processCountdownToTimeTemplate handles countdown template commands
func (p *Proxy) processCountdownToTimeTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	countdown, ok := cgCommand.JsonData.(*types.Countdown)
	if !ok {
		return originalCommand, fmt.Errorf("invalid JSON data for Countdown template: %s", originalCommand)
	}

	countdownData := p.sheetsData.GetCountdownToTime()

	countdown.Title = countdownData.Title
	countdown.TimerHours = countdownData.CountdownTime

	log.Printf("Processing Countdown to Time: Title=%s, TimerMinutes=%s, TimerHours=%s",
		countdown.Title, countdown.TimerMinutes, countdown.TimerHours)

	return cgCommand.Command()
}

func getBarDanceComp(standings []*gTypes.DetailedDanceComp, layer int) (*types.Bar, error) {
	mapping := map[int]int{
		51: 0, // First place
		52: 1, // Second place
		53: 2,
		54: 3,
		55: 4,
		56: 5,
		57: 6,
		58: 7,
		59: 8,
		60: 9,
		61: 10,
	}

	if layer < 51 || layer > 61 {
		return nil, fmt.Errorf("invalid layer %d for bar template", layer)
	}
	if len(standings) <= mapping[layer] {
		return nil, fmt.Errorf("not enough dance competition standings for layer %d", layer)
	}

	return &types.Bar{
		Number: standings[mapping[layer]].TotalScore,
		Title:  standings[mapping[layer]].Name,
	}, nil
}

// processBarTemplate handles bar template commands
func (p *Proxy) processBarTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	bar, ok := cgCommand.JsonData.(*types.Bar)
	if !ok {
		return originalCommand, fmt.Errorf("invalid JSON data for Bar template: %s", originalCommand)
	}

	danceCompStandings := p.sheetsData.GetDetailedDanceComp()
	barData, err := getBarDanceComp(danceCompStandings, *cgCommand.Layer)
	if err != nil {
		return originalCommand, fmt.Errorf("failed to get bar data: %w", err)
	}

	bar.Number = barData.Number
	bar.Title = barData.Title

	return cgCommand.Command()
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
