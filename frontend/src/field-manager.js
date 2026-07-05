import { APIService } from "./api.js";
import { DOMUtils } from "./dom-utils.js";
import { CSS_CLASSES, SELECTORS, FIELD_TYPES, INPUT_TYPES } from "./constants.js";
import { LayoutManager } from "./layout.js";
import { parseRange } from "./range-utils.js";

/**
 * FieldManager — creates and manages the custom field rows inside each widget.
 *
 * Each field row maps a CasparCG template key to either a data-source location
 * or a directly typed string value.
 */
export const FieldManager = {
  async add(widgetCard) {
    return this.addFromConfig(widgetCard, null);
  },

  async addFromConfig(widgetCard, config = null) {
    const container = DOMUtils.querySelector(
      `.${CSS_CLASSES.CUSTOM_FIELDS}`,
      widgetCard,
    );
    if (!container) return;

    const sources = await APIService.getDataSources();
    const sourcesHtml = DOMUtils.createOptionsHTML(sources);

    const key = config?.key || "";
    const type = config?.type || FIELD_TYPES.STRING;
    const id = config?.id || "";
    const source = config?.source || sources[0] || "";
    const inputType = config?.inputType || INPUT_TYPES.DATASOURCE;
    const value = config?.value || "";
    const range = config?.range || "";
    const offset = config?.offset ?? 0;

    const isDirect = inputType === INPUT_TYPES.DIRECT;
    const isRange = inputType === INPUT_TYPES.RANGE;
    const isDatasource = !isDirect && !isRange;

    const row = DOMUtils.createElement("div", CSS_CLASSES.FIELD_ROW);
    row.innerHTML = `
      <div class="${CSS_CLASSES.EDIT_ONLY} field-row-edit">
        <input type="text" placeholder="Key" class="f-key" value="${key}">
        <select class="f-type">
          <option value="${FIELD_TYPES.STRING}" ${type === FIELD_TYPES.STRING ? "selected" : ""}>String</option>
          <option value="${FIELD_TYPES.INT}" ${type === FIELD_TYPES.INT ? "selected" : ""}>Int</option>
          <option value="${FIELD_TYPES.FLOAT}" ${type === FIELD_TYPES.FLOAT ? "selected" : ""}>Float</option>
        </select>
        <select class="f-input-type">
          <option value="${INPUT_TYPES.DATASOURCE}" ${isDatasource ? "selected" : ""}>Data Source</option>
          <option value="${INPUT_TYPES.DIRECT}" ${isDirect ? "selected" : ""}>Direct Input</option>
          <option value="${INPUT_TYPES.RANGE}" ${isRange ? "selected" : ""}>Data Source Range</option>
        </select>
        <div class="f-datasource-inputs" style="${isDatasource ? "" : "display:none"}">
          <input type="text" placeholder="Location" class="f-id" value="${id}">
          <select class="f-source">
            ${sourcesHtml}
          </select>
        </div>
        <div class="f-direct-inputs" style="${isDirect ? "" : "display:none"}">
          <input type="text" placeholder="Value" class="f-value" value="${value}">
        </div>
        <div class="f-range-inputs" style="${isRange ? "" : "display:none"}">
          <input type="text" placeholder="Range e.g. Sheet1!A1:A10" class="f-range" value="${range}">
          <select class="f-source">
            ${sourcesHtml}
          </select>
          <input type="number" placeholder="Offset" class="f-offset" min="0" value="${offset}">
        </div>
        <button class="delete-row-btn" aria-label="Remove field">❌</button>
      </div>
      <div class="${CSS_CLASSES.LIVE_ONLY}">
        <strong class="live-key-display"></strong>:
        <span class="live-value-display">Loading...</span>
      </div>
    `;

    if (source) {
      DOMUtils.querySelectorAll(".f-source", row).forEach((sourceSelect) => {
        sourceSelect.value = source;
      });
    }

    const inputTypeSelect = DOMUtils.querySelector(SELECTORS.FIELD_INPUT_TYPE, row);
    inputTypeSelect?.addEventListener("change", () => {
      const mode = inputTypeSelect.value;
      DOMUtils.querySelector(SELECTORS.FIELD_DATASOURCE_INPUTS, row).style.display = mode === INPUT_TYPES.DATASOURCE ? "" : "none";
      DOMUtils.querySelector(SELECTORS.FIELD_DIRECT_INPUTS, row).style.display = mode === INPUT_TYPES.DIRECT ? "" : "none";
      DOMUtils.querySelector(SELECTORS.FIELD_RANGE_INPUTS, row).style.display = mode === INPUT_TYPES.RANGE ? "" : "none";
      LayoutManager.scheduleAutoSave();
    });

    DOMUtils.querySelector(".delete-row-btn", row)?.addEventListener(
      "click",
      () => {
        row.remove();
        LayoutManager.scheduleAutoSave();
      },
    );

    container.appendChild(row);
  },

  async updateLiveData(row) {
    const keyInput = DOMUtils.querySelector(SELECTORS.FIELD_KEY, row);
    const typeSelect = DOMUtils.querySelector(SELECTORS.FIELD_TYPE, row);
    const inputTypeSelect = DOMUtils.querySelector(SELECTORS.FIELD_INPUT_TYPE, row);

    if (!keyInput || !typeSelect) return;

    const key = keyInput.value || "Unnamed Key";
    const inputType = inputTypeSelect?.value || INPUT_TYPES.DATASOURCE;

    const keyDisplay = DOMUtils.querySelector(SELECTORS.LIVE_KEY_DISPLAY, row);
    const valueDisplay = DOMUtils.querySelector(SELECTORS.LIVE_VALUE_DISPLAY, row);

    if (keyDisplay) keyDisplay.textContent = key;

    if (valueDisplay) {
      if (inputType === INPUT_TYPES.DIRECT) {
        const directInput = DOMUtils.querySelector(SELECTORS.FIELD_DIRECT_VALUE, row);
        valueDisplay.textContent = directInput?.value ?? "";
      } else if (inputType === INPUT_TYPES.RANGE) {
        const type = typeSelect.value;
        const rangeInput = DOMUtils.querySelector(SELECTORS.FIELD_RANGE, row);
        const offsetInput = DOMUtils.querySelector(SELECTORS.FIELD_OFFSET, row);
        const sourceSelect = DOMUtils.querySelector(`${SELECTORS.FIELD_RANGE_INPUTS} .f-source`, row);
        if (!rangeInput || !sourceSelect) return;

        const source = sourceSelect.value;
        const offset = parseInt(offsetInput?.value, 10) || 0;

        valueDisplay.textContent = "Fetching...";
        try {
          const keys = parseRange(rangeInput.value);
          const identifier = keys[offset];
          if (identifier === undefined) {
            valueDisplay.textContent = "Offset out of range";
            return;
          }
          valueDisplay.textContent = await APIService.fetchLiveData(
            identifier,
            type,
            source,
          );
        } catch (error) {
          console.error("Failed to fetch live range data:", error);
          valueDisplay.textContent = "Error loading data";
        }
      } else {
        const type = typeSelect.value;
        const idInput = DOMUtils.querySelector(SELECTORS.FIELD_ID, row);
        const sourceSelect = DOMUtils.querySelector(`${SELECTORS.FIELD_DATASOURCE_INPUTS} .f-source`, row);
        if (!idInput || !sourceSelect) return;

        const identifier = idInput.value;
        const source = sourceSelect.value;

        valueDisplay.textContent = "Fetching...";
        try {
          valueDisplay.textContent = await APIService.fetchLiveData(
            identifier,
            type,
            source,
          );
        } catch (error) {
          console.error("Failed to fetch live data:", error);
          valueDisplay.textContent = "Error loading data";
        }
      }
    }
  },
};
