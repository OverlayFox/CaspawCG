import { APIService } from "./api.js";
import { DOMUtils } from "./dom-utils.js";
import { CSS_CLASSES, SELECTORS, FIELD_TYPES } from "./constants.js";
import { LayoutManager } from "./layout.js";

/**
 * FieldManager — creates and manages the custom field rows inside each widget.
 *
 * Each field row maps a CasparCG template key to a data-source location.
 * In edit mode the row shows inputs; in live mode it shows the current value.
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

    const row = DOMUtils.createElement("div", CSS_CLASSES.FIELD_ROW);
    row.innerHTML = `
      <div class="${CSS_CLASSES.EDIT_ONLY} field-row-edit">
        <input type="text" placeholder="Key" class="f-key" value="${key}">
        <select class="f-type">
          <option value="${FIELD_TYPES.STRING}" ${type === FIELD_TYPES.STRING ? "selected" : ""}>String</option>
          <option value="${FIELD_TYPES.INT}" ${type === FIELD_TYPES.INT ? "selected" : ""}>Int</option>
          <option value="${FIELD_TYPES.FLOAT}" ${type === FIELD_TYPES.FLOAT ? "selected" : ""}>Float</option>
        </select>
        <input type="text" placeholder="Location" class="f-id" value="${id}">
        <select class="f-source">
          ${sourcesHtml}
        </select>
        <button class="delete-row-btn" aria-label="Remove field">❌</button>
      </div>
      <div class="${CSS_CLASSES.LIVE_ONLY}">
        <strong class="live-key-display"></strong>:
        <span class="live-value-display">Loading...</span>
      </div>
    `;

    if (source) {
      const sourceSelect = DOMUtils.querySelector(".f-source", row);
      if (sourceSelect) sourceSelect.value = source;
    }

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
    const idInput = DOMUtils.querySelector(SELECTORS.FIELD_ID, row);
    const sourceSelect = DOMUtils.querySelector(SELECTORS.FIELD_SOURCE, row);

    if (!keyInput || !typeSelect || !idInput || !sourceSelect) return;

    const key = keyInput.value || "Unnamed Key";
    const type = typeSelect.value;
    const identifier = idInput.value;
    const source = sourceSelect.value;

    const keyDisplay = DOMUtils.querySelector(SELECTORS.LIVE_KEY_DISPLAY, row);
    const valueDisplay = DOMUtils.querySelector(SELECTORS.LIVE_VALUE_DISPLAY, row);

    if (keyDisplay) keyDisplay.textContent = key;

    if (valueDisplay) {
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
  },
};
