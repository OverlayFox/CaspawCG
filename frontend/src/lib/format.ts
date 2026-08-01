import type { types } from "../../wailsjs/go/models";

/** Pure, read-only display formatting. No parsing of user input, no effect on data sent to Go. */

export function formatFileSize(bytes: number | undefined | null): string {
  if (!bytes) return "—";
  if (bytes >= 1_000_000) return `${(bytes / 1_000_000).toFixed(1)} MB`;
  if (bytes >= 1_000) return `${(bytes / 1_000).toFixed(1)} KB`;
  return `${bytes} B`;
}

export function formatFrameRate(frameRate: types.FrameRate | undefined | null): string {
  if (!frameRate) return "—";
  return `${frameRate.Num}/${frameRate.Den}`;
}
