import { GridStack } from "gridstack";
import "gridstack/dist/gridstack.min.css";

import { AppState } from "./state.js";
import { DOMUtils } from "./dom-utils.js";
import { SELECTORS } from "./constants.js";
import { LayoutManager } from "./layout.js";
import { WidgetManager } from "./widget-manager.js";
import { ModeManager } from "./mode-manager.js";
import { initLiveEvents } from "./events.js";

async function initializeApp() {
  AppState.grid = GridStack.init({
    cellHeight: 100,
    margin: 10,
    float: true,
  });

  AppState.grid.on("change", () => LayoutManager.scheduleAutoSave());

  initLiveEvents();

  await LayoutManager.loadLayout(WidgetManager);

  DOMUtils.querySelector(SELECTORS.ADD_WIDGET_BTN)?.addEventListener(
    "click",
    async () => {
      await WidgetManager.create();
      LayoutManager.scheduleAutoSave();
    },
  );

  DOMUtils.querySelector(SELECTORS.TOGGLE_MODE_BTN)?.addEventListener(
    "click",
    () => ModeManager.toggleMode(),
  );
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initializeApp);
} else {
  initializeApp();
}
