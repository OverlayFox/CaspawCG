import { GridStack } from "gridstack";
import "gridstack/dist/gridstack.min.css";

import { APIService } from "./api.js";
import { SELECTORS } from "./constants.js";
import { DOMUtils } from "./dom-utils.js";
import { initLiveEvents } from "./events.js";
import { GroupManager } from "./group-manager.js";
import { LayoutManager } from "./layout.js";
import { MediaWidgetManager } from "./media-widget-manager.js";
import { ModeManager } from "./mode-manager.js";
import { AppState } from "./state.js";
import { WidgetManager } from "./widget-manager.js";

// Parses "1", "1-3", "1,3-5" into a sorted, deduplicated array of integers.
// Returns null if the input is blank (meaning: all channels).
// Throws if any token is invalid.
function parseChannelInput(raw) {
  const trimmed = raw.trim();
  if (!trimmed) return null;

  const channels = new Set();

  for (const token of trimmed.split(",")) {
    const part = token.trim();
    const rangeMatch = part.match(/^(\d+)-(\d+)$/);
    if (rangeMatch) {
      const start = parseInt(rangeMatch[1], 10);
      const end = parseInt(rangeMatch[2], 10);
      if (start > end) throw new Error(`Invalid range "${part}"`);
      for (let i = start; i <= end; i++) channels.add(i);
    } else if (/^\d+$/.test(part)) {
      channels.add(parseInt(part, 10));
    } else {
      throw new Error(`Invalid token "${part}"`);
    }
  }

  return [...channels].sort((a, b) => a - b);
}

function showConfirm() {
  return new Promise((resolve) => {
    const modal = document.getElementById("confirm-modal");
    const okBtn = document.getElementById("confirm-modal-ok");
    const cancelBtn = document.getElementById("confirm-modal-cancel");

    const cleanup = (result) => {
      modal.hidden = true;
      okBtn.removeEventListener("click", onOk);
      cancelBtn.removeEventListener("click", onCancel);
      resolve(result);
    };

    const onOk = () => cleanup(true);
    const onCancel = () => cleanup(false);

    okBtn.addEventListener("click", onOk);
    cancelBtn.addEventListener("click", onCancel);
    modal.hidden = false;
    okBtn.focus();
  });
}

async function initializeApp() {
  AppState.grid = GridStack.init({
    cellHeight: 100,
    margin: 10,
    float: true,
  });

  AppState.grid.on("change", () => LayoutManager.scheduleAutoSave());

  initLiveEvents();

  LayoutManager.setGroupManager(GroupManager);
  LayoutManager.setMediaWidgetManager(MediaWidgetManager);
  await LayoutManager.loadLayout(WidgetManager, GroupManager);

  DOMUtils.querySelector(SELECTORS.ADD_WIDGET_BTN)?.addEventListener(
    "click",
    async () => {
      await WidgetManager.create();
      LayoutManager.scheduleAutoSave();
    },
  );

  DOMUtils.querySelector(SELECTORS.ADD_MEDIA_BTN)?.addEventListener(
    "click",
    async () => {
      await MediaWidgetManager.create();
      LayoutManager.scheduleAutoSave();
    },
  );

  DOMUtils.querySelector(SELECTORS.ADD_GROUP_BTN)?.addEventListener(
    "click",
    async () => {
      await GroupManager.create();
      LayoutManager.scheduleAutoSave();
    },
  );

  DOMUtils.querySelector(SELECTORS.TOGGLE_MODE_BTN)?.addEventListener(
    "click",
    () => ModeManager.toggleMode(),
  );

  DOMUtils.querySelector(SELECTORS.CLEAR_ALL_BTN)?.addEventListener(
    "click",
    async () => {
      const input = DOMUtils.querySelector(SELECTORS.CLEAR_CHANNELS_INPUT);
      let channels;
      try {
        channels = parseChannelInput(input?.value ?? "");
      } catch (e) {
        alert(`Invalid channel input: ${e.message}`);
        return;
      }

      if (await showConfirm()) {
        if (channels === null) {
          APIService.clearAll();
        } else {
          APIService.clearChannels(channels);
        }
      }
    },
  );
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initializeApp);
} else {
  initializeApp();
}
