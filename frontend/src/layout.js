/**
 * Layout Manager - handles saving and loading layout state
 */
export const LayoutManager = {
  saveTimeout: null,

  serializeLayout(grid, CSS_CLASSES, SELECTORS, FIELD_TYPES, DOMUtils) {
    const widgets = [];
    const gridItems = grid.getGridItems();

    gridItems.forEach((item) => {
      const widgetCard = DOMUtils.querySelector(
        `.${CSS_CLASSES.WIDGET_CARD}`,
        item,
      );
      if (!widgetCard) return;

      const node = item.gridstackNode;
      const widgetId =
        item.getAttribute("data-widget-id") ||
        `widget-${Date.now()}-${Math.random()}`;
      const dropdown = DOMUtils.querySelector(".api-dropdown", widgetCard);
      const layerInput = DOMUtils.querySelector(".layer-input", widgetCard);
      const channelInput = DOMUtils.querySelector(".channel-input", widgetCard);
      const posXInput = DOMUtils.querySelector(".pos-x-input", widgetCard);
      const posYInput = DOMUtils.querySelector(".pos-y-input", widgetCard);
      const sizeXInput = DOMUtils.querySelector(".size-x-input", widgetCard);
      const sizeYInput = DOMUtils.querySelector(".size-y-input", widgetCard);

      // Collect fields
      const fieldRows = DOMUtils.querySelectorAll(
        `.${CSS_CLASSES.FIELD_ROW}`,
        widgetCard,
      );
      const fields = [];
      fieldRows.forEach((row) => {
        const keyInput = DOMUtils.querySelector(SELECTORS.FIELD_KEY, row);
        const typeSelect = DOMUtils.querySelector(SELECTORS.FIELD_TYPE, row);
        const idInput = DOMUtils.querySelector(SELECTORS.FIELD_ID, row);
        const sourceSelect = DOMUtils.querySelector(
          SELECTORS.FIELD_SOURCE,
          row,
        );

        if (keyInput && keyInput.value) {
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
        posX: posXInput?.value ? parseInt(posXInput.value, 10) : null,
        posY: posYInput?.value ? parseInt(posYInput.value, 10) : null,
        sizeX: sizeXInput?.value ? parseFloat(sizeXInput.value) : null,
        sizeY: sizeYInput?.value ? parseFloat(sizeYInput.value) : null,
        fields: fields,
      });
    });

    return {
      version: 1,
      widgets: widgets,
    };
  },

  async saveLayout(grid, CSS_CLASSES, SELECTORS, FIELD_TYPES, DOMUtils) {
    try {
      const layout = this.serializeLayout(
        grid,
        CSS_CLASSES,
        SELECTORS,
        FIELD_TYPES,
        DOMUtils,
      );
      await window.go.ui.UIService.SaveLayout(layout);
      console.log("Layout saved successfully");
    } catch (error) {
      console.error("Failed to save layout:", error);
    }
  },

  async loadLayout(widgetManager) {
    try {
      const layout = await window.go.ui.UIService.LoadLayout();
      if (layout && layout.widgets && layout.widgets.length > 0) {
        console.log("Loading layout with", layout.widgets.length, "widgets");
        for (const widgetConfig of layout.widgets) {
          await widgetManager.createFromConfig(widgetConfig);
        }
      }
    } catch (error) {
      console.error("Failed to load layout:", error);
    }
  },

  scheduleAutoSave(grid, CSS_CLASSES, SELECTORS, FIELD_TYPES, DOMUtils) {
    // Debounce auto-save to avoid saving too frequently
    if (this.saveTimeout) {
      clearTimeout(this.saveTimeout);
    }
    this.saveTimeout = setTimeout(() => {
      this.saveLayout(grid, CSS_CLASSES, SELECTORS, FIELD_TYPES, DOMUtils);
    }, 1000);
  },
};
