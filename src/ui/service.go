package ui

import (
	"context"
	"sync"
	"time"

	"github.com/overlayfox/casparcg-amcp-go/types/responses"

	"github.com/overlayfox/caspaw-cg/src/types"
	"github.com/overlayfox/caspaw-cg/src/ui/update"
)

// UIService bridges the UI with the GoLang system
type UIService struct {
	app               *App
	datasourceManager types.DatasourceManager
	casparCGClient    types.CasparCGClient
	updateHandler     *update.Handler

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewUIService(upstreamCtx context.Context, app *App, datasourceManager types.DatasourceManager, casparCGClient types.CasparCGClient) *UIService {
	ctx, cancel := context.WithCancel(upstreamCtx)
	return &UIService{
		app:               app,
		datasourceManager: datasourceManager,
		casparCGClient:    casparCGClient,
		updateHandler:     update.NewUpdateHandler(ctx, app.logger, datasourceManager, casparCGClient),
		ctx:               ctx,
		cancel:            cancel,
	}
}

func (u *UIService) SaveLayout(config LayoutConfig) error {
	u.app.logger.Info().Msg("Saving layout configuration")
	return SaveLayout(config)
}

func (u *UIService) LoadLayout() (LayoutConfig, error) {
	u.app.logger.Info().Msg("Loading layout configuration")
	return LoadLayout()
}

func (u *UIService) GetDataSources() []string {
	names := u.datasourceManager.GetDataSourceNames()
	if len(names) == 0 {
		return []string{
			"No datasources available",
		}
	}
	return names
}

func (u *UIService) GetCasparCGTemplates() []string {
	templates, err := u.casparCGClient.GetTemplates()
	if err != nil {
		u.app.logger.Error().Err(err).Msg("Failed to get templates from CasparCG client")
		return nil
	}
	return templates
}

func (u *UIService) GetCasparCGMedia() []string {
	media, err := u.casparCGClient.GetMedia()
	if err != nil {
		u.app.logger.Error().Err(err).Msg("Failed to get media from CasparCG client")
		return nil
	}
	return media
}

func (u *UIService) GetCasparCGMediaInfo(filename string) (responses.CINF, error) {
	info, err := u.casparCGClient.GetMediaInfo(filename)
	if err != nil {
		u.app.logger.Error().Err(err).Msgf("Failed to get media info for '%s' from CasparCG client", filename)
		return responses.CINF{}, err
	}
	return info, nil
}

// pushCGData, stopCGData, nextCGData, playMedia and stopMedia are the fire-and-forget
// workers shared by the single-item and group/update-job entry points below. Channel
// parsing and field coercion always happen once, synchronously, in the calling public
// method before any of these run.

func (u *UIService) pushCGData(template string, layer int, channels []int, data map[string]any, sizing types.Sizing, delay time.Duration) {
	u.wg.Go(func() {
		if err := u.casparCGClient.AddCGData(template, layer, channels, data, sizing, delay); err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to push CG data to template '%s' on layer %d, channels %v", template, layer, channels)
		}
	})
}

func (u *UIService) stopCGData(template string, layer int, channels []int, delay time.Duration) {
	u.wg.Go(func() {
		if err := u.casparCGClient.StopCGData(template, layer, channels, delay); err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to stop CG data for template '%s' on layer %d, channels %v", template, layer, channels)
		}
	})
}

func (u *UIService) nextCGData(template string, layer int, channels []int, delay time.Duration) {
	u.wg.Go(func() {
		if err := u.casparCGClient.NextCGData(template, layer, channels, delay); err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to go to next CG data for template '%s' on layer %d, channels %v", template, layer, channels)
		}
	})
}

func (u *UIService) PushCasparCGData(template string, layer int, channelExpr string, fields []types.LiteralField, sizing types.Sizing, delayMs int) error {
	channels, err := types.ParseChannelExpression(channelExpr)
	if err != nil {
		return err
	}
	u.pushCGData(template, layer, channels, types.BuildDataMap(fields), sizing, time.Duration(delayMs)*time.Millisecond)
	return nil
}

func (u *UIService) StopCasparCGData(template string, layer int, channelExpr string, delayMs int) error {
	channels, err := types.ParseChannelExpression(channelExpr)
	if err != nil {
		return err
	}
	u.stopCGData(template, layer, channels, time.Duration(delayMs)*time.Millisecond)
	return nil
}

func (u *UIService) NextCasparCGData(template string, layer int, channelExpr string, delayMs int) error {
	channels, err := types.ParseChannelExpression(channelExpr)
	if err != nil {
		return err
	}
	u.nextCGData(template, layer, channels, time.Duration(delayMs)*time.Millisecond)
	return nil
}

// RangeField describes a single template field that should be continuously
// resolved from a range of locations in a data source.
type RangeField struct {
	CasparKey string
	Type      types.DataType
	Source    string
	Range     string
	Offset    int
}

// UpdateCasparCGData pushes an initial snapshot of literalFields plus the current values of
// rangeFields, then starts an update job that continuously re-resolves rangeFields from their
// data sources and pushes the results to the template at the specified interval.
//
// It returns a unique identifier for the update job.
func (u *UIService) UpdateCasparCGData(template string, layer int, channelExpr string, literalFields []types.LiteralField, rangeFields []RangeField, sizing types.Sizing, delayMs, updateIntervalMs int) (uuid string, err error) {
	channels, err := types.ParseChannelExpression(channelExpr)
	if err != nil {
		return "", err
	}

	casparMaps := make(map[string]*update.Resolver, len(rangeFields))
	for _, rf := range rangeFields {
		dataRange, err := types.NewRange(rf.Range)
		if err != nil {
			u.app.logger.Error().Err(err).Str("range", rf.Range).Msg("Failed to parse range")
			return "", err
		}
		for i := range dataRange.Locations {
			dataRange.Locations[i].Type = rf.Type
		}

		ds, err := u.datasourceManager.GetDataSource(rf.Source)
		if err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to get datasource '%s'", rf.Source)
			return "", err
		}

		resolver := update.NewResolver(ds, dataRange, rf.Offset)
		casparMaps[rf.CasparKey] = &resolver
	}

	resolvedData := types.BuildDataMap(literalFields)
	for casparKey, resolver := range casparMaps {
		value, err := resolver.GetData()
		if err != nil {
			u.app.logger.Error().Err(err).Str("casparKey", casparKey).Msg("Failed to get data from datasource")
		}
		resolvedData[casparKey] = value
		resolver.Advance()
	}
	u.pushCGData(template, layer, channels, resolvedData, sizing, time.Duration(delayMs)*time.Millisecond)

	uuid = u.updateHandler.AddUpdateJob(template, layer, channels, u.casparCGClient, casparMaps, time.Duration(updateIntervalMs)*time.Millisecond)
	return uuid, nil
}

// RemoveUpdateJob stops and removes the update job identified by uuid.
func (u *UIService) RemoveUpdateJob(uuid string) error {
	return u.updateHandler.RemoveUpdateJob(uuid)
}

// PrimeDataSources resolves and primes every field subscription's data source in one
// batch call (grouping locations by source so each source is only primed once), then
// returns each subscription's canonical live-data identifier and its initial value.
func (u *UIService) PrimeDataSources(subs []types.FieldSubscription) ([]types.FieldSubscriptionResult, error) {
	results := make([]types.FieldSubscriptionResult, len(subs))
	identifiers := make([]string, len(subs))
	locationsBySource := make(map[string][]types.Location)

	for i, sub := range subs {
		identifier, locations, err := sub.ResolveLocations()
		if err != nil {
			results[i] = types.FieldSubscriptionResult{Error: err.Error()}
			continue
		}
		identifiers[i] = identifier
		locationsBySource[sub.Source] = append(locationsBySource[sub.Source], locations...)
	}

	primeFailed := make(map[string]string)
	for source, locations := range locationsBySource {
		ds, err := u.datasourceManager.GetDataSource(source)
		if err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to get datasource '%s'", source)
			primeFailed[source] = err.Error()
			continue
		}

		u.app.logger.Info().Msgf("Priming datasource '%s' with %d location(s)", source, len(locations))
		if err := ds.Prime(locations); err != nil { // TODO: add removal of primed data when a new PrimeDataSources is called
			u.app.logger.Error().Err(err).Msgf("Failed to prime datasource '%s'", source)
			primeFailed[source] = err.Error()
		}
	}

	for i, sub := range subs {
		if results[i].Error != "" {
			continue
		}
		if errMsg, ok := primeFailed[sub.Source]; ok {
			results[i] = types.FieldSubscriptionResult{Identifier: identifiers[i], Error: errMsg}
			continue
		}

		ds, err := u.datasourceManager.GetDataSource(sub.Source)
		if err != nil {
			results[i] = types.FieldSubscriptionResult{Identifier: identifiers[i], Error: err.Error()}
			continue
		}
		data, err := ds.Get(identifiers[i])
		if err != nil {
			results[i] = types.FieldSubscriptionResult{Identifier: identifiers[i], Error: err.Error()}
			continue
		}
		results[i] = types.FieldSubscriptionResult{Identifier: identifiers[i], Value: data.Value}
	}

	return results, nil
}

func (u *UIService) RemoveDataSourcesPrimes() {
	u.app.logger.Debug().Msg("Removing all primed data from all datasources")

	for _, dsName := range u.datasourceManager.GetDataSourceNames() {
		ds, err := u.datasourceManager.GetDataSource(dsName)
		if err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to get datasource '%s'", dsName)
			continue
		}
		if err := ds.RemoveAllPrimes(); err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to remove primes for datasource '%s'", dsName)
		}
	}
}

type CGDataGroup struct {
	Template    string
	Layer       int
	ChannelExpr string
	Fields      []types.LiteralField
	Sizing      types.Sizing
	DelayMs     int
}

func (g CGDataGroup) resolve() (channels []int, data map[string]any, delay time.Duration, err error) {
	channels, err = types.ParseChannelExpression(g.ChannelExpr)
	if err != nil {
		return nil, nil, 0, err
	}
	return channels, types.BuildDataMap(g.Fields), time.Duration(g.DelayMs) * time.Millisecond, nil
}

// PushCasparCGDataGroup, StopCasparCGDataGroup and NextCasparCGDataGroup validate every
// item up front and return the first error encountered without pushing/stopping/nexting
// anything, so a group action either fully applies or fully fails.

func (u *UIService) PushCasparCGDataGroup(dataGroups []CGDataGroup) error {
	for _, g := range dataGroups {
		channels, data, delay, err := g.resolve()
		if err != nil {
			return err
		}
		u.pushCGData(g.Template, g.Layer, channels, data, g.Sizing, delay)
	}
	return nil
}

func (u *UIService) StopCasparCGDataGroup(dataGroups []CGDataGroup) error {
	for _, g := range dataGroups {
		channels, _, delay, err := g.resolve()
		if err != nil {
			return err
		}
		u.stopCGData(g.Template, g.Layer, channels, delay)
	}
	return nil
}

func (u *UIService) NextCasparCGDataGroup(dataGroups []CGDataGroup) error {
	for _, g := range dataGroups {
		channels, _, delay, err := g.resolve()
		if err != nil {
			return err
		}
		u.nextCGData(g.Template, g.Layer, channels, delay)
	}
	return nil
}

func (u *UIService) playMedia(filename string, layer int, channels []int, loop bool, delay time.Duration) {
	u.wg.Go(func() {
		if err := u.casparCGClient.PlayMedia(filename, layer, channels, loop, delay); err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to play media '%s' on layer %d, channels %v", filename, layer, channels)
		}
	})
}

func (u *UIService) stopMedia(layer int, channels []int, delay time.Duration) {
	u.wg.Go(func() {
		if err := u.casparCGClient.StopMedia(layer, channels, delay); err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to stop media on layer %d, channels %v", layer, channels)
		}
	})
}

func (u *UIService) PlayCasparCGMedia(filename string, layer int, channelExpr string, loop bool, delayMs int) error {
	channels, err := types.ParseChannelExpression(channelExpr)
	if err != nil {
		return err
	}
	u.playMedia(filename, layer, channels, loop, time.Duration(delayMs)*time.Millisecond)
	return nil
}

func (u *UIService) StopCasparCGMedia(layer int, channelExpr string, delayMs int) error {
	channels, err := types.ParseChannelExpression(channelExpr)
	if err != nil {
		return err
	}
	u.stopMedia(layer, channels, time.Duration(delayMs)*time.Millisecond)
	return nil
}

type MediaGroupItem struct {
	Filename    string
	Layer       int
	ChannelExpr string
	Loop        bool
	DelayMs     int
}

func (u *UIService) PlayCasparCGMediaGroup(items []MediaGroupItem) error {
	for _, item := range items {
		channels, err := types.ParseChannelExpression(item.ChannelExpr)
		if err != nil {
			return err
		}
		u.playMedia(item.Filename, item.Layer, channels, item.Loop, time.Duration(item.DelayMs)*time.Millisecond)
	}
	return nil
}

func (u *UIService) StopCasparCGMediaGroup(items []MediaGroupItem) error {
	for _, item := range items {
		channels, err := types.ParseChannelExpression(item.ChannelExpr)
		if err != nil {
			return err
		}
		u.stopMedia(item.Layer, channels, time.Duration(item.DelayMs)*time.Millisecond)
	}
	return nil
}

func (u *UIService) ClearChannels(channelExpr string) error {
	channels, err := types.ParseChannelExpression(channelExpr)
	if err != nil {
		return err
	}
	u.wg.Go(func() {
		u.casparCGClient.ClearChannels(channels)
	})
	return nil
}

func (u *UIService) ClearAll() {
	u.wg.Go(func() {
		u.casparCGClient.ClearAll()
	})
}

func (u *UIService) Close() {
	u.cancel()
	u.wg.Wait()
}
