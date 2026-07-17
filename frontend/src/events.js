/**
 * Event Constants
 */
const EVENT_TYPES = {
  LIVE_DATA_UPDATE: "live-data-update",
};

const SPECIAL_IDENTIFIERS = {
  CASPAR_KEEP_ALIVE: "CasparCGKeepAlive",
};

const CSS_CLASSES = {
  IS_LIVE: "is-live",
  FIELD_ROW: "field-row",
  STATUS_DOT: "status-dot",
  STATUS_ONLINE: "status-online",
  STATUS_OFFLINE: "status-offline",
  CLIENT_CHIP: "client-chip",
};

const SELECTORS = {
  FIELD_ID: ".f-id",
  LIVE_VALUE_DISPLAY: ".live-value-display",
  CASPAR_CLIENTS_CONTAINER: "#caspar-clients-container",
};

const ANIMATION_DURATION = 300;
const HIGHLIGHT_COLOR = "#2ecc71";

/**
 * Import ConnectionStateManager to handle reconnection logic
 */
import { ConnectionStateManager } from "./api.js";
import { FieldManager } from "./field-manager.js";

/**
 * DOM Utilities for event handling
 */
const EventDOMUtils = {
  querySelector(selector, parent = document) {
    return parent.querySelector(selector);
  },

  querySelectorAll(selector, parent = document) {
    return parent.querySelectorAll(selector);
  },

  createElement(tag, config = {}) {
    const element = document.createElement(tag);
    if (config.id) element.id = config.id;
    if (config.className) element.className = config.className;
    if (config.textContent) element.textContent = config.textContent;
    if (config.innerHTML) element.innerHTML = config.innerHTML;
    return element;
  },
};

/**
 * Data Update Handler - manages live data updates for field rows
 */
const DataUpdateHandler = {
  handle(data) {
    if (!this.isLiveMode()) {
      return;
    }

    const fieldRows = EventDOMUtils.querySelectorAll(`.${CSS_CLASSES.FIELD_ROW}`);
    fieldRows.forEach((row) => {
      let identifier;
      try {
        identifier = FieldManager.getLiveIdentifier(row);
      } catch {
        return;
      }

      if (identifier !== null && identifier === data.identifier) {
        this.updateFieldValue(row, data.value);
      }
    });
  },

  updateFieldValue(row, value) {
    const valueDisplay = EventDOMUtils.querySelector(
      SELECTORS.LIVE_VALUE_DISPLAY,
      row,
    );

    if (!valueDisplay) {
      console.warn("Value display element not found in field row");
      return;
    }

    try {
      // Handle different value types
      if (typeof value === "object" && value !== null) {
        valueDisplay.textContent = JSON.stringify(value, null, 2);
      } else {
        valueDisplay.textContent = value;
      }

      // Visual feedback animation
      this.highlightUpdate(valueDisplay);
    } catch (error) {
      console.error("Error updating field value:", error);
      valueDisplay.textContent = "Error displaying value";
    }
  },

  highlightUpdate(element) {
    element.style.color = HIGHLIGHT_COLOR;
    setTimeout(() => {
      element.style.color = "";
    }, ANIMATION_DURATION);
  },

  isLiveMode() {
    return document.body.classList.contains(CSS_CLASSES.IS_LIVE);
  },
};

/**
 * CasparCG Status Manager - handles CasparCG client status updates
 */
const CasparStatusManager = {
  _connectionStates: new Map(), // Track per-client connection state

  update(clientData) {
    if (!this.validateClientData(clientData)) {
      console.error("Invalid client data received:", clientData);
      return;
    }

    const { host, port, isAlive } = clientData;
    const container = EventDOMUtils.querySelector(
      SELECTORS.CASPAR_CLIENTS_CONTAINER,
    );

    if (!container) {
      console.warn("CasparCG clients container not found in DOM");
      return;
    }

    const clientId = this.generateClientId(host, port);
    const wasConnected = this._connectionStates.get(clientId) || false;
    this._connectionStates.set(clientId, isAlive);

    let chip = document.getElementById(clientId);

    if (!chip) {
      chip = this.createClientChip(clientId, host, port, isAlive);
      container.appendChild(chip);
    } else {
      this.updateClientStatus(chip, isAlive);
    }

    // Notify ConnectionStateManager of state change
    // (it will handle reconnection logic and data refresh)
    if (!wasConnected && isAlive) {
      ConnectionStateManager.handleConnectionChange(isAlive);
    } else if (wasConnected && !isAlive) {
      ConnectionStateManager.handleConnectionChange(isAlive);
    }
  },

  validateClientData(data) {
    return (
      data &&
      typeof data === "object" &&
      typeof data.host === "string" &&
      (typeof data.port === "number" || typeof data.port === "string") &&
      typeof data.isAlive === "boolean"
    );
  },

  generateClientId(host, port) {
    // Create a safe, unique HTML ID from host and port
    return `caspar-${host}-${port}`.replace(/[^a-zA-Z0-9-]/g, "-");
  },

  createClientChip(id, host, port, isAlive) {
    const chip = EventDOMUtils.createElement("div", {
      id,
      className: CSS_CLASSES.CLIENT_CHIP,
    });

    const dot = EventDOMUtils.createElement("div", {
      className: `${CSS_CLASSES.STATUS_DOT} ${
        isAlive ? CSS_CLASSES.STATUS_ONLINE : CSS_CLASSES.STATUS_OFFLINE
      }`,
    });

    const text = EventDOMUtils.createElement("span", {
      textContent: `${host}:${port}`,
    });

    chip.appendChild(dot);
    chip.appendChild(text);

    return chip;
  },

  updateClientStatus(chip, isAlive) {
    const dot = EventDOMUtils.querySelector(`.${CSS_CLASSES.STATUS_DOT}`, chip);

    if (!dot) {
      console.warn("Status dot not found in client chip");
      return;
    }

    if (isAlive) {
      dot.classList.remove(CSS_CLASSES.STATUS_OFFLINE);
      dot.classList.add(CSS_CLASSES.STATUS_ONLINE);
    } else {
      dot.classList.remove(CSS_CLASSES.STATUS_ONLINE);
      dot.classList.add(CSS_CLASSES.STATUS_OFFLINE);
    }
  },
};

/**
 * Event Router - routes events to appropriate handlers
 */
const EventRouter = {
  route(eventType, data) {
    try {
      if (eventType === EVENT_TYPES.LIVE_DATA_UPDATE) {
        // Handle special identifiers
        if (data.identifier === SPECIAL_IDENTIFIERS.CASPAR_KEEP_ALIVE) {
          CasparStatusManager.update(data.value);
        }

        // Handle regular field updates
        DataUpdateHandler.handle(data);
      } else {
        console.warn(`Unhandled event type: ${eventType}`);
      }
    } catch (error) {
      console.error(`Error routing event (${eventType}):`, error);
    }
  },
};

/**
 * Initialize live event listeners
 */
export function initLiveEvents() {
  console.log("Initializing live events listener...");

  if (!window.runtime || !window.runtime.EventsOn) {
    console.error(
      "Wails runtime not available. Event listeners cannot be initialized.",
    );
    return;
  }

  try {
    window.runtime.EventsOn(EVENT_TYPES.LIVE_DATA_UPDATE, (data) => {
      console.info("Live data update received:", data);
      EventRouter.route(EVENT_TYPES.LIVE_DATA_UPDATE, data);
    });

    console.log("Live events listener initialized successfully");
  } catch (error) {
    console.error("Failed to initialize live events:", error);
  }
}
