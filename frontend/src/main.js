import { initLiveEvents } from "./events";

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

  async pushCGData(template, data) {
    try {
      await window.go.ui.UIService.PushCasparCGData(template, data);
    } catch (error) {
      console.error("Failed to push CG data:", error);
    }
  },

  async fetchLiveData(identifier, type, source) {
    // Future implementation:
    // return await window.go.ui.UIService.FetchData(identifier, type, source);

    // Temporary dummy data logic
    return new Promise((resolve) => {
      setTimeout(() => {
        let dummyData;
        if (type === FIELD_TYPES.INT) {
          dummyData = Math.floor(Math.random() * 1000);
        } else if (type === FIELD_TYPES.FLOAT) {
          dummyData = (Math.random() * 100).toFixed(2);
        } else {
          dummyData = `Data from ${source}`;
        }
        resolve(dummyData);
      }, 500);
    });
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
    const options = await APIService.getTemplateOptions();
    const optionsHtml = DOMUtils.createOptionsHTML(options);

    const widgetElement = `
      <div class="grid-stack-item">
        <div class="grid-stack-item-content ${CSS_CLASSES.WIDGET_CARD}">
          <div class="widget-header">
            <strong>Dynamic Element</strong>
          </div>
          <div class="widget-header-controls">
            <select class="api-dropdown ${CSS_CLASSES.EDIT_ONLY}">
              ${optionsHtml}
            </select>
            <button class="${CSS_CLASSES.ACTION_BTN}" data-action="execute">Execute</button>
            <button class="${CSS_CLASSES.DELETE_BTN} ${CSS_CLASSES.EDIT_ONLY}" data-action="remove">Remove</button>
          </div>
          <div class="${CSS_CLASSES.CUSTOM_FIELDS}"></div>
          <button class="${CSS_CLASSES.ADD_FIELD_BTN} ${CSS_CLASSES.EDIT_ONLY}">➕ Add Custom Field</button>
        </div>
      </div>
    `;

    const gridItem = AppState.grid.addWidget(widgetElement, { w: 3, h: 3 });
    this.attachEventListeners(gridItem);
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
      });
    }

    // Execute button handler
    const executeBtn = DOMUtils.querySelector(
      `.${CSS_CLASSES.ACTION_BTN}`,
      widgetCard,
    );
    if (executeBtn) {
      executeBtn.addEventListener("click", () => {
        this.executeWidgetAction(widgetCard);
      });
    }

    // Add field button handler
    const addFieldBtn = DOMUtils.querySelector(
      `.${CSS_CLASSES.ADD_FIELD_BTN}`,
      widgetCard,
    );
    if (addFieldBtn) {
      addFieldBtn.addEventListener("click", () => {
        FieldManager.add(widgetCard);
      });
    }
  },

  executeWidgetAction(widgetCard) {
    const dropdown = DOMUtils.querySelector(".api-dropdown", widgetCard);
    const selectedValue = dropdown?.value;

    if (!selectedValue) {
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

    console.log("Executing with data:", data);
    APIService.pushCGData(selectedValue, data);
  },
};

/**
 * Field Manager - handles custom field operations
 */
const FieldManager = {
  async add(widgetCard) {
    const container = DOMUtils.querySelector(
      `.${CSS_CLASSES.CUSTOM_FIELDS}`,
      widgetCard,
    );
    if (!container) return;

    const sources = await APIService.getDataSources();
    const sourcesHtml = DOMUtils.createOptionsHTML(sources);

    const row = DOMUtils.createElement("div", CSS_CLASSES.FIELD_ROW);
    row.innerHTML = `
      <div class="${CSS_CLASSES.EDIT_ONLY} field-row-edit">
        <input type="text" placeholder="Key" class="f-key">
        <select class="f-type">
          <option value="${FIELD_TYPES.STRING}">String</option>
          <option value="${FIELD_TYPES.INT}">Int</option>
          <option value="${FIELD_TYPES.FLOAT}">Float</option>
        </select>
        <input type="text" placeholder="Location" class="f-id">
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

    const deleteBtn = DOMUtils.querySelector(".delete-row-btn", row);
    if (deleteBtn) {
      deleteBtn.addEventListener("click", () => row.remove());
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
      this.enterLiveMode(modeBtn, addBtn);
    } else {
      this.enterEditMode(modeBtn, addBtn);
    }
  },

  enterLiveMode(modeBtn, addBtn) {
    AppState.grid.enableMove(false);
    AppState.grid.enableResize(false);
    document.body.classList.add(CSS_CLASSES.IS_LIVE);
    modeBtn.textContent = "Current: LIVE MODE";
    modeBtn.classList.add(CSS_CLASSES.MODE_LIVE);
    addBtn.style.display = "none";

    const fieldRows = DOMUtils.querySelectorAll(`.${CSS_CLASSES.FIELD_ROW}`);
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
function initializeApp() {
  // Initialize GridStack
  AppState.grid = GridStack.init({
    cellHeight: 100,
    margin: 10,
    float: true,
  });

  // Initialize event listeners
  initLiveEvents();

  // Attach button event listeners
  const addWidgetBtn = DOMUtils.querySelector(SELECTORS.ADD_WIDGET_BTN);
  if (addWidgetBtn) {
    addWidgetBtn.addEventListener("click", () => WidgetManager.create());
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
