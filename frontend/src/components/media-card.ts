import { LitElement, html } from "lit";
import { property, state } from "lit/decorators.js";
import { ui } from "../../wailsjs/go/models";
import * as api from "../lib/api";
import { connectionState } from "../lib/connection-state";
import { formatFileSize, formatFrameRate } from "../lib/format";

/** CaspMediaCard — a CasparCG media-file playback element: file/layer/channel/delay/loop. */
export class CaspMediaCard extends LitElement {
  @property({ type: String }) widgetId = String(Date.now());
  @property({ type: String }) widgetName = "Media Element";
  @property({ type: String }) filename = "";
  @property({ type: Number }) layer = 1;
  @property({ type: String }) channelExpr = "1";
  @property({ type: Number }) delayMs = 0;
  @property({ type: Boolean }) loop = false;

  @state() private mediaOptions: string[] = [];
  @state() private mediaInfoText: { filename: string; type: string; size: string; frames: number | string; frameRate: string } | null = null;
  @state() private error = "";

  protected override createRenderRoot() {
    return this;
  }

  override async connectedCallback() {
    super.connectedCallback();
    connectionState.subscribe(this.onConnectionData);
    this.mediaOptions = await connectionState.getMediaOptions();
    if (this.filename) await this.refreshMediaInfo();
  }

  override disconnectedCallback() {
    connectionState.unsubscribe(this.onConnectionData);
    super.disconnectedCallback();
  }

  private onConnectionData = (data: { media: string[] }) => {
    if (data.media.length) this.mediaOptions = data.media;
  };

  private async refreshMediaInfo() {
    if (!this.filename) {
      this.mediaInfoText = null;
      return;
    }
    try {
      const info = await api.getCasparCGMediaInfo(this.filename);
      this.mediaInfoText = {
        filename: info.Filename || "—",
        type: info.Type || "—",
        size: formatFileSize(info.FileSize),
        frames: info.FrameCount ?? "—",
        frameRate: formatFrameRate(info.FrameRate),
      };
    } catch (e) {
      console.error("Failed to fetch media info:", e);
      this.mediaInfoText = null;
    }
  }

  toConfig(position: { x: number; y: number; w: number; h: number }): ui.MediaWidgetConfig {
    return ui.MediaWidgetConfig.createFrom({
      id: this.widgetId,
      x: position.x,
      y: position.y,
      w: position.w,
      h: position.h,
      name: this.widgetName,
      filename: this.filename,
      layer: this.layer,
      channelExpr: this.channelExpr,
      delayMs: this.delayMs,
      loop: this.loop,
    });
  }

  static fromConfig(config: ui.MediaWidgetConfig): CaspMediaCard {
    const card = document.createElement("casp-media-card") as CaspMediaCard;
    card.widgetId = config.id || String(Date.now());
    card.widgetName = config.name || "Media Element";
    card.filename = config.filename || "";
    card.layer = config.layer || 1;
    card.channelExpr = config.channelExpr || "1";
    card.delayMs = config.delayMs || 0;
    card.loop = config.loop || false;
    return card;
  }

  toMediaGroupItem(): ui.MediaGroupItem {
    return ui.MediaGroupItem.createFrom({
      Filename: this.filename,
      Layer: this.layer,
      ChannelExpr: this.channelExpr,
      Loop: this.loop,
      DelayMs: this.delayMs,
    });
  }

  async play() {
    this.error = "";
    if (!this.filename) {
      this.error = "Select a media file first.";
      return;
    }
    try {
      await api.playMedia(this.filename, this.layer, this.channelExpr, this.loop, this.delayMs);
    } catch (e) {
      this.error = String(e);
    }
  }

  async stop() {
    this.error = "";
    try {
      await api.stopMedia(this.layer, this.channelExpr, this.delayMs);
    } catch (e) {
      this.error = String(e);
    }
  }

  private onRemove = () => {
    this.dispatchEvent(new CustomEvent("casp-remove", { bubbles: true, detail: { element: this } }));
  };

  protected override render() {
    const escapedName = this.widgetName.replace(/"/g, "&quot;");
    const onChange = () => this.dispatchEvent(new CustomEvent("casp-change", { bubbles: true }));

    return html`
      <div class="widget-header">
        <input
          type="text"
          class="widget-name-input edit-only"
          .value=${this.widgetName}
          placeholder="Element name"
          @input=${(e: Event) => {
            this.widgetName = (e.target as HTMLInputElement).value;
          }}
          @change=${onChange}
        />
        <span class="widget-name-display live-only">${escapedName}</span>
      </div>
      <div class="widget-header-controls">
        <select
          class="media-dropdown edit-only"
          .value=${this.filename}
          @change=${async (e: Event) => {
            this.filename = (e.target as HTMLSelectElement).value;
            await this.refreshMediaInfo();
            onChange();
          }}
        >
          ${this.mediaOptions.includes(this.filename) || !this.filename ? "" : html`<option value=${this.filename}>${this.filename}</option>`}
          ${this.mediaOptions.map((m) => html`<option value=${m}>${m}</option>`)}
        </select>
        <div class="input-group edit-only">
          <label>Layer:</label>
          <input
            type="number"
            class="layer-input"
            min="1"
            max="9999"
            .value=${String(this.layer)}
            @change=${(e: Event) => {
              this.layer = parseInt((e.target as HTMLInputElement).value, 10) || 1;
              onChange();
            }}
          />
        </div>
        <div class="input-group edit-only">
          <label>Channel:</label>
          <input
            type="text"
            class="channel-input"
            placeholder="e.g. 1 or 1,2 or 1-3"
            .value=${this.channelExpr}
            @change=${(e: Event) => {
              this.channelExpr = (e.target as HTMLInputElement).value;
              onChange();
            }}
          />
        </div>
        <button class="action-btn live-only" @click=${this.play}>Play</button>
        <button class="action-btn live-only" @click=${this.stop}>Stop</button>
        <button class="delete-btn edit-only" @click=${this.onRemove}>Remove</button>
      </div>
      <div class="widget-position-size-controls">
        <div class="input-group">
          <label>Delay (ms):</label>
          <input
            type="number"
            class="delay-input"
            min="0"
            max="60000"
            .value=${String(this.delayMs)}
            @change=${(e: Event) => {
              this.delayMs = parseInt((e.target as HTMLInputElement).value, 10) || 0;
              onChange();
            }}
          />
        </div>
        <div class="input-group">
          <label>Loop:</label>
          <input
            type="checkbox"
            class="loop-input"
            .checked=${this.loop}
            @change=${(e: Event) => {
              this.loop = (e.target as HTMLInputElement).checked;
              onChange();
            }}
          />
        </div>
      </div>
      ${this.error ? html`<div class="widget-error">${this.error}</div>` : ""}
      <div class="media-info-panel edit-only">
        ${this.mediaInfoText
          ? html`
              <div class="media-info-row"><span class="media-info-label">File:</span><span class="media-info-value">${this.mediaInfoText.filename}</span></div>
              <div class="media-info-row"><span class="media-info-label">Type:</span><span class="media-info-value">${this.mediaInfoText.type}</span></div>
              <div class="media-info-row"><span class="media-info-label">Size:</span><span class="media-info-value">${this.mediaInfoText.size}</span></div>
              <div class="media-info-row"><span class="media-info-label">Frames:</span><span class="media-info-value">${this.mediaInfoText.frames}</span></div>
              <div class="media-info-row">
                <span class="media-info-label">Frame rate:</span><span class="media-info-value">${this.mediaInfoText.frameRate}</span>
              </div>
            `
          : html`<span class="media-info-placeholder">Select a file to see details</span>`}
      </div>
    `;
  }
}

customElements.define("casp-media-card", CaspMediaCard);
