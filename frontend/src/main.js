import { GridStack } from "gridstack";
import "gridstack/dist/gridstack.min.css";

import { initLiveEvents } from "./events";
import { LayoutManager } from "./layout";

/**
 * Application State
 */
const AppState = {
  isLiveMode: false,
  grid: null,
};

/**
 * Constants
 */
const CSS_CLASSES = {
  EDIT_ONLY: "edit-only",
  LIVE_ONLY: "live-only",
  IS_LIVE: "is-live",
  MODE_LIVE: "mode-live",
  FIELD_ROW: "field-row",
  WIDGET_CARD: "widget-card",
  CUSTOM_FIELDS: "custom-fields-container",
  DELETE_BTN: "delete-btn",
  ACTION_BTN: "action-btn",
  ADD_FIELD_BTN: "add-field-btn",
};

const SELECTORS = {
  TOGGLE_MODE_BTN: "#toggle-mode-btn",
  ADD_WIDGET_BTN: "#add-widget-btn",
  FIELD_KEY: ".f-key",
  FIELD_TYPE: ".f-type",
  FIELD_ID: ".f-id",
  FIELD_SOURCE: ".f-source",
  LIVE_KEY_DISPLAY: ".live-key-display",
  LIVE_VALUE_DISPLAY: ".live-value-display",
};

const FIELD_TYPES = {
  STRING: "string",
  INT: "int",
  FLOAT: "float",
};

/**
 * API Service - handles all backend communication
 */
const APIService = {
  async getTemplateOptions() {
    try {
      return await window.go.ui.UIService.GetCasparCGTemplates();
    } catch (error) {
      console.error("Failed to fetch template options:", error);
      return [];
    }
  },

  async getDataSources() {
    try {
      return await window.go.ui.UIService.GetDataSources();
    } catch (error) {
      console.error("Failed to fetch data sources:", error);
      return [];
    }
  },

  async pushCGData(
    template,
    layer = 1,
    channel = 1,
    data,
    posX = null,
    posY = null,
    sizeX = null,
    sizeY = null,
  ) {
    try {
      const sizing = {
        posX: posX !== null ? parseInt(posX, 10) : 0,
        posY: posY !== null ? parseInt(posY, 10) : 0,
        sizeX: sizeX !== null ? parseFloat(sizeX) : 100,
        sizeY: sizeY !== null ? parseFloat(sizeY) : 100,
      };

      await window.go.ui.UIService.PushCasparCGData(
        template,
        layer,
        channel,
        data,
        sizing,
      );
    } catch (error) {
      console.error("Failed to push CG data:", error);
    }
  },

  async stopCGData(template, layer = 1, channel = 1) {
    try {
      await window.go.ui.UIService.StopCasparCGData(template, layer, channel);
    } catch (error) {
      console.error("Failed to stop CG data:", error);
    }
  },

  async fetchLiveData(identifier, type, source) {
    const result = await window.go.ui.UIService.GetDataSourceValue(source, {
      Key: identifier,
      Type: type,
    });
    return result.Value;
  },
};

/**
 * DOM Utilities
 */
const DOMUtils = {
  createElement(tag, className = "", innerHTML = "") {
    const element = document.createElement(tag);
    if (className) element.className = className;
    if (innerHTML) element.innerHTML = innerHTML;
    return element;
  },

  createOptionsHTML(options) {
    return options
      .map((opt) => `<option value="${opt}">${opt}</option>`)
      .join("");
  },

  querySelector(selector, parent = document) {
    return parent.querySelector(selector);
  },

  querySelectorAll(selector, parent = document) {
    return parent.querySelectorAll(selector);
  },
};

/**
 * Widget Manager - handles all widget-related operations
 */
const WidgetManager = {
  async create() {
    return this.createFromConfig(null);
  },

  async createFromConfig(config = null) {
    const options = await APIService.getTemplateOptions();
    const optionsHtml = DOMUtils.createOptionsHTML(options);

    const template = config?.template || "";
    const layer = config?.layer || 1;
    const channel = config?.channel || 1;
    const posX = config?.posX ?? 0;
    const posY = config?.posY ?? 0;
    const sizeX = config?.sizeX ?? 100;
    const sizeY = config?.sizeY ?? 100;

    const widgetElement = `
      <div class="grid-stack-item" data-widget-id="${config?.id || Date.now()}">
        <div class="grid-stack-item-content ${CSS_CLASSES.WIDGET_CARD}">
          <div class="widget-header">
            <strong>Dynamic Element</strong>
          </div>
          <div class="widget-header-controls">
            <select class="api-dropdown ${CSS_CLASSES.EDIT_ONLY}">
              ${optionsHtml}
            </select>
            <div class="input-group ${CSS_CLASSES.EDIT_ONLY}">
              <label for="layer-input">Layer:</label>
              <input type="number" class="layer-input" id="layer-input" min="1" max="9999" value="${layer}">
            </div>
            <div class="input-group ${CSS_CLASSES.EDIT_ONLY}">
              <label for="channel-input">Channel:</label>
              <input type="number" class="channel-input" id="channel-input" min="1" max="9999" value="${channel}">
            </div>
            <button class="${CSS_CLASSES.ACTION_BTN} ${CSS_CLASSES.LIVE_ONLY}" data-action="execute">Execute</button>
            <button class="${CSS_CLASSES.ACTION_BTN} ${CSS_CLASSES.LIVE_ONLY}" data-action="stop">Stop</button>
            <button class="${CSS_CLASSES.DELETE_BTN} ${CSS_CLASSES.EDIT_ONLY}" data-action="remove">Remove</button>
          </div>
          <div class="widget-position-size-controls">
            <div class="input-group">
              <label>Pos X (pixels):</label>
              <input type="number" class="pos-x-input" min="0" max="1920" value="${posX}">
            </div>
            <div class="input-group">
              <label>Pos Y (pixels):</label>
              <input type="number" class="pos-y-input" min="0" max="1080" value="${posY}">
            </div>
            <div class="input-group">
              <label>Size X (%):</label>
              <input type="number" class="size-x-input" min="0" max="100" value="${sizeX}">
            </div>
            <div class="input-group">
              <label>Size Y (%):</label>
              <input type="number" class="size-y-input" min="0" max="100" value="${sizeY}">
            </div>
          </div>
          <div class="${CSS_CLASSES.CUSTOM_FIELDS}"></div>
          <button class="${CSS_CLASSES.ADD_FIELD_BTN} ${CSS_CLASSES.EDIT_ONLY}">➕ Add Custom Field</button>
        </div>
      </div>
    `;

    const gridOptions = {
      w: config?.w || 7,
      h: config?.h || 2,
      minW: 7,
      minH: 2,
    };

    if (config) {
      gridOptions.x = config.x;
      gridOptions.y = config.y;
    }

    const gridItem = AppState.grid.addWidget(widgetElement, gridOptions);

    // Set template selection if provided
    if (template) {
      const dropdown = DOMUtils.querySelector(".api-dropdown", gridItem);
      if (dropdown) dropdown.value = template;
    }

    // Restore fields if provided
    if (config?.fields && config.fields.length > 0) {
      const widgetCard = DOMUtils.querySelector(
        `.${CSS_CLASSES.WIDGET_CARD}`,
        gridItem,
      );
      for (const field of config.fields) {
        await FieldManager.addFromConfig(widgetCard, field);
      }
    }

    this.attachEventListeners(gridItem);
    return gridItem;
  },

  attachEventListeners(gridItem) {
    const widgetCard = DOMUtils.querySelector(
      `.${CSS_CLASSES.WIDGET_CARD}`,
      gridItem,
    );

    // Delete button handler
    const deleteBtn = DOMUtils.querySelector(
      `.${CSS_CLASSES.DELETE_BTN}`,
      widgetCard,
    );
    if (deleteBtn) {
      deleteBtn.addEventListener("click", () => {
        AppState.grid.removeWidget(gridItem);
        LayoutManager.scheduleAutoSave(
          AppState.grid,
          CSS_CLASSES,
          SELECTORS,
          FIELD_TYPES,
          DOMUtils,
        );
      });
    }

    // Execute and Stop button handlers
    const actionBtns = DOMUtils.querySelectorAll(
      `.${CSS_CLASSES.ACTION_BTN}`,
      widgetCard,
    );
    actionBtns.forEach((btn) => {
      btn.addEventListener("click", (e) => {
        const action = e.target.dataset.action;
        if (action === "execute") {
          this.startWidgetAction(widgetCard);
        } else if (action === "stop") {
          this.stopWidgetAction(widgetCard);
        }
      });
    });

    // Add field button handler
    const addFieldBtn = DOMUtils.querySelector(
      `.${CSS_CLASSES.ADD_FIELD_BTN}`,
      widgetCard,
    );
    if (addFieldBtn) {
      addFieldBtn.addEventListener("click", () => {
        FieldManager.add(widgetCard);
        LayoutManager.scheduleAutoSave(
          AppState.grid,
          CSS_CLASSES,
          SELECTORS,
          FIELD_TYPES,
          DOMUtils,
        );
      });
    }

    // Add change listeners for auto-save
    const inputs = widgetCard.querySelectorAll("input, select");
    inputs.forEach((input) => {
      input.addEventListener("change", () => {
        LayoutManager.scheduleAutoSave(
          AppState.grid,
          CSS_CLASSES,
          SELECTORS,
          FIELD_TYPES,
          DOMUtils,
        );
      });
    });
  },

  stopWidgetAction(widgetCard) {
    const dropdown = DOMUtils.querySelector(".api-dropdown", widgetCard);
    const layerInput = DOMUtils.querySelector(".layer-input", widgetCard);
    const channelInput = DOMUtils.querySelector(".channel-input", widgetCard);
    const posXInput = DOMUtils.querySelector(".pos-x-input", widgetCard);
    const posYInput = DOMUtils.querySelector(".pos-y-input", widgetCard);
    const sizeXInput = DOMUtils.querySelector(".size-x-input", widgetCard);
    const sizeYInput = DOMUtils.querySelector(".size-y-input", widgetCard);

    const selectedTemplate = dropdown?.value;
    const layer = parseInt(layerInput?.value, 10) || 1;
    const channel = parseInt(channelInput?.value, 10) || 1;

    // These can be null if not provided
    const posX = posXInput?.value ? parseInt(posXInput.value, 10) : null;
    const posY = posYInput?.value ? parseInt(posYInput.value, 10) : null;
    const sizeX = sizeXInput?.value ? parseFloat(sizeXInput.value) : null;
    const sizeY = sizeYInput?.value ? parseFloat(sizeYInput.value) : null;

    if (!selectedTemplate) {
      console.error("No template selected for stopping.");
      return;
    }

    console.log(
      `Stopping template on layer ${layer}, channel ${channel}: ${selectedTemplate}`,
    );
    APIService.stopCGData(
      selectedTemplate,
      layer,
      channel,
      posX,
      posY,
      sizeX,
      sizeY,
    );
  },

  startWidgetAction(widgetCard) {
    const dropdown = DOMUtils.querySelector(".api-dropdown", widgetCard);
    const layerInput = DOMUtils.querySelector(".layer-input", widgetCard);
    const channelInput = DOMUtils.querySelector(".channel-input", widgetCard);
    const posXInput = DOMUtils.querySelector(".pos-x-input", widgetCard);
    const posYInput = DOMUtils.querySelector(".pos-y-input", widgetCard);
    const sizeXInput = DOMUtils.querySelector(".size-x-input", widgetCard);
    const sizeYInput = DOMUtils.querySelector(".size-y-input", widgetCard);

    const selectedTemplate = dropdown?.value;
    const layer = parseInt(layerInput?.value, 10) || 1;
    const channel = parseInt(channelInput?.value, 10) || 1;

    // These can be null if not provided
    const posX = posXInput?.value ? parseInt(posXInput.value, 10) : null;
    const posY = posYInput?.value ? parseInt(posYInput.value, 10) : null;
    const sizeX = sizeXInput?.value ? parseFloat(sizeXInput.value) : null;
    const sizeY = sizeYInput?.value ? parseFloat(sizeYInput.value) : null;

    if (!selectedTemplate) {
      console.error("No template selected for execution.");
      return;
    }

    const fieldRows = DOMUtils.querySelectorAll(
      `.${CSS_CLASSES.FIELD_ROW}`,
      widgetCard,
    );

    const data = {};
    fieldRows.forEach((row) => {
      const keyInput = DOMUtils.querySelector(SELECTORS.FIELD_KEY, row);
      const typeSelect = DOMUtils.querySelector(SELECTORS.FIELD_TYPE, row);
      const idInput = DOMUtils.querySelector(SELECTORS.FIELD_ID, row);

      if (!keyInput || !keyInput.value) return; // Skip if no key

      const key = keyInput.value;
      const type = typeSelect?.value || FIELD_TYPES.STRING;
      let value = idInput?.value || "";

      // Convert value based on type
      if (type === FIELD_TYPES.INT) {
        value = parseInt(value, 10) || 0;
      } else if (type === FIELD_TYPES.FLOAT) {
        value = parseFloat(value) || 0.0;
      }

      data[key] = value;
    });

    console.log(
      `Starting template on layer ${layer}, channel ${channel}, pos: (${posX}, ${posY}), size: (${sizeX}%, ${sizeY}%) with data:`,
      data,
    );
    APIService.pushCGData(
      selectedTemplate,
      layer,
      channel,
      data,
      posX,
      posY,
      sizeX,
      sizeY,
    );
  },
};

/**
 * Field Manager - handles custom field operations
 */
const FieldManager = {
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

    // Set source selection if provided
    if (source) {
      const sourceSelect = DOMUtils.querySelector(".f-source", row);
      if (sourceSelect) sourceSelect.value = source;
    }

    const deleteBtn = DOMUtils.querySelector(".delete-row-btn", row);
    if (deleteBtn) {
      deleteBtn.addEventListener("click", () => {
        row.remove();
        LayoutManager.scheduleAutoSave(
          AppState.grid,
          CSS_CLASSES,
          SELECTORS,
          FIELD_TYPES,
          DOMUtils,
        );
      });
    }

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
    const valueDisplay = DOMUtils.querySelector(
      SELECTORS.LIVE_VALUE_DISPLAY,
      row,
    );

    if (keyDisplay) keyDisplay.textContent = key;
    if (valueDisplay) {
      valueDisplay.textContent = "Fetching...";

      try {
        const liveData = await APIService.fetchLiveData(
          identifier,
          type,
          source,
        );
        valueDisplay.textContent = liveData;
      } catch (error) {
        console.error("Failed to fetch live data:", error);
        valueDisplay.textContent = "Error loading data";
      }
    }
  },
};

/**
 * Mode Manager - handles edit/live mode switching
 */
const ModeManager = {
  async toggleMode() {
    AppState.isLiveMode = !AppState.isLiveMode;
    const modeBtn = DOMUtils.querySelector(SELECTORS.TOGGLE_MODE_BTN);
    const addBtn = DOMUtils.querySelector(SELECTORS.ADD_WIDGET_BTN);

    if (!modeBtn || !addBtn) return;

    if (AppState.isLiveMode) {
      await this.enterLiveMode(modeBtn, addBtn);
    } else {
      this.enterEditMode(modeBtn, addBtn);
    }
  },

  async enterLiveMode(modeBtn, addBtn) {
    const sourceMap = new Map(); // sourceName -> [{Key, Type}]

    const fieldRows = DOMUtils.querySelectorAll(`.${CSS_CLASSES.FIELD_ROW}`);
    fieldRows.forEach((row) => {
      const idInput = DOMUtils.querySelector(SELECTORS.FIELD_ID, row);
      const typeSelect = DOMUtils.querySelector(SELECTORS.FIELD_TYPE, row);
      const sourceSelect = DOMUtils.querySelector(SELECTORS.FIELD_SOURCE, row);

      const locationKey = idInput?.value;
      const type = typeSelect?.value || FIELD_TYPES.STRING;
      const source = sourceSelect?.value;

      if (!locationKey || !source) return;

      if (!sourceMap.has(source)) sourceMap.set(source, []);
      sourceMap.get(source).push({ Key: locationKey, Type: type });
    });

    // Prime each data source with its locations
    await Promise.all(
      Array.from(sourceMap.entries()).map(([source, locations]) =>
        window.go.ui.UIService.PrimeDataSource(source, locations).catch((err) =>
          console.error(`Failed to prime data source '${source}':`, err),
        ),
      ),
    );

    AppState.grid.enableMove(false);
    AppState.grid.enableResize(false);
    document.body.classList.add(CSS_CLASSES.IS_LIVE);
    modeBtn.textContent = "Current: LIVE MODE";
    modeBtn.classList.add(CSS_CLASSES.MODE_LIVE);
    addBtn.style.display = "none";

    fieldRows.forEach((row) => FieldManager.updateLiveData(row));
  },

  enterEditMode(modeBtn, addBtn) {
    AppState.grid.enableMove(true);
    AppState.grid.enableResize(true);
    document.body.classList.remove(CSS_CLASSES.IS_LIVE);
    modeBtn.textContent = "Current: EDIT MODE";
    modeBtn.classList.remove(CSS_CLASSES.MODE_LIVE);
    addBtn.style.display = "block";
  },
};

/**
 * Application Initialization
 */
async function initializeApp() {
  // Initialize GridStack
  AppState.grid = GridStack.init({
    cellHeight: 100,
    margin: 10,
    float: true,
  });

  // Listen for grid changes (move, resize) and auto-save
  AppState.grid.on("change", () => {
    LayoutManager.scheduleAutoSave(
      AppState.grid,
      CSS_CLASSES,
      SELECTORS,
      FIELD_TYPES,
      DOMUtils,
    );
  });

  // Initialize event listeners
  initLiveEvents();

  // Load saved layout
  await LayoutManager.loadLayout(WidgetManager);

  // Attach button event listeners
  const addWidgetBtn = DOMUtils.querySelector(SELECTORS.ADD_WIDGET_BTN);
  if (addWidgetBtn) {
    addWidgetBtn.addEventListener("click", async () => {
      await WidgetManager.create();
      LayoutManager.scheduleAutoSave(
        AppState.grid,
        CSS_CLASSES,
        SELECTORS,
        FIELD_TYPES,
        DOMUtils,
      );
    });
  }

  const toggleModeBtn = DOMUtils.querySelector(SELECTORS.TOGGLE_MODE_BTN);
  if (toggleModeBtn) {
    toggleModeBtn.addEventListener("click", () => ModeManager.toggleMode());
  }
}

// Initialize the application when DOM is ready
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initializeApp);
} else {
  initializeApp();
}
