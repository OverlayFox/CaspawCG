/**
 * Typed wrapper over the Wails-generated Go bindings. Every function here is a thin
 * pass-through to a UIService method in src/ui/service.go — no parsing, no unit
 * conversion, no coercion happens here, all of that lives in Go now. Errors are not
 * caught here; callers decide how to surface them (e.g. show the message to the user).
 */
import { types, ui } from "../../wailsjs/go/models";
import * as UIService from "../../wailsjs/go/ui/UIService";

export async function getDataSources(): Promise<string[]> {
  return UIService.GetDataSources();
}

export async function getCasparCGTemplates(): Promise<string[]> {
  return UIService.GetCasparCGTemplates();
}

export async function getCasparCGMedia(): Promise<string[]> {
  return UIService.GetCasparCGMedia();
}

export async function getCasparCGMediaInfo(filename: string) {
  return UIService.GetCasparCGMediaInfo(filename);
}

export async function pushCGData(
  template: string,
  layer: number,
  channelExpr: string,
  fields: types.LiteralField[],
  sizing: types.Sizing,
  delayMs: number,
): Promise<void> {
  return UIService.PushCasparCGData(
    template,
    layer,
    channelExpr,
    fields,
    sizing,
    delayMs,
  );
}

export async function stopCGData(
  template: string,
  layer: number,
  channelExpr: string,
  delayMs: number,
): Promise<void> {
  return UIService.StopCasparCGData(template, layer, channelExpr, delayMs);
}

export async function nextCGData(
  template: string,
  layer: number,
  channelExpr: string,
  delayMs: number,
): Promise<void> {
  return UIService.NextCasparCGData(template, layer, channelExpr, delayMs);
}

export async function updateCGData(
  template: string,
  layer: number,
  channelExpr: string,
  literalFields: types.LiteralField[],
  rangeFields: ui.RangeField[],
  sizing: types.Sizing,
  delayMs: number,
  updateIntervalMs: number,
): Promise<string> {
  return UIService.UpdateCasparCGData(
    template,
    layer,
    channelExpr,
    literalFields,
    rangeFields,
    sizing,
    delayMs,
    updateIntervalMs,
  );
}

export async function removeUpdateJob(uuid: string): Promise<void> {
  return UIService.RemoveUpdateJob(uuid);
}

export async function primeDataSources(
  subs: types.FieldSubscription[],
): Promise<types.FieldSubscriptionResult[]> {
  return UIService.PrimeDataSources(subs);
}

export async function removeDataSources(): Promise<void> {
  return UIService.RemoveDataSourcesPrimes();
}

export async function pushCGDataGroup(
  dataGroups: ui.CGDataGroup[],
): Promise<void> {
  return UIService.PushCasparCGDataGroup(dataGroups);
}

export async function stopCGDataGroup(
  dataGroups: ui.CGDataGroup[],
): Promise<void> {
  return UIService.StopCasparCGDataGroup(dataGroups);
}

export async function nextCGDataGroup(
  dataGroups: ui.CGDataGroup[],
): Promise<void> {
  return UIService.NextCasparCGDataGroup(dataGroups);
}

export async function playMedia(
  filename: string,
  layer: number,
  channelExpr: string,
  loop: boolean,
  delayMs: number,
): Promise<void> {
  return UIService.PlayCasparCGMedia(
    filename,
    layer,
    channelExpr,
    loop,
    delayMs,
  );
}

export async function stopMedia(
  layer: number,
  channelExpr: string,
  delayMs: number,
): Promise<void> {
  return UIService.StopCasparCGMedia(layer, channelExpr, delayMs);
}

export async function playMediaGroup(
  items: ui.MediaGroupItem[],
): Promise<void> {
  return UIService.PlayCasparCGMediaGroup(items);
}

export async function stopMediaGroup(
  items: ui.MediaGroupItem[],
): Promise<void> {
  return UIService.StopCasparCGMediaGroup(items);
}

export async function clearChannels(channelExpr: string): Promise<void> {
  return UIService.ClearChannels(channelExpr);
}

export async function clearAll(): Promise<void> {
  return UIService.ClearAll();
}

export async function saveLayout(config: ui.LayoutConfig): Promise<void> {
  return UIService.SaveLayout(config);
}

export async function loadLayout(): Promise<ui.LayoutConfig> {
  return UIService.LoadLayout();
}
