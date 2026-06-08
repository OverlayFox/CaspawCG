import { AppState } from "./state.js";
import { DOMUtils } from "./dom-utils.js";
import { CSS_CLASSES, SELECTORS, GROUP_CONTAINER_CLASS, FIELD_TYPES } from "./constants.js";

let _mediaWidgetManager = null;

/**
 * LayoutManager — saves and restores the grid layout to/from the backend.
 *
 * Receives WidgetManager and GroupManager as parameters (or via setter) to
 * avoid circular imports. Groups are serialized separately from top-level widgets.
 */

let _groupManager = null;

export const LayoutManager = {
  saveTimeout: null,

  setGroupManager(gm) {
    _groupManager = gm;
  },

  setMediaWidgetManager(mwm) {
    _mediaWidgetManager = mwm;
  },

  serializeLayout() {
    const grid = AppState.grid;
    const widgets = [];

    grid.getGridItems().forEach((item) => {
      // Group containers are serialized by GroupManager, not here.
      if (item.classList.contains(GROUP_CONTAINER_CLASS)) return;

      const widgetCard = DOMUtils.querySelector(`.${CSS_CLASSES.WIDGET_CARD}`, item);
      if (!widgetCard) return;

      const node = item.gridstackNode;
      const widgetId =
        item.getAttribute("data-widget-id") || `widget-${Date.now()}-${Math.random()}`;

      const dropdown = DOMUtils.querySelector(".api-dropdown", widgetCard);
      const layerInput = DOMUtils.querySelector(".layer-input", widgetCard);
      const channelInput = DOMUtils.querySelector(".channel-input", widgetCard);
      const posXInput = DOMUtils.querySelector(".pos-x-input", widgetCard);
      const posYInput = DOMUtils.querySelector(".pos-y-input", widgetCard);
      const sizeXInput = DOMUtils.querySelector(".size-x-input", widgetCard);
      const sizeYInput = DOMUtils.querySelector(".size-y-input", widgetCard);
      const delayInput = DOMUtils.querySelector(".delay-input", widgetCard);

      const fields = [];
      DOMUtils.querySelectorAll(`.${CSS_CLASSES.FIELD_ROW}`, widgetCard).forEach((row) => {
        const keyInput = DOMUtils.querySelector(SELECTORS.FIELD_KEY, row);
        const typeSelect = DOMUtils.querySelector(SELECTORS.FIELD_TYPE, row);
        const idInput = DOMUtils.querySelector(SELECTORS.FIELD_ID, row);
        const sourceSelect = DOMUtils.querySelector(SELECTORS.FIELD_SOURCE, row);

        if (keyInput?.value) {
          fields.push({
            key: keyInput.value,
            type: typeSelect?.value || FIELD_TYPES.STRING,
            id: idInput?.value || "",
            source: sourceSelect?.value || "",
          });
        }
      });

      widgets.push({
        id: widgetId,
        x: node.x,
        y: node.y,
        w: node.w,
        h: node.h,
        template: dropdown?.value || "",
        layer: parseInt(layerInput?.value, 10) || 1,
        channel: parseInt(channelInput?.value, 10) || 1,
        channelExpr: channelInput?.value || "1",
        posX: posXInput?.value ? parseInt(posXInput.value, 10) : null,
        posY: posYInput?.value ? parseInt(posYInput.value, 10) : null,
        sizeX: sizeXInput?.value ? parseFloat(sizeXInput.value) : null,
        sizeY: sizeYInput?.value ? parseFloat(sizeYInput.value) : null,
        delay: delayInput?.value ? parseInt(delayInput.value, 10) : 0,
        fields,
      });
    });

    const mediaWidgets = [];
    grid.getGridItems().forEach((item) => {
      if (item.classList.contains(GROUP_CONTAINER_CLASS)) return;

      const mediaCard = DOMUtils.querySelector(`.${CSS_CLASSES.MEDIA_WIDGET_CARD}`, item);
      if (!mediaCard) return;

      const node = item.gridstackNode;
      const widgetId =
        item.getAttribute("data-media-widget-id") || `media-${Date.now()}-${Math.random()}`;

      const dropdown = DOMUtils.querySelector(".media-dropdown", mediaCard);
      const layerInput = DOMUtils.querySelector(".layer-input", mediaCard);
      const channelInput = DOMUtils.querySelector(".channel-input", mediaCard);
      const delayInput = DOMUtils.querySelector(".delay-input", mediaCard);
      const loopInput = DOMUtils.querySelector(".loop-input", mediaCard);

      mediaWidgets.push({
        id: widgetId,
        x: node.x,
        y: node.y,
        w: node.w,
        h: node.h,
        filename: dropdown?.value || "",
        layer: parseInt(layerInput?.value, 10) || 1,
        channel: parseInt(channelInput?.value, 10) || 1,
        channelExpr: channelInput?.value || "1",
        delay: delayInput?.value ? parseInt(delayInput.value, 10) : 0,
        loop: loopInput?.checked ?? false,
      });
    });

    const groups = _groupManager ? _groupManager.serializeGroups() : [];
    return { version: 1, widgets, groups, mediaWidgets };
  },

  async saveLayout() {
    try {
      const layout = this.serializeLayout();
      await window.go.ui.UIService.SaveLayout(layout);
      console.log("Layout saved successfully");
    } catch (error) {
      console.error("Failed to save layout:", error);
    }
  },

  async loadLayout(widgetManager, groupManager) {
    try {
      const layout = await window.go.ui.UIService.LoadLayout();
      if (layout?.widgets?.length > 0) {
        for (const widgetConfig of layout.widgets) {
          await widgetManager.createFromConfig(widgetConfig);
        }
      }
      if (layout?.groups?.length > 0) {
        for (const groupConfig of layout.groups) {
          await groupManager.createFromConfig(groupConfig);
        }
      }
      if (layout?.mediaWidgets?.length > 0 && _mediaWidgetManager) {
        for (const mediaConfig of layout.mediaWidgets) {
          await _mediaWidgetManager.createFromConfig(mediaConfig);
        }
      }
    } catch (error) {
      console.error("Failed to load layout:", error);
    }
  },

  scheduleAutoSave() {
    if (this.saveTimeout) clearTimeout(this.saveTimeout);
    this.saveTimeout = setTimeout(() => this.saveLayout(), 1000);
  },
};
