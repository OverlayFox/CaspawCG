import { LitElement, html } from "lit";
import { property, state } from "lit/decorators.js";
import { types, ui } from "../../wailsjs/go/models";
import * as api from "../lib/api";
import { connectionState } from "../lib/connection-state";
import { CaspFieldRow } from "./field-row";

/**
 * CaspTemplateCard — a dynamic CasparCG template element: template/layer/channel,
 * position/size, delay/update-interval, and a list of custom fields. Holds an update
 * job UUID while a range-field live update is running.
 */
export class CaspTemplateCard extends LitElement {
  @property({ type: String }) widgetId = String(Date.now());
  @property({ type: String }) widgetName = "Dynamic Element";
  @property({ type: String }) template = "";
  @property({ type: Number }) layer = 1;
  @property({ type: String }) channelExpr = "1";
  @property({ type: Number }) posX = 0;
  @property({ type: Number }) posY = 0;
  @property({ type: Number }) sizeX = 100;
  @property({ type: Number }) sizeY = 100;
  @property({ type: Number }) delayMs = 0;
  @property({ type: Number }) updateIntervalMs = 0;
  @property({ type: Number }) scheduleMinElements = 0;
  @property({ type: String }) scheduleStartTimeColumn = "";
  @property({ type: String }) scheduleEndTimeColumn = "";

  @state() private templateOptions: string[] = [];
  @state() private error = "";

  private updateJobUuid: string | null = null;
  /** Field configs to hydrate into casp-field-row children once the first render has
   * created the .custom-fields-container they belong in (see firstUpdated). */
  private queuedFields: ui.FieldConfig[] = [];

  protected override createRenderRoot() {
    return this;
  }

  override async connectedCallback() {
    super.connectedCallback();
    connectionState.subscribe(this.onConnectionData);
    this.templateOptions = await connectionState.getTemplateOptions();
  }

  protected override firstUpdated() {
    const container = this.querySelector(".custom-fields-container");
    for (const field of this.queuedFields) {
      container?.appendChild(CaspFieldRow.fromConfig(field));
    }
    this.queuedFields = [];
  }

  override disconnectedCallback() {
    connectionState.unsubscribe(this.onConnectionData);
    super.disconnectedCallback();
  }

  private onConnectionData = (data: { templates: string[] }) => {
    if (data.templates.length) this.templateOptions = data.templates;
  };

  private fieldRows(): CaspFieldRow[] {
    return Array.from(this.querySelectorAll("casp-field-row"));
  }

  hasUpdateJob(): boolean {
    return this.updateJobUuid !== null;
  }

  async removeUpdateJobIfAny() {
    if (this.updateJobUuid) {
      await api
        .removeUpdateJob(this.updateJobUuid)
        .catch((e: unknown) =>
          console.error("Failed to remove update job:", e),
        );
      this.updateJobUuid = null;
    }
  }

  private addField = () => {
    const row = document.createElement("casp-field-row") as CaspFieldRow;
    this.querySelector(".custom-fields-container")?.appendChild(row);
    this.dispatchEvent(new CustomEvent("casp-change", { bubbles: true }));
  };

  toConfig(position: {
    x: number;
    y: number;
    w: number;
    h: number;
  }): ui.TemplateConfig {
    return ui.TemplateConfig.createFrom({
      id: this.widgetId,
      x: position.x,
      y: position.y,
      w: position.w,
      h: position.h,
      name: this.widgetName,
      template: this.template,
      channelExpr: this.channelExpr,
      layer: this.layer,
      sizing: types.Sizing.createFrom({
        posX: this.posX,
        posY: this.posY,
        sizeX: this.sizeX,
        sizeY: this.sizeY,
      }),
      delayMs: this.delayMs,
      updateIntervalMs: this.updateIntervalMs,
      scheduleMinElements: this.scheduleMinElements,
      scheduleStartTimeColumn: this.scheduleStartTimeColumn,
      scheduleEndTimeColumn: this.scheduleEndTimeColumn,
      fields: this.fieldRows().map((row) => row.toConfig()),
    });
  }

  static fromConfig(config: ui.TemplateConfig): CaspTemplateCard {
    const card = document.createElement(
      "casp-template-card",
    ) as CaspTemplateCard;
    card.widgetId = config.id || String(Date.now());
    card.widgetName = config.name || "Dynamic Element";
    card.template = config.template || "";
    card.channelExpr = config.channelExpr || "1";
    card.layer = config.layer || 1;
    card.posX = config.sizing?.posX ?? 0;
    card.posY = config.sizing?.posY ?? 0;
    card.sizeX = config.sizing?.sizeX ?? 100;
    card.sizeY = config.sizing?.sizeY ?? 100;
    card.delayMs = config.delayMs || 0;
    card.updateIntervalMs = config.updateIntervalMs || 0;
    card.scheduleMinElements = config.scheduleMinElements || 0;
    card.scheduleStartTimeColumn = config.scheduleStartTimeColumn || "";
    card.scheduleEndTimeColumn = config.scheduleEndTimeColumn || "";
    card.queuedFields = config.fields || [];
    return card;
  }

  /** Builds the request Go needs to push/stop/next/update this template. Returns null
   * and sets a visible error if fields resolve to a field with no key. */
  private literalFields(): types.LiteralField[] {
    return this.fieldRows()
      .filter(
        (row) =>
          row.inputType !== "range" && row.inputType !== "schedule" && row.key,
      )
      .map((row) =>
        types.LiteralField.createFrom({
          CasparKey: row.key,
          Type: row.fieldType,
          RawValue: row.currentRawValue(),
        }),
      );
  }

  private rangeFields(): ui.RangeField[] {
    return this.fieldRows()
      .filter((row) => row.inputType === "range" && row.key)
      .map((row) =>
        ui.RangeField.createFrom({
          CasparKey: row.key,
          Type: row.fieldType,
          Source: row.source,
          Range: row.range,
          Offset: row.offset,
        }),
      );
  }

  private scheduleFields(): ui.RangeField[] {
    return this.fieldRows()
      .filter((row) => row.inputType === "schedule" && row.key)
      .map((row) =>
        ui.RangeField.createFrom({
          CasparKey: row.key,
          Type: row.fieldType,
          Source: row.source,
          Range: row.range,
          Offset: row.offset,
        }),
      );
  }

  private sizing(): types.Sizing {
    return types.Sizing.createFrom({
      posX: this.posX,
      posY: this.posY,
      sizeX: this.sizeX,
      sizeY: this.sizeY,
    });
  }

  /** Builds this template's contribution to a group push/stop/next call, or null if it
   * has no template selected. Range fields aren't supported in a group push (there's
   * no batched equivalent of an update job) so only literal fields are included. */
  toGroupItemOrNull(): ui.CGDataGroup | null {
    if (!this.template) return null;
    return ui.CGDataGroup.createFrom({
      Template: this.template,
      Layer: this.layer,
      ChannelExpr: this.channelExpr,
      Fields: this.literalFields(),
      Sizing: this.sizing(),
      DelayMs: this.delayMs,
    });
  }

  async execute() {
    this.error = "";
    if (!this.template) {
      this.error = "Select a template first.";
      return;
    }
    try {
      const scheduleFields = this.scheduleFields();
      const rangeFields = this.rangeFields();
      if (this.updateIntervalMs === 0) {
        await this.removeUpdateJobIfAny();
      }
      if (scheduleFields.length > 0 && this.updateIntervalMs > 0) {
        await this.removeUpdateJobIfAny();
        this.updateJobUuid = await api.scheduleCGData(
          this.template,
          this.layer,
          this.channelExpr,
          this.literalFields(),
          scheduleFields,
          this.sizing(),
          this.delayMs,
          this.updateIntervalMs,
          this.scheduleMinElements,
          this.scheduleStartTimeColumn,
          this.scheduleEndTimeColumn,
        );
        return;
      }
      if (rangeFields.length > 0 && this.updateIntervalMs > 0) {
        await this.removeUpdateJobIfAny();
        this.updateJobUuid = await api.updateCGData(
          this.template,
          this.layer,
          this.channelExpr,
          this.literalFields(),
          rangeFields,
          this.sizing(),
          this.delayMs,
          this.updateIntervalMs,
        );
        return;
      }
      await api.pushCGData(
        this.template,
        this.layer,
        this.channelExpr,
        this.literalFields(),
        this.sizing(),
        this.delayMs,
      );
    } catch (e) {
      this.error = String(e);
    }
  }

  async stop() {
    this.error = "";
    if (!this.template) return;
    try {
      await api.stopCGData(
        this.template,
        this.layer,
        this.channelExpr,
        this.delayMs,
      );
      await this.removeUpdateJobIfAny();
    } catch (e) {
      this.error = String(e);
    }
  }

  async next() {
    this.error = "";
    if (!this.template) return;
    try {
      await api.nextCGData(
        this.template,
        this.layer,
        this.channelExpr,
        this.delayMs,
      );
    } catch (e) {
      this.error = String(e);
    }
  }

  // Dispatches a bubbling request to be removed; the enclosing container (app-root for
  // a top-level grid item, group-card for a group-nested entry) decides how — a
  // top-level item must go through GridStack's removeWidget, a nested entry can just
  // be detached from the DOM. This component never removes its own grid wrapper.
  private onRemove = async () => {
    await this.removeUpdateJobIfAny();
    this.dispatchEvent(
      new CustomEvent("casp-remove", {
        bubbles: true,
        detail: { element: this },
      }),
    );
  };

  protected override render() {
    const escapedName = this.widgetName.replace(/"/g, "&quot;");
    const onChange = () =>
      this.dispatchEvent(new CustomEvent("casp-change", { bubbles: true }));

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
          class="api-dropdown edit-only"
          .value=${this.template}
          @change=${(e: Event) => {
            this.template = (e.target as HTMLSelectElement).value;
            onChange();
          }}
        >
          ${
            this.templateOptions.includes(this.template) || !this.template
              ? ""
              : html`<option value=${this.template}>${this.template}</option>`
          }
          ${this.templateOptions.map((t) => html`<option value=${t}>${t}</option>`)}
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
              this.layer =
                parseInt((e.target as HTMLInputElement).value, 10) || 1;
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
        <button
          class="action-btn live-only"
          data-action="execute"
          @click=${this.execute}
        >
          Execute
        </button>
        <button class="action-btn live-only" @click=${this.next}>Next</button>
        <button
          class="action-btn live-only"
          data-action="stop"
          @click=${this.stop}
        >
          Stop
        </button>
        <button class="delete-btn edit-only" @click=${this.onRemove}>
          Remove
        </button>
      </div>
      <div class="widget-position-size-controls">
        <div class="widget-controls-row">
          <div class="input-group">
            <label>Pos X (px):</label>
            <input
              type="number"
              class="pos-x-input"
              min="0"
              max="1920"
              .value=${String(this.posX)}
              @change=${(e: Event) => {
                this.posX =
                  parseInt((e.target as HTMLInputElement).value, 10) || 0;
                onChange();
              }}
            />
          </div>
          <div class="input-group">
            <label>Pos Y (px):</label>
            <input
              type="number"
              class="pos-y-input"
              min="0"
              max="1080"
              .value=${String(this.posY)}
              @change=${(e: Event) => {
                this.posY =
                  parseInt((e.target as HTMLInputElement).value, 10) || 0;
                onChange();
              }}
            />
          </div>
        </div>
        <div class="widget-controls-row">
          <div class="input-group">
            <label>Size X (%):</label>
            <input
              type="number"
              class="size-x-input"
              min="0"
              max="100"
              .value=${String(this.sizeX)}
              @change=${(e: Event) => {
                this.sizeX =
                  parseFloat((e.target as HTMLInputElement).value) || 100;
                onChange();
              }}
            />
          </div>
          <div class="input-group">
            <label>Size Y (%):</label>
            <input
              type="number"
              class="size-y-input"
              min="0"
              max="100"
              .value=${String(this.sizeY)}
              @change=${(e: Event) => {
                this.sizeY =
                  parseFloat((e.target as HTMLInputElement).value) || 100;
                onChange();
              }}
            />
          </div>
        </div>
        <div class="widget-controls-row">
          <div class="input-group">
            <label>Delay (ms):</label>
            <input
              type="number"
              class="delay-input"
              min="0"
              max="60000"
              .value=${String(this.delayMs)}
              @change=${(e: Event) => {
                this.delayMs =
                  parseInt((e.target as HTMLInputElement).value, 10) || 0;
                onChange();
              }}
            />
          </div>
          <div class="input-group">
            <label>Update Interval (ms):</label>
            <input
              type="number"
              class="update-interval-input"
              min="0"
              max="600000"
              .value=${String(this.updateIntervalMs)}
              @change=${(e: Event) => {
                this.updateIntervalMs =
                  parseInt((e.target as HTMLInputElement).value, 10) || 0;
                onChange();
              }}
            />
          </div>
        </div>
        <div class="widget-controls-row edit-only">
          <div class="input-group">
            <label>Schedule Min Elements:</label>
            <input
              type="number"
              class="schedule-min-elements-input"
              min="0"
              .value=${String(this.scheduleMinElements)}
              @change=${(e: Event) => {
                this.scheduleMinElements =
                  parseInt((e.target as HTMLInputElement).value, 10) || 0;
                onChange();
              }}
            />
          </div>
          <div class="input-group">
            <label>Start Time Column:</label>
            <input
              type="text"
              class="schedule-start-time-column-input"
              maxlength="1"
              .value=${this.scheduleStartTimeColumn}
              @input=${(e: Event) => {
                const v = (e.target as HTMLInputElement).value.toUpperCase();
                this.scheduleStartTimeColumn = /^[A-Z]$/.test(v) ? v : "";
                onChange();
              }}
            />
          </div>
          <div class="input-group">
            <label>End Time Column:</label>
            <input
              type="text"
              class="schedule-end-time-column-input"
              maxlength="1"
              .value=${this.scheduleEndTimeColumn}
              @input=${(e: Event) => {
                const v = (e.target as HTMLInputElement).value.toUpperCase();
                this.scheduleEndTimeColumn = /^[A-Z]$/.test(v) ? v : "";
                onChange();
              }}
            />
          </div>
        </div>
      </div>
      ${this.error ? html`<div class="widget-error">${this.error}</div>` : ""}
      <div class="custom-fields-container"></div>
      <button class="add-field-btn edit-only" @click=${this.addField}>
        ➕ Add Custom Field
      </button>
    `;
  }
}

customElements.define("casp-template-card", CaspTemplateCard);
