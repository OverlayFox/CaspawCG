import { APIService } from "./api.js";
import { DOMUtils } from "./dom-utils.js";
import { CSS_CLASSES, SELECTORS, FIELD_TYPES } from "./constants.js";
import { AppState } from "./state.js";
import { FieldManager } from "./field-manager.js";
import { LayoutManager } from "./layout.js";

/**
 * WidgetManager — creates and manages the dashboard widget cards.
 *
 * Each widget wraps a CasparCG template with a layer/channel assignment,
 * optional position/size overrides, and a set of custom field mappings
 * (managed by FieldManager).
 */
export const WidgetManager = {
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

    if (template) {
      const dropdown = DOMUtils.querySelector(".api-dropdown", gridItem);
      if (dropdown) dropdown.value = template;
    }

    if (config?.fields?.length > 0) {
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

    DOMUtils.querySelector(`.${CSS_CLASSES.DELETE_BTN}`, widgetCard)
      ?.addEventListener("click", () => {
        AppState.grid.removeWidget(gridItem);
        LayoutManager.scheduleAutoSave();
      });

    DOMUtils.querySelectorAll(`.${CSS_CLASSES.ACTION_BTN}`, widgetCard).forEach(
      (btn) => {
        btn.addEventListener("click", (e) => {
          if (e.target.dataset.action === "execute") {
            this.startWidgetAction(widgetCard);
          } else if (e.target.dataset.action === "stop") {
            this.stopWidgetAction(widgetCard);
          }
        });
      },
    );

    DOMUtils.querySelector(`.${CSS_CLASSES.ADD_FIELD_BTN}`, widgetCard)
      ?.addEventListener("click", () => {
        FieldManager.add(widgetCard);
        LayoutManager.scheduleAutoSave();
      });

    widgetCard.querySelectorAll("input, select").forEach((input) => {
      input.addEventListener("change", () => LayoutManager.scheduleAutoSave());
    });
  },

  stopWidgetAction(widgetCard) {
    const selectedTemplate = DOMUtils.querySelector(".api-dropdown", widgetCard)?.value;
    if (!selectedTemplate) {
      console.error("No template selected for stopping.");
      return;
    }

    const layer = parseInt(DOMUtils.querySelector(".layer-input", widgetCard)?.value, 10) || 1;
    const channel = parseInt(DOMUtils.querySelector(".channel-input", widgetCard)?.value, 10) || 1;

    console.log(`Stopping template on layer ${layer}, channel ${channel}: ${selectedTemplate}`);
    APIService.stopCGData(selectedTemplate, layer, channel);
  },

  async startWidgetAction(widgetCard) {
    const selectedTemplate = DOMUtils.querySelector(".api-dropdown", widgetCard)?.value;
    if (!selectedTemplate) {
      console.error("No template selected for execution.");
      return;
    }

    const layer = parseInt(DOMUtils.querySelector(".layer-input", widgetCard)?.value, 10) || 1;
    const channel = parseInt(DOMUtils.querySelector(".channel-input", widgetCard)?.value, 10) || 1;
    const posX = DOMUtils.querySelector(".pos-x-input", widgetCard)?.value
      ? parseInt(DOMUtils.querySelector(".pos-x-input", widgetCard).value, 10)
      : null;
    const posY = DOMUtils.querySelector(".pos-y-input", widgetCard)?.value
      ? parseInt(DOMUtils.querySelector(".pos-y-input", widgetCard).value, 10)
      : null;
    const sizeX = DOMUtils.querySelector(".size-x-input", widgetCard)?.value
      ? parseFloat(DOMUtils.querySelector(".size-x-input", widgetCard).value)
      : null;
    const sizeY = DOMUtils.querySelector(".size-y-input", widgetCard)?.value
      ? parseFloat(DOMUtils.querySelector(".size-y-input", widgetCard).value)
      : null;

    const data = {};
    const fetchPromises = [];

    DOMUtils.querySelectorAll(`.${CSS_CLASSES.FIELD_ROW}`, widgetCard).forEach(
      (row) => {
        const keyInput = DOMUtils.querySelector(SELECTORS.FIELD_KEY, row);
        if (!keyInput?.value) return;

        const key = keyInput.value;
        const type = DOMUtils.querySelector(SELECTORS.FIELD_TYPE, row)?.value || FIELD_TYPES.STRING;
        const identifier = DOMUtils.querySelector(SELECTORS.FIELD_ID, row)?.value || "";
        const source = DOMUtils.querySelector(SELECTORS.FIELD_SOURCE, row)?.value || "";

        if (identifier && source) {
          fetchPromises.push(
            APIService.fetchLiveData(identifier, type, source)
              .then((value) => { data[key] = value; })
              .catch((err) => {
                console.error(`Failed to fetch live data for key "${key}":`, err);
                data[key] = identifier;
              }),
          );
        } else {
          let value = identifier;
          if (type === FIELD_TYPES.INT) value = parseInt(value, 10) || 0;
          else if (type === FIELD_TYPES.FLOAT) value = parseFloat(value) || 0.0;
          data[key] = value;
        }
      },
    );

    await Promise.all(fetchPromises);

    console.log(
      `Starting template on layer ${layer}, channel ${channel}, pos: (${posX}, ${posY}), size: (${sizeX}%, ${sizeY}%) with data:`,
      data,
    );
    APIService.pushCGData(selectedTemplate, layer, channel, data, posX, posY, sizeX, sizeY);
  },
};
