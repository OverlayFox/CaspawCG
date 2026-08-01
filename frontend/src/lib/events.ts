import { EventsOn } from "../../wailsjs/runtime/runtime";
import { CASPAR_KEEP_ALIVE_IDENTIFIER, type CasparCGKeepAlive, type LiveDataUpdate } from "../types/wails-events";

/** CustomEvent name republished on `window` for every "live-data-update" from Go. */
export const LIVE_DATA_EVENT = "casp-live-data";

/** CustomEvent name republished on `window` for CasparCG keep-alive updates. */
export const CASPAR_STATUS_EVENT = "casp-caspar-status";

export type LiveDataEvent = CustomEvent<LiveDataUpdate>;
export type CasparStatusEvent = CustomEvent<CasparCGKeepAlive>;

/**
 * Sets up the single Wails event listener for the app and republishes each update as
 * a plain DOM CustomEvent, so components can listen with the standard
 * addEventListener/removeEventListener pattern in connectedCallback/disconnectedCallback
 * instead of every component talking to the Wails runtime directly.
 */
export function initLiveEvents(): void {
  EventsOn("live-data-update", (payload: LiveDataUpdate) => {
    if (payload.identifier === CASPAR_KEEP_ALIVE_IDENTIFIER) {
      window.dispatchEvent(new CustomEvent(CASPAR_STATUS_EVENT, { detail: payload.value as CasparCGKeepAlive }));
      return;
    }
    window.dispatchEvent(new CustomEvent(LIVE_DATA_EVENT, { detail: payload }));
  });
}
