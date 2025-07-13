package casperproxy

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/OverlayFox/CaspawCG/casperProxy/types"
	gTypes "github.com/OverlayFox/CaspawCG/types"
)

// handleCGCommand processes CG (Character Generator) commands
func (p *Proxy) handleCGCommand(command types.Command, originalCommand string) (string, error) {
	cgCommand, ok := command.(*types.CommandCG)
	if !ok {
		return "", fmt.Errorf("expected CG command but got %T", command)
	}

	// Only process ADD calls
	if cgCommand.Call == nil || *cgCommand.Call != types.CommandCallADD {
		return "", fmt.Errorf("CG command is not an ADD call: %s", *cgCommand.Call)
	}

	if cgCommand.TemplatePath == nil {
		return "", fmt.Errorf("CG command missing template path: %s", originalCommand)
	}

	return p.processCGTemplate(cgCommand, originalCommand)
}

// processCGTemplate handles different CG template types
func (p *Proxy) processCGTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	templateProcessors := map[types.TemplatePaths]func(*types.CommandCG, string) (string, error){
		types.TemplatePathCountdown:       p.processCountdownTemplate,
		types.TemplatePathCountdownToTime: p.processCountdownToTimeTemplate,
		types.TemplatePathTitle:           p.passThroughTemplate,
		types.TemplatePathBarRed:          p.processBarTemplate,
		types.TemplatePathBarBlue:         p.processBarTemplate,
		types.TemplatePathSchedule:        p.processScheduleTemplate,
		types.TemplatePathDanceComp:       p.processDetailedDanceCompTemplate,
		types.TemplatePathLowerThird:      p.processLowerThirdTemplate,
	}

	processor, exists := templateProcessors[*cgCommand.TemplatePath]
	if !exists {
		return "", fmt.Errorf("unsupported CG template path: %s", *cgCommand.TemplatePath)
	}

	return processor(cgCommand, originalCommand)
}

// passThroughTemplate returns the original command unchanged
func (p *Proxy) passThroughTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	return originalCommand, nil
}

//
// Lower Third
//

// processLowerThirdTemplate handles lower third template commands
func (p *Proxy) processLowerThirdTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	lowerThird, ok := cgCommand.JsonData.(*types.LowerThird)
	if !ok {
		return "", fmt.Errorf("invalid JSON data for Lower Third template: %s", originalCommand)
	}

	lowerThirdData, err := p.getLowerThirdData(*cgCommand.Layer)
	if err != nil {
		return "", fmt.Errorf("failed to get lower third data: %w", err)
	}

	lowerThird.Name = lowerThirdData.Row1
	lowerThird.Info = lowerThirdData.Row2

	return cgCommand.Command()
}

// getLowerThirdData retrieves lower third data based on layer
func (p *Proxy) getLowerThirdData(layer int) (*gTypes.LowerThird, error) {
	switch layer {
	case lowerThirdSingleLayer:
		return p.sheetsData.GetLowerThirdSingle(), nil
	case lowerThirdDuoLayer1:
		data, _ := p.sheetsData.GetLowerThirdDuo()
		return data, nil
	case lowerThirdDuoLayer2:
		_, data := p.sheetsData.GetLowerThirdDuo()
		return data, nil
	default:
		return nil, fmt.Errorf("unsupported layer %d for lower third template", layer)
	}
}

//
// Detailed Dance Competition
//

// processDetailedDanceCompTemplate handles detailed dance competition template commands
func (p *Proxy) processDetailedDanceCompTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	danceComp, ok := cgCommand.JsonData.(*types.DetailedDanceComp)
	if !ok {
		return "", fmt.Errorf("invalid JSON data for Detailed Dance Competition template: %s", originalCommand)
	}

	finalist := p.sheetsData.GetDetailedDanceCompSingle()
	if finalist == nil {
		return "", fmt.Errorf("no detailed dance competition data available")
	}

	p.populateDanceCompData(danceComp, finalist)

	if picturePath := p.getDanceCompPicturePath(finalist.Name); picturePath != "" {
		danceComp.PicturePath = picturePath
	}

	return cgCommand.Command()
}

// populateDanceCompData fills the dance competition data structure
func (p *Proxy) populateDanceCompData(danceComp *types.DetailedDanceComp, finalist *gTypes.DetailedDanceComp) {
	danceComp.Name = finalist.Name
	danceComp.TotalScore = finalist.TotalScore
	danceComp.AppearanceScore = finalist.Appearance
	danceComp.ProfessionalismScore = finalist.Professionalism
	danceComp.ConsistencyScore = finalist.Consistency
	danceComp.ComplexityScore = finalist.Complexity
	danceComp.DecibelsScore = finalist.Decibels
	danceComp.OriginalityScore = finalist.Originality
	danceComp.QuantumScore = finalist.Quantum
	danceComp.Attribution = p.getPictureAttribution(finalist.Name)
}

// getDanceCompPicturePath generates and validates the picture path for a contestant
func (p *Proxy) getDanceCompPicturePath(name string) string {
	cleanName := strings.NewReplacer(" ", "_", "&", "_").Replace(name)
	pictureFileName := "danceComp/contestant_" + strings.ToLower(cleanName) + ".png"

	absPicturePath, err := filepath.Abs("../casparCG/template/images/" + pictureFileName)
	if err != nil {
		p.logger.Warn().Str("picture_file", pictureFileName).Msg("Error resolving absolute path for picture file")
		return ""
	}

	if _, err := os.Stat(absPicturePath); err == nil {
		return pictureFileName
	} else if !os.IsNotExist(err) {
		p.logger.Error().Err(err).Str("picture_file", absPicturePath).Msgf("Error checking picture files existence")
	}

	return ""
}

func (p *Proxy) getPictureAttribution(name string) string {
	photographer, err := p.sheetsData.GetAttribution(name)
	if err != nil {
		p.logger.Error().Err(err).Str("contestant", name).Msg("Error retrieving picture attribution for dance contestants picture")
		return ""
	}

	if photographer != "" {
		return fmt.Sprintf("Picture by: @%s", photographer)
	}

	return ""
}

//
// Schedule
//

// processScheduleTemplate handles schedule template commands
func (p *Proxy) processScheduleTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	bar, ok := cgCommand.JsonData.(*types.ScheduleBar)
	if !ok {
		return "", fmt.Errorf("invalid JSON data for Schedule Bar template: %s", originalCommand)
	}

	scheduleData, err := p.getScheduleBarData(*cgCommand.Layer)
	if err != nil {
		return "", fmt.Errorf("failed to get schedule data: %w", err)
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

// getScheduleBarData retrieves and formats schedule data for a specific layer
func (p *Proxy) getScheduleBarData(layer int) (*types.ScheduleBar, error) {
	if layer < scheduleBarStartLayer || layer > scheduleBarEndLayer {
		return nil, fmt.Errorf("invalid layer '%d' for schedule bar template", layer)
	}

	schedule := p.sheetsData.GetCurrentSchedule()
	scheduleIndex := layer - scheduleBarStartLayer

	if len(schedule) <= scheduleIndex {
		return nil, fmt.Errorf("not enough schedule rows for layer '%d'", layer)
	}

	scheduleRow := schedule[scheduleIndex]
	bar := &types.ScheduleBar{
		StartTime: scheduleRow.StartTime.Format("15:04"),
		EndTime:   scheduleRow.EndTime.Format("15:04"),
		Hotel:     scheduleRow.Hotel,
		Room:      scheduleRow.Room,
	}

	p.formatScheduleTitle(bar, scheduleRow.Title)

	return bar, nil
}

// formatScheduleTitle splits the title across multiple rows based on content and length
func (p *Proxy) formatScheduleTitle(bar *types.ScheduleBar, title string) {
	if strings.Contains(title, "-") {
		// Split on dash
		parts := strings.SplitN(title, "-", 2)
		bar.Row1 = ""
		bar.Row2 = strings.TrimSpace(parts[0])
		bar.Row3 = strings.TrimSpace(parts[1])
		return
	}

	if len(title) <= scheduleBarMaxChars {
		// Short title goes on row1
		bar.Row1 = title
		bar.Row2 = ""
		bar.Row3 = ""
		return
	}

	// Long title: split at last space before max chars
	splitIdx := p.findSplitIndex(title, scheduleBarMaxChars)
	bar.Row1 = ""
	bar.Row2 = strings.TrimSpace(title[:splitIdx])
	bar.Row3 = strings.TrimSpace(title[splitIdx:])
}

// findSplitIndex finds the best position to split a string at or before maxIndex
func (p *Proxy) findSplitIndex(text string, maxIndex int) int {
	if len(text) <= maxIndex {
		return len(text)
	}

	// Look for last space before maxIndex
	for i := maxIndex; i >= 0; i-- {
		if i < len(text) && text[i] == ' ' {
			return i
		}
	}

	// No space found, split at maxIndex
	return maxIndex
}

//
// Countdowns
//

// processCountdownTemplate handles countdown template commands
func (p *Proxy) processCountdownTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	countdown, ok := cgCommand.JsonData.(*types.Countdown)
	if !ok {
		return "", fmt.Errorf("invalid JSON data for Countdown template: %s", originalCommand)
	}

	countdownData := p.sheetsData.GetCountdown()
	countdown.Title = countdownData.Title
	countdown.TimerHours = countdownData.CountdownTime

	return cgCommand.Command()
}

// processCountdownToTimeTemplate handles countdown template commands
func (p *Proxy) processCountdownToTimeTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	countdown, ok := cgCommand.JsonData.(*types.Countdown)
	if !ok {
		return "", fmt.Errorf("invalid JSON data for Countdown template: %s", originalCommand)
	}

	countdownData := p.sheetsData.GetCountdownToTime()
	countdown.Title = countdownData.Title
	countdown.TimerHours = countdownData.CountdownTime

	log.Printf("Processing Countdown to time: Title=%s,  TimerHours=%s TimerMinutes=%s", countdown.Title, countdown.TimerHours, countdown.TimerMinutes)

	return cgCommand.Command()
}

//
// Bars
//

// processBarTemplate handles bar template commands
func (p *Proxy) processBarTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	bar, ok := cgCommand.JsonData.(*types.Bar)
	if !ok {
		return "", fmt.Errorf("invalid JSON data for Bar template: %s", originalCommand)
	}

	barData, err := p.getBarDanceCompData(*cgCommand.Layer)
	if err != nil {
		return originalCommand, fmt.Errorf("failed to get bar data: %w", err)
	}

	bar.Number = barData.Number
	bar.Title = barData.Title

	return cgCommand.Command()
}

// getBarDanceCompData retrieves bar data for dance competition standings
func (p *Proxy) getBarDanceCompData(layer int) (*types.Bar, error) {
	if layer < barTemplateStartLayer || layer > barTemplateEndLayer {
		return nil, fmt.Errorf("invalid layer %d for bar template", layer)
	}

	standings := p.sheetsData.GetDetailedDanceComp()
	standingIndex, exists := layerMapping[layer]
	if !exists {
		return nil, fmt.Errorf("invalid layer %d for bar template", layer)
	}

	if len(standings) <= standingIndex {
		return nil, fmt.Errorf("not enough dance competition standings for layer %d", layer)
	}

	standing := standings[standingIndex]
	return &types.Bar{
		Number: standing.TotalScore,
		Title:  standing.Name,
	}, nil
}
