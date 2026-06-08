// Parses "1", "1-3", "1,3-5" into a sorted, deduplicated array of integers.
// Returns null if the input is blank (meaning: all channels).
// Throws if any token is invalid.
export function parseChannelInput(raw) {
  const trimmed = raw.trim();
  if (!trimmed) return null;

  const channels = new Set();

  for (const token of trimmed.split(",")) {
    const part = token.trim();
    const rangeMatch = part.match(/^(\d+)-(\d+)$/);
    if (rangeMatch) {
      const start = parseInt(rangeMatch[1], 10);
      const end = parseInt(rangeMatch[2], 10);
      if (start > end) throw new Error(`Invalid range "${part}"`);
      for (let i = start; i <= end; i++) channels.add(i);
    } else if (/^\d+$/.test(part)) {
      channels.add(parseInt(part, 10));
    } else {
      throw new Error(`Invalid token "${part}"`);
    }
  }

  return [...channels].sort((a, b) => a - b);
}
