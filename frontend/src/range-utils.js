/**
 * Client-side port of src/types/datasources.go's parseCellRef/NewRange —
 * keep these two in sync if the range string format ever changes.
 */
export function parseCellRef(cell) {
  if (!cell) return null;

  let i = 0;
  while (i < cell.length && /[a-zA-Z]/.test(cell[i])) i++;
  if (i === 0 || i === cell.length) return null;

  const col = cell.slice(0, i).toUpperCase();
  const rowStr = cell.slice(i);
  if (!/^\d+$/.test(rowStr)) return null;

  const row = parseInt(rowStr, 10);
  if (row <= 0) return null;

  return { col, row };
}

// splitSheetPrefix splits a "Sheet1!A1:A10"-style (optionally quoted, e.g. "'CG-System'!A1:A10")
// string into its bare (unquoted) sheet name and the remaining body. ok is false if no
// "!"-qualified sheet name is present.
function splitSheetPrefix(input) {
  const bang = (input || "").indexOf("!");
  if (bang <= 0) return null;

  let sheet = input.slice(0, bang);
  if (sheet.length >= 2 && sheet[0] === "'" && sheet[sheet.length - 1] === "'") {
    sheet = sheet.slice(1, -1);
  }
  return { sheet, body: input.slice(bang + 1) };
}

// normalizeLocationKey normalizes a raw "sheet!A1"-style key into a Sheets-API-safe,
// quoted "'sheet'!A1" key. Google rejects unquoted sheet names containing characters
// like "-" or spaces, so always quote — quoting a simple name (e.g. 'Sheet1') is valid too.
export function normalizeLocationKey(input) {
  const split = splitSheetPrefix(input);
  if (!split) {
    throw new Error(`Invalid location format: ${input} (missing sheet name, expected e.g. 'Sheet1!A1')`);
  }
  return `'${split.sheet}'!${split.body}`;
}

export function parseRange(input) {
  const split = splitSheetPrefix(input);
  if (!split) {
    throw new Error(`Invalid range format: ${input} (missing sheet name, expected e.g. 'Sheet1!A1:A10')`);
  }
  const { sheet, body } = split;

  const parts = body.split(":");
  if (parts.length !== 2) {
    throw new Error(`Invalid range format: ${input}`);
  }

  const start = parseCellRef(parts[0]);
  const end = parseCellRef(parts[1]);
  if (!start || !end) {
    throw new Error(`Invalid range format: ${input}`);
  }
  if (start.col !== end.col) {
    throw new Error(`Invalid range format: ${input} (range must span a single column)`);
  }
  if (end.row < start.row) {
    throw new Error(`Invalid range format: ${input} (end row before start row)`);
  }

  // Google rejects unquoted sheet names containing characters like "-" or spaces,
  // so always quote — quoting a simple name (e.g. 'Sheet1') is valid too.
  const keys = [];
  for (let row = start.row; row <= end.row; row++) {
    keys.push(`'${sheet}'!${start.col}${row}`);
  }
  return keys;
}
