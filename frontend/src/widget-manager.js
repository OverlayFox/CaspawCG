import { APIService, ConnectionStateManager } from "./api.js";
import {
  CSS_CLASSES,
  FIELD_TYPES,
  INPUT_TYPES,
  SELECTORS,
} from "./constants.js";
import { DOMUtils } from "./dom-utils.js";
import { FieldManager } from "./field-manager.js";
import { LayoutManager } from "./layout.js";
import { AppState } from "./state.js";
import { parseChannelInput } from "./utils.js";

/**
 * WidgetManager — creates and manages dynamic element widget cards.
 *
 * Cards can live either as top-level GridStack items (createFromConfig) or
 * embedded inside a group container (createForGroup). Both share the same
 * inner HTML and event-listener logic; only the remove callback differs.
 */
export const WidgetManager = {
  async create() {
    return this.createFromConfig(null);
  },

  // Returns the inner HTML shared by both grid items and group-embedded cards.
  _buildInnerCardHTML(config, optionsHtml) {
    const template = config?.template || "";
    const layer = config?.layer || 1;
    const channel = config?.channelExpr || config?.channel || 1;
    const posX = config?.posX ?? 0;
    const posY = config?.posY ?? 0;
    const sizeX = config?.sizeX ?? 100;
    const sizeY = config?.sizeY ?? 100;
    const widgetName = config?.name || "Dynamic Element";
    const escapedName = widgetName.replace(/"/g, "&quot;");

    return `
      <div class="widget-header">
        <input type="text" class="widget-name-input ${CSS_CLASSES.EDIT_ONLY}" value="${escapedName}" placeholder="Element name">
        <span class="widget-name-display ${CSS_CLASSES.LIVE_ONLY}">${widgetName}</span>
      </div>
      <div class="widget-header-controls">
        <select class="api-dropdown ${CSS_CLASSES.EDIT_ONLY}">
          ${optionsHtml}
        </select>
        <div class="input-group ${CSS_CLASSES.EDIT_ONLY}">
          <label>Layer:</label>
          <input type="number" class="layer-input" min="1" max="9999" value="${layer}">
        </div>
        <div class="input-group ${CSS_CLASSES.EDIT_ONLY}">
          <label>Channel:</label>
          <input type="text" class="channel-input" placeholder="e.g. 1 or 1,2 or 1-3" value="${channel}">
        </div>
        <button class="${CSS_CLASSES.ACTION_BTN} ${CSS_CLASSES.LIVE_ONLY}" data-action="execute">Execute</button>
        <button class="${CSS_CLASSES.ACTION_BTN} ${CSS_CLASSES.LIVE_ONLY}" data-action="next">Next</button>
        <button class="${CSS_CLASSES.ACTION_BTN} ${CSS_CLASSES.LIVE_ONLY}" data-action="stop">Stop</button>
        <button class="${CSS_CLASSES.DELETE_BTN} ${CSS_CLASSES.EDIT_ONLY}" data-action="remove">Remove</button>
      </div>
      <div class="widget-position-size-controls">
        <div class="widget-controls-row">
          <div class="input-group">
            <label>Pos X (px):</label>
            <input type="number" class="pos-x-input" min="0" max="1920" value="${posX}">
          </div>
          <div class="input-group">
            <label>Pos Y (px):</label>
            <input type="number" class="pos-y-input" min="0" max="1080" value="${posY}">
          </div>
        </div>
        <div class="widget-controls-row">
          <div class="input-group">
            <label>Size X (%):</label>
            <input type="number" class="size-x-input" min="0" max="100" value="${sizeX}">
          </div>
          <div class="input-group">
            <label>Size Y (%):</label>
            <input type="number" class="size-y-input" min="0" max="100" value="${sizeY}">
          </div>
        </div>
        <div class="widget-controls-row">
          <div class="input-group">
            <label>Delay (ms):</label>
            <input type="number" class="delay-input" min="0" max="60000" value="${config?.delay || 0}">
          </div>
        </div>
      </div>
      <div class="${CSS_CLASSES.CUSTOM_FIELDS}"></div>
      <button class="${CSS_CLASSES.ADD_FIELD_BTN} ${CSS_CLASSES.EDIT_ONLY}">➕ Add Custom Field</button>
    `;
  },

  async _restoreCardState(widgetCard, config) {
    if (config?.template) {
      const dropdown = DOMUtils.querySelector(".api-dropdown", widgetCard);
      if (dropdown) dropdown.value = config.template;
    }
    if (config?.fields?.length > 0) {
      for (const field of config.fields) {
        await FieldManager.addFromConfig(widgetCard, field);
      }
    }
  },

  _attachCardListeners(widgetCard, onRemove) {
    const nameInput = widgetCard.querySelector(".widget-name-input");
    const nameDisplay = widgetCard.querySelector(".widget-name-display");

    nameInput?.addEventListener("input", () => {
      if (nameDisplay) nameDisplay.textContent = nameInput.value;
      LayoutManager.scheduleAutoSave();
    });

    DOMUtils.querySelector(
      `.${CSS_CLASSES.DELETE_BTN}`,
      widgetCard,
    )?.addEventListener("click", onRemove);

    DOMUtils.querySelectorAll(`.${CSS_CLASSES.ACTION_BTN}`, widgetCard).forEach(
      (btn) => {
        btn.addEventListener("click", (e) => {
          if (e.target.dataset.action === "execute")
            this.startWidgetAction(widgetCard);
          else if (e.target.dataset.action === "next")
            this.nextWidgetAction(widgetCard);
          else if (e.target.dataset.action === "stop")
            this.stopWidgetAction(widgetCard);
        });
      },
    );

    DOMUtils.querySelector(
      `.${CSS_CLASSES.ADD_FIELD_BTN}`,
      widgetCard,
    )?.addEventListener("click", () => {
      FieldManager.add(widgetCard);
      LayoutManager.scheduleAutoSave();
    });

    widgetCard.querySelectorAll("input, select").forEach((input) => {
      input.addEventListener("change", () => LayoutManager.scheduleAutoSave());
    });
  },

  async createFromConfig(config = null) {
    const options = await APIService.getTemplateOptions();
    // Ensure saved template is in options even if server is offline
    const mergedOptions = this._mergeWithSavedValue(options, config?.template);
    const optionsHtml = DOMUtils.createOptionsHTML(mergedOptions);
    const widgetId = config?.id || Date.now();

    const widgetElement = `
      <div class="grid-stack-item" data-widget-id="${widgetId}">
        <div class="grid-stack-item-content ${CSS_CLASSES.WIDGET_CARD}">
          ${this._buildInnerCardHTML(config, optionsHtml)}
        </div>
      </div>
    `;

    const gridOptions = {
      w: config?.w || 4,
      h: config?.h || 3,
      minW: 3,
      minH: 2,
    };
    if (config) {
      gridOptions.x = config.x;
      gridOptions.y = config.y;
    }

    const gridItem = AppState.grid.addWidget(widgetElement, gridOptions);
    const widgetCard = DOMUtils.querySelector(
      `.${CSS_CLASSES.WIDGET_CARD}`,
      gridItem,
    );

    await this._restoreCardState(widgetCard, config);

    this._attachCardListeners(widgetCard, () => {
      AppState.grid.removeWidget(gridItem);
      LayoutManager.scheduleAutoSave();
    });

    return gridItem;
  },

  // Creates a widget card for embedding inside a group container (no GridStack wrapper).
  async createForGroup(config = null) {
    const options = await APIService.getTemplateOptions();
    // Ensure saved template is in options even if server is offline
    const mergedOptions = this._mergeWithSavedValue(options, config?.template);
    const optionsHtml = DOMUtils.createOptionsHTML(mergedOptions);
    const widgetId = config?.id || `w-${Date.now()}`;

    const entry = document.createElement("div");
    entry.className = "group-widget-entry";
    entry.setAttribute("data-widget-id", widgetId);
    entry.setAttribute("data-widget-type", "dynamic");
    entry.innerHTML = `<div class="${CSS_CLASSES.WIDGET_CARD}">${this._buildInnerCardHTML(config, optionsHtml)}</div>`;

    const widgetCard = entry.querySelector(`.${CSS_CLASSES.WIDGET_CARD}`);

    await this._restoreCardState(widgetCard, config);

    this._attachCardListeners(widgetCard, () => {
      entry.remove();
      LayoutManager.scheduleAutoSave();
    });

    return entry;
  },

  stopWidgetAction(widgetCard) {
    const template = DOMUtils.querySelector(".api-dropdown", widgetCard)?.value;
    if (!template) {
      console.error("No template selected for stopping.");
      return;
    }

    const layer =
      parseInt(DOMUtils.querySelector(".layer-input", widgetCard)?.value, 10) ||
      1;

    let channels;
    try {
      channels = parseChannelInput(
        DOMUtils.querySelector(".channel-input", widgetCard)?.value ?? "1",
      ) || [1];
    } catch (e) {
      alert(`Invalid channel input: ${e.message}`);
      return;
    }

    APIService.stopCGData(template, layer, channels);
  },

  nextWidgetAction(widgetCard) {
    const template = DOMUtils.querySelector(".api-dropdown", widgetCard)?.value;
    if (!template) {
      console.error("No template selected for next.");
      return;
    }

    const layer =
      parseInt(DOMUtils.querySelector(".layer-input", widgetCard)?.value, 10) ||
      1;

    let channels;
    try {
      channels = parseChannelInput(
        DOMUtils.querySelector(".channel-input", widgetCard)?.value ?? "1",
      ) || [1];
    } catch (e) {
      alert(`Invalid channel input: ${e.message}`);
      return;
    }

    const delayVal = DOMUtils.querySelector(".delay-input", widgetCard)?.value;
    const delay = delayVal ? parseInt(delayVal, 10) * 1_000_000 : 0;

    APIService.nextCGData(template, layer, channels, delay);
  },

  async collectWidgetData(widgetCard) {
    const template = DOMUtils.querySelector(".api-dropdown", widgetCard)?.value;
    if (!template) return null;

    const layer =
      parseInt(DOMUtils.querySelector(".layer-input", widgetCard)?.value, 10) ||
      1;

    let channels;
    try {
      channels = parseChannelInput(
        DOMUtils.querySelector(".channel-input", widgetCard)?.value ?? "1",
      ) || [1];
    } catch (e) {
      alert(`Invalid channel input: ${e.message}`);
      return null;
    }

    const posXVal = DOMUtils.querySelector(".pos-x-input", widgetCard)?.value;
    const posYVal = DOMUtils.querySelector(".pos-y-input", widgetCard)?.value;
    const sizeXVal = DOMUtils.querySelector(".size-x-input", widgetCard)?.value;
    const sizeYVal = DOMUtils.querySelector(".size-y-input", widgetCard)?.value;
    const delayVal = DOMUtils.querySelector(".delay-input", widgetCard)?.value;

    const sizing = {
      posX: posXVal ? parseInt(posXVal, 10) : 0,
      posY: posYVal ? parseInt(posYVal, 10) : 0,
      sizeX: sizeXVal ? parseFloat(sizeXVal) : 100,
      sizeY: sizeYVal ? parseFloat(sizeYVal) : 100,
    };
    const delay = delayVal ? parseInt(delayVal, 10) * 1_000_000 : 0;

    const data = {};
    DOMUtils.querySelectorAll(`.${CSS_CLASSES.FIELD_ROW}`, widgetCard).forEach(
      (row) => {
        const keyInput = DOMUtils.querySelector(SELECTORS.FIELD_KEY, row);
        if (!keyInput?.value) return;

        const key = keyInput.value;
        const type =
          DOMUtils.querySelector(SELECTORS.FIELD_TYPE, row)?.value ||
          FIELD_TYPES.STRING;
        const inputType =
          DOMUtils.querySelector(SELECTORS.FIELD_INPUT_TYPE, row)?.value ||
          INPUT_TYPES.DATASOURCE;

        let rawValue;
        if (inputType === INPUT_TYPES.DIRECT) {
          rawValue =
            DOMUtils.querySelector(SELECTORS.FIELD_DIRECT_VALUE, row)?.value ??
            "";
        } else {
          const liveDisplay = DOMUtils.querySelector(
            SELECTORS.LIVE_VALUE_DISPLAY,
            row,
          );
          const identifier =
            DOMUtils.querySelector(SELECTORS.FIELD_ID, row)?.value || "";
          rawValue = liveDisplay ? liveDisplay.textContent.trim() : identifier;
        }

        let value = rawValue;
        if (type === FIELD_TYPES.INT) value = parseInt(value, 10) || 0;
        else if (type === FIELD_TYPES.FLOAT) value = parseFloat(value) || 0.0;
        data[key] = value;
      },
    );

    return { template, layer, channels, data, sizing, delay };
  },

  async startWidgetAction(widgetCard) {
    const cgData = await this.collectWidgetData(widgetCard);
    if (!cgData) {
      console.error("No template selected for execution.");
      return;
    }
    APIService.pushCGData(
      cgData.template,
      cgData.layer,
      cgData.channels,
      cgData.data,
      cgData.sizing.posX,
      cgData.sizing.posY,
      cgData.sizing.sizeX,
      cgData.sizing.sizeY,
      cgData.delay,
    );
  },

  /**
   * Merge saved template with available options.
   * Ensures the saved selection is always available in the dropdown.
   */
  _mergeWithSavedValue(options, savedValue) {
    if (!savedValue) return options;
    if (options.includes(savedValue)) return options;
    // Add saved value at the beginning so it's preserved
    return [savedValue, ...options];
  },

  /**
   * Initialize connection monitoring - call this once on app startup.
   */
  init() {
    ConnectionStateManager.subscribe((data) => {
      if (data.templates && data.templates.length > 0) {
        this.refreshAllTemplateDropdowns(data.templates);
      }
    });
  },

  /**
   * Refresh all template dropdowns with new options.
   * Preserves currently selected values if they still exist.
   */
  refreshAllTemplateDropdowns(templates) {
    const allDropdowns = document.querySelectorAll(
      `.${CSS_CLASSES.WIDGET_CARD} .api-dropdown`,
    );

    allDropdowns.forEach((dropdown) => {
      const currentValue = dropdown.value;
      const optionsHtml = DOMUtils.createOptionsHTML(templates);
      dropdown.innerHTML = optionsHtml;

      // Restore previous selection if it still exists
      if (currentValue && templates.includes(currentValue)) {
        dropdown.value = currentValue;
      }
    });

    if (allDropdowns.length > 0) {
      console.log(
        `Refreshed ${allDropdowns.length} template dropdown(s) with ${templates.length} template(s)`,
      );
    }
  },
};
