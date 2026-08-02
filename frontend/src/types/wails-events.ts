/**
 * Types for the one-way Go -> JS event stream (window.runtime.EventsOn).
 * Wails only generates bindings for types that pass through a bound method
 * parameter/return; these are only ever emitted, so they're hand-written here.
 */

/** The envelope Go emits on the "live-data-update" channel for every event. */
export interface LiveDataUpdate {
  identifier: string;
  value: unknown;
}

/** The special identifier used for CasparCG connection keep-alive events. */
export const CASPAR_KEEP_ALIVE_IDENTIFIER = "CasparCGKeepAlive";

/** Shape of `value` on a keep-alive LiveDataUpdate (identifier === CASPAR_KEEP_ALIVE_IDENTIFIER). */
export interface CasparCGKeepAlive {
  host: string;
  port: number;
  isAlive: boolean;
}
