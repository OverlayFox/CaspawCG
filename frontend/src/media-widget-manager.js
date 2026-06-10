import { APIService } from "./api.js";
import { CSS_CLASSES } from "./constants.js";
import { DOMUtils } from "./dom-utils.js";
import { LayoutManager } from "./layout.js";
import { AppState } from "./state.js";
import { parseChannelInput } from "./utils.js";

export const MediaWidgetManager = {
  async create() {
    return this.createFromConfig(null);
  },

  _formatFileSize(bytes) {
    if (!bytes) return "—";
    if (bytes >= 1_000_000) return `${(bytes / 1_000_000).toFixed(1)} MB`;
    if (bytes >= 1_000) return `${(bytes / 1_000).toFixed(1)} KB`;
    return `${bytes} B`;
  },

  _buildInnerCardHTML(config, optionsHtml) {
    const filename = config?.filename || "";
    const layer = config?.layer || 1;
    const channel = config?.channelExpr || config?.channel || 1;
    const mediaName = config?.name || "Media Element";
    const escapedName = mediaName.replace(/"/g, "&quot;");

    return `
      <div class="widget-header">
        <input type="text" class="widget-name-input ${CSS_CLASSES.EDIT_ONLY}" value="${escapedName}" placeholder="Element name">
        <span class="widget-name-display ${CSS_CLASSES.LIVE_ONLY}">${mediaName}</span>
      </div>
      <div class="widget-header-controls">
        <select class="media-dropdown ${CSS_CLASSES.EDIT_ONLY}">
          ${optionsHtml}
        </select>
        <div class="input-group ${CSS_CLASSES.EDIT_ONLY}">
          <label>Layer:</label>
          <input type="number" class="layer-input" min="1" max="9999" value="${layer}">
        </div>
        <div class="input-group ${CSS_CLASSES.EDIT_ONLY}">
          <label>Channel:</label>
          <input type="text" class="channel-input" placeholder="e.g. 1 or 1,2 or 1-3" value="${channel}">
        </div>
        <button class="${CSS_CLASSES.ACTION_BTN} ${CSS_CLASSES.LIVE_ONLY}" data-action="play">Play</button>
        <button class="${CSS_CLASSES.ACTION_BTN} ${CSS_CLASSES.LIVE_ONLY}" data-action="stop">Stop</button>
        <button class="${CSS_CLASSES.DELETE_BTN} ${CSS_CLASSES.EDIT_ONLY}" data-action="remove">Remove</button>
      </div>
      <div class="widget-position-size-controls">
        <div class="input-group">
          <label>Delay (ms):</label>
          <input type="number" class="delay-input" min="0" max="60000" value="${config?.delay || 0}">
        </div>
        <div class="input-group">
          <label>Loop:</label>
          <input type="checkbox" class="loop-input" ${config?.loop ? "checked" : ""}>
        </div>
      </div>
      <div class="${CSS_CLASSES.MEDIA_INFO_PANEL} ${CSS_CLASSES.EDIT_ONLY}">
        <span class="media-info-placeholder">Select a file to see details</span>
      </div>
    `;
  },

  _renderMediaInfo(panel, info) {
    if (!info || (!info.Filename && !info.filename)) {
      panel.innerHTML = `<span class="media-info-placeholder">No info available</span>`;
      return;
    }

    const filename = info.Filename || info.filename || "—";
    const type = info.Type || info.type || "—";
    const fileSize = this._formatFileSize(info.FileSize || info.filesize);
    const frameCount = info.FrameCount ?? info.framecount ?? "—";
    const frameRate = info.FrameRate || info.framerate;
    const frameRateStr = frameRate
      ? `${frameRate.Num ?? frameRate.num}/${frameRate.Den ?? frameRate.den}`
      : "—";

    panel.innerHTML = `
      <div class="media-info-row"><span class="media-info-label">File:</span><span class="media-info-value">${filename}</span></div>
      <div class="media-info-row"><span class="media-info-label">Type:</span><span class="media-info-value">${type}</span></div>
      <div class="media-info-row"><span class="media-info-label">Size:</span><span class="media-info-value">${fileSize}</span></div>
      <div class="media-info-row"><span class="media-info-label">Frames:</span><span class="media-info-value">${frameCount}</span></div>
      <div class="media-info-row"><span class="media-info-label">Frame rate:</span><span class="media-info-value">${frameRateStr}</span></div>
    `;
  },

  async _restoreCardState(mediaCard, config) {
    if (config?.filename) {
      const dropdown = DOMUtils.querySelector(".media-dropdown", mediaCard);
      if (dropdown) dropdown.value = config.filename;

      const panel = DOMUtils.querySelector(
        `.${CSS_CLASSES.MEDIA_INFO_PANEL}`,
        mediaCard,
      );
      if (panel) {
        const info = await APIService.getMediaInfo(config.filename);
        this._renderMediaInfo(panel, info);
      }
    }
  },

  _attachCardListeners(mediaCard, onRemove) {
    const nameInput = mediaCard.querySelector(".widget-name-input");
    const nameDisplay = mediaCard.querySelector(".widget-name-display");

    nameInput?.addEventListener("input", () => {
      if (nameDisplay) nameDisplay.textContent = nameInput.value;
      LayoutManager.scheduleAutoSave();
    });

    DOMUtils.querySelector(
      `.${CSS_CLASSES.DELETE_BTN}`,
      mediaCard,
    )?.addEventListener("click", onRemove);

    DOMUtils.querySelectorAll(`.${CSS_CLASSES.ACTION_BTN}`, mediaCard).forEach(
      (btn) => {
        btn.addEventListener("click", (e) => {
          if (e.target.dataset.action === "play")
            this.playMediaAction(mediaCard);
          else if (e.target.dataset.action === "stop")
            this.stopMediaAction(mediaCard);
        });
      },
    );

    const dropdown = DOMUtils.querySelector(".media-dropdown", mediaCard);
    if (dropdown) {
      dropdown.addEventListener("change", async () => {
        const panel = DOMUtils.querySelector(
          `.${CSS_CLASSES.MEDIA_INFO_PANEL}`,
          mediaCard,
        );
        if (panel && dropdown.value) {
          const info = await APIService.getMediaInfo(dropdown.value);
          this._renderMediaInfo(panel, info);
        }
        LayoutManager.scheduleAutoSave();
      });
    }

    mediaCard.querySelectorAll("input").forEach((input) => {
      input.addEventListener("change", () => LayoutManager.scheduleAutoSave());
      input.addEventListener("input", () => LayoutManager.scheduleAutoSave());
    });
  },

  async createFromConfig(config = null) {
    const options = await APIService.getMediaOptions();
    const optionsHtml = DOMUtils.createOptionsHTML(options);
    const widgetId = config?.id || Date.now();

    const widgetElement = `
      <div class="grid-stack-item" data-media-widget-id="${widgetId}">
        <div class="grid-stack-item-content ${CSS_CLASSES.MEDIA_WIDGET_CARD}">
          ${this._buildInnerCardHTML(config, optionsHtml)}
        </div>
      </div>
    `;

    const gridOptions = {
      w: config?.w || 5,
      h: config?.h || 2,
      minW: 5,
      minH: 2,
    };
    if (config) {
      gridOptions.x = config.x;
      gridOptions.y = config.y;
    }

    const gridItem = AppState.grid.addWidget(widgetElement, gridOptions);
    const mediaCard = DOMUtils.querySelector(
      `.${CSS_CLASSES.MEDIA_WIDGET_CARD}`,
      gridItem,
    );

    await this._restoreCardState(mediaCard, config);

    this._attachCardListeners(mediaCard, () => {
      AppState.grid.removeWidget(gridItem);
      LayoutManager.scheduleAutoSave();
    });

    return gridItem;
  },

  stopMediaAction(mediaCard) {
    const layer =
      parseInt(DOMUtils.querySelector(".layer-input", mediaCard)?.value, 10) ||
      1;
    const delayVal = DOMUtils.querySelector(".delay-input", mediaCard)?.value;
    const delay = delayVal ? parseInt(delayVal, 10) * 1_000_000 : 0;

    let channels;
    try {
      channels = parseChannelInput(
        DOMUtils.querySelector(".channel-input", mediaCard)?.value ?? "1",
      ) || [1];
    } catch (e) {
      alert(`Invalid channel input: ${e.message}`);
      return;
    }

    APIService.stopMedia(layer, channels, delay);
  },

  playMediaAction(mediaCard) {
    const filename = DOMUtils.querySelector(
      ".media-dropdown",
      mediaCard,
    )?.value;
    if (!filename) {
      console.error("No media file selected for playback.");
      return;
    }

    const layer =
      parseInt(DOMUtils.querySelector(".layer-input", mediaCard)?.value, 10) ||
      1;
    const loop =
      DOMUtils.querySelector(".loop-input", mediaCard)?.checked ?? false;
    const delayVal = DOMUtils.querySelector(".delay-input", mediaCard)?.value;
    const delay = delayVal ? parseInt(delayVal, 10) * 1_000_000 : 0;

    let channels;
    try {
      channels = parseChannelInput(
        DOMUtils.querySelector(".channel-input", mediaCard)?.value ?? "1",
      ) || [1];
    } catch (e) {
      alert(`Invalid channel input: ${e.message}`);
      return;
    }

    APIService.playMedia(filename, layer, channels, loop, delay);
  },
};
