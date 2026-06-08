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
import { parseChannelInput } from "./utils.js";
import { WidgetManager } from "./widget-manager.js";

function showConfirm(channels) {
  return new Promise((resolve) => {
    const modal = document.getElementById("confirm-modal");
    const okBtn = document.getElementById("confirm-modal-ok");
    const cancelBtn = document.getElementById("confirm-modal-cancel");
    const message = modal.querySelector(".modal-message");
    if (message) {
      message.textContent =
        channels === null
          ? "Are you sure you want to clear all Video Layers in CasparCG?"
          : `Are you sure you want to clear everything on channel${channels.length > 1 ? "s" : ""} ${channels.join(", ")}?`;
    }

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

      if (await showConfirm(channels)) {
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
