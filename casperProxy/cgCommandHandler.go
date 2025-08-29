package casperproxy

import (
	"errors"
	"fmt"

	"github.com/OverlayFox/CaspawCG/casperProxy/types"
	gTypes "github.com/OverlayFox/CaspawCG/types"
)

var (
	ErrNotCGAddCall = errors.New("CG command is not an ADD call")
)

// handleCGCommand processes CG (Character Generator) commands
func (p *Proxy) handleCGCommand(command types.Command, originalCommand string) (string, error) {
	cgCommand, ok := command.(*types.CommandCG)
	if !ok {
		return "", fmt.Errorf("expected CG command but got %T", command)
	}

	if cgCommand.Call == nil {
		return "", fmt.Errorf("%w: %s", ErrNotCGAddCall, *cgCommand.Call)
	}

	switch *cgCommand.Call {
	case types.CommandCallADD:
		if cgCommand.TemplatePath == nil {
			return "", fmt.Errorf("CG command missing template path: %s", originalCommand)
		}
		return p.processAddCall(cgCommand, originalCommand)
	case types.CommandCallUPDATE:
		// p.scheduleHandler.ForceUpdate()
		return "", nil
	case types.CommandCallSTOP:
		return p.processStopCall(cgCommand, originalCommand)
	default:
		return "", fmt.Errorf("%w: %s", ErrNotCGAddCall, *cgCommand.Call)
	}
}

// processCGTemplate handles different CG template types
func (p *Proxy) processAddCall(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	templateProcessors := map[types.TemplatePaths]func(*types.CommandCG, string) (string, error){
		types.TemplatePathCountdown:         p.processCountdownTemplate,
		types.TemplatePathCountdownToTime:   p.processCountdownToTimeTemplate,
		types.TemplatePathDJCountdown:       p.processDJCountdownTemplate,
		types.TemplatePathDJCountdownToTime: p.processDJCountdownToTimeTemplate,
		types.TemplatePathTitle:             p.passThroughTemplate,
		types.TemplatePathSchedule:          p.processScheduleTemplate,
		types.TemplatePathLowerThird:        p.processLowerThirdTemplate,
	}

	processor, exists := templateProcessors[*cgCommand.TemplatePath]
	if !exists {
		return "", fmt.Errorf("unsupported CG template path: %s", *cgCommand.TemplatePath)
	}

	return processor(cgCommand, originalCommand)
}

func (p *Proxy) processStopCall(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	if cgCommand.Layer == nil {
		return "", fmt.Errorf("CG command missing layer: %s", originalCommand)
	}

	if *cgCommand.Layer == 41 {
		p.scheduleHandler.Stop()
		p.scheduleHandler = nil
	}

	return originalCommand, nil
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

	return cgCommand.BuildCommand()
}

// getLowerThirdData retrieves lower third data based on layer
func (p *Proxy) getLowerThirdData(layer int) (*gTypes.LowerThird, error) {
	switch layer {
	case lowerThirdSingleLayer:
		return p.sheetsData.GetLowerThird01(), nil
	case lowerThirdDuoLayer1:
		return p.sheetsData.GetLowerThird02(), nil
	case lowerThirdDuoLayer2:
		return p.sheetsData.GetLowerThirdDJ(), nil
	default:
		return nil, fmt.Errorf("unsupported layer %d for lower third template", layer)
	}
}

//
// Schedule
//

// processScheduleTemplate handles schedule template commands
func (p *Proxy) processScheduleTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	p.logger.Debug().Msg("Processing schedule template command")
	if p.scheduleHandler == nil {
		p.scheduleHandler = NewScheduleHandler(p.ctx, p.updateCh, p.sheetsData)
	}
	return p.scheduleHandler.Start(cgCommand)
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

	countdownData := p.sheetsData.GetCountdownDuration()
	countdown.Title = countdownData.Text
	countdown.TimerHours = countdownData.CountdownTime
	countdown.TimerMinutes = "00:00"

	return cgCommand.BuildCommand()
}

// processCountdownToTimeTemplate handles countdown template commands
func (p *Proxy) processCountdownToTimeTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	countdown, ok := cgCommand.JsonData.(*types.Countdown)
	if !ok {
		return "", fmt.Errorf("invalid JSON data for Countdown template: %s", originalCommand)
	}

	countdownData := p.sheetsData.GetCountdown()
	countdown.Title = countdownData.Text
	countdown.TimerHours = countdownData.CountdownTime
	countdown.TimerMinutes = "00:00"

	return cgCommand.BuildCommand()
}

// processCountdownTemplate handles countdown template commands
func (p *Proxy) processDJCountdownTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	countdown, ok := cgCommand.JsonData.(*types.DJCountdown)
	if !ok {
		return "", fmt.Errorf("invalid JSON data for Countdown template: %s", originalCommand)
	}

	countdownData := p.sheetsData.GetDJCountdownDuration()
	djLogo, err := getDJLogo(countdownData.Name)
	if err != nil {
		return "", fmt.Errorf("failed to get DJ logo: %w", err)
	}

	countdown.Name = countdownData.Name
	countdown.Genre = countdownData.Genre
	countdown.Logo = djLogo
	countdown.TimerHours = countdownData.CountdownTime
	countdown.TimerMinutes = "00:00"

	return cgCommand.BuildCommand()
}

// processCountdownToTimeTemplate handles countdown template commands
func (p *Proxy) processDJCountdownToTimeTemplate(cgCommand *types.CommandCG, originalCommand string) (string, error) {
	countdown, ok := cgCommand.JsonData.(*types.DJCountdown)
	if !ok {
		return "", fmt.Errorf("invalid JSON data for Countdown template: %s", originalCommand)
	}

	countdownData := p.sheetsData.GetDJCountdown()
	djLogo, err := getDJLogo(countdownData.Name)
	if err != nil {
		return "", fmt.Errorf("failed to get DJ logo: %w", err)
	}

	countdown.Name = countdownData.Name
	countdown.Genre = countdownData.Genre
	countdown.Logo = djLogo
	countdown.TimerHours = countdownData.CountdownTime
	countdown.TimerMinutes = "00:00"

	return cgCommand.BuildCommand()
}
