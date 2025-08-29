package casperproxy

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/OverlayFox/CaspawCG/casperProxy/types"
	gTypes "github.com/OverlayFox/CaspawCG/types"
)

type ScheduleHandler struct {
	sheetsHandler    gTypes.SheetsData
	updateCmdChannel chan *types.CommandCG

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewScheduleHandler(upstreamCtx context.Context, updateCmdChannel chan *types.CommandCG, sheetsHandler gTypes.SheetsData) *ScheduleHandler {
	ctx, cancel := context.WithCancel(upstreamCtx)
	return &ScheduleHandler{
		sheetsHandler:    sheetsHandler,
		ctx:              ctx,
		cancel:           cancel,
		updateCmdChannel: updateCmdChannel,
	}
}

func (h *ScheduleHandler) Start(command *types.CommandCG) (string, error) {
	schedule, ok := command.JsonData.(*types.Schedule)
	if !ok {
		return "", fmt.Errorf("invalid JSON data for Schedule template: '%s'", *command.Call)
	}

	if err := populateScheduleCommand(0, schedule, h.sheetsHandler.GetCurrentSchedule()); err != nil {
		return "", err
	}

	h.wg.Add(1)
	go h.update(command)

	return command.BuildCommand()
}

func (h *ScheduleHandler) update(cmd *types.CommandCG) {
	defer h.wg.Done()

	currentIndex := 0
	jsonScheduleData, ok := cmd.JsonData.(*types.Schedule)
	if !ok {
		fmt.Printf("Error: failed to cast JsonData to Schedule")
		return
	}

	builder := types.UpdateCommandBuilder{
		Channel:  *cmd.Channel,
		Layer:    *cmd.Layer,
		CGLayer:  *cmd.CgLayer,
		JsonData: jsonScheduleData,
	}
	updateScheduleCommand := types.NewUpdateCommand(builder)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	fmt.Printf("Started schedule updates for channel %d\n", *cmd.Channel)
	for {
		select {
		case <-h.ctx.Done():
			fmt.Printf("Stopping schedule updates for channel %d\n", *cmd.Channel)
			return
		case <-ticker.C:
			fmt.Printf("Updating schedule for channel %d\n", *cmd.Channel)
			schedule := h.sheetsHandler.GetCurrentSchedule()
			currentIndex = (currentIndex + 1) % len(schedule)
			populateScheduleCommand(currentIndex, jsonScheduleData, schedule)
			select {
			case h.updateCmdChannel <- updateScheduleCommand:
			case <-h.ctx.Done():
				fmt.Printf("Stopping schedule updates for channel %d\n", *cmd.Channel)
				return
			default:
				fmt.Printf("Warning: update channel full, dropping schedule update for channel %d\n", *cmd.Channel)
				// If the upstream channel is full, we drop the update
			}
		}
	}
}

func (h *ScheduleHandler) Stop() {
	h.cancel()
	h.wg.Wait()

	// We do not close the updateCmdChannel because we don't own it
	h.sheetsHandler = nil
	h.updateCmdChannel = nil
}

// populateScheduleCommand populates the scheduleCommand struct with 4 consecutive
// schedule rows, starting from startIndex and wrapping around if it reaches the end.
func populateScheduleCommand(startIndex int, scheduleCommand *types.Schedule, schedule []*gTypes.ScheduleRow) error {
	scheduleLen := len(schedule)

	if scheduleLen == 0 {
		return errors.New("schedule slice is empty, cannot populate command")
	}

	populateRow := func(source *gTypes.ScheduleRow) (string, string, string, string) {
		djLogo, err := getDJLogo(source.Name)
		if err != nil {
			djLogo = "fallback.png"
		}

		return source.StartTime.Format("15:04"),
			source.Genre,
			source.Name,
			djLogo
	}

	// --- First Slot ---
	firstIndex := (startIndex + 0) % scheduleLen
	scheduleCommand.Time01, scheduleCommand.Genre01, scheduleCommand.Name01, scheduleCommand.Logo01 = populateRow(schedule[firstIndex])

	// --- Second Slot ---
	secondIndex := (startIndex + 1) % scheduleLen
	scheduleCommand.Time02, scheduleCommand.Genre02, scheduleCommand.Name02, scheduleCommand.Logo02 = populateRow(schedule[secondIndex])

	// --- Third Slot ---
	thirdIndex := (startIndex + 2) % scheduleLen
	scheduleCommand.Time03, scheduleCommand.Genre03, scheduleCommand.Name03, scheduleCommand.Logo03 = populateRow(schedule[thirdIndex])

	return nil
}
