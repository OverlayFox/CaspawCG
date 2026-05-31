import { AppState } from "./state.js";
import { DOMUtils } from "./dom-utils.js";
import { CSS_CLASSES, SELECTORS } from "./constants.js";
import { FieldManager } from "./field-manager.js";

/**
 * ModeManager — switches the UI between Edit and Live mode.
 *
 * Edit mode: grid is draggable/resizable, inputs are visible.
 * Live mode: grid is locked, inputs are hidden, live data is displayed.
 */
export const ModeManager = {
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
    // Group subscribed locations by data source so we can prime each source once
    const sourceMap = new Map();

    DOMUtils.querySelectorAll(`.${CSS_CLASSES.FIELD_ROW}`).forEach((row) => {
      const locationKey = DOMUtils.querySelector(SELECTORS.FIELD_ID, row)?.value;
      const type = DOMUtils.querySelector(SELECTORS.FIELD_TYPE, row)?.value || "string";
      const source = DOMUtils.querySelector(SELECTORS.FIELD_SOURCE, row)?.value;

      if (!locationKey || !source) return;

      if (!sourceMap.has(source)) sourceMap.set(source, []);
      sourceMap.get(source).push({ Key: locationKey, Type: type });
    });

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

    DOMUtils.querySelectorAll(`.${CSS_CLASSES.FIELD_ROW}`).forEach((row) =>
      FieldManager.updateLiveData(row),
    );
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
