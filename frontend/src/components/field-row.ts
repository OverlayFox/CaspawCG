import { LitElement, html } from "lit";
import { property, state } from "lit/decorators.js";
import { types, ui } from "../../wailsjs/go/models";
import * as api from "../lib/api";
import { LIVE_DATA_EVENT, type LiveDataEvent } from "../lib/events";

export type FieldType = "string" | "int" | "float";
export type FieldInputType = "datasource" | "direct" | "range" | "schedule";

/**
 * One custom-field row inside a template card: maps a CasparCG template key to either a
 * data-source location, an offset into a data-source range, or a directly typed value.
 *
 * All parsing of the range/location string happens in Go (PrimeDataSources); this
 * component only collects raw text and displays whatever Go resolves it to.
 */
export class CaspFieldRow extends LitElement {
  @property({ type: String }) key = "";
  @property({ type: String }) fieldType: FieldType = "string";
  @property({ type: String }) inputType: FieldInputType = "datasource";
  @property({ type: String }) location = "";
  @property({ type: String }) source = "";
  @property({ type: String }) directValue = "";
  @property({ type: String }) range = "";
  @property({ type: Number }) offset = 0;

  @state() private dataSources: string[] = [];
  @state() private liveIdentifier: string | null = null;
  @state() private liveValue = "Loading...";

  protected override createRenderRoot() {
    return this;
  }

  override async connectedCallback() {
    super.connectedCallback();
    window.addEventListener(
      LIVE_DATA_EVENT,
      this.handleLiveData as EventListener,
    );
    this.dataSources = await api.getDataSources();
    const [firstSource] = this.dataSources;
    if (!this.source && firstSource) this.source = firstSource;
  }

  override disconnectedCallback() {
    window.removeEventListener(
      LIVE_DATA_EVENT,
      this.handleLiveData as EventListener,
    );
    super.disconnectedCallback();
  }

  private handleLiveData = (e: LiveDataEvent) => {
    if (
      this.liveIdentifier !== null &&
      e.detail.identifier === this.liveIdentifier
    ) {
      this.liveValue =
        typeof e.detail.value === "object"
          ? JSON.stringify(e.detail.value)
          : // The object case is already handled above; TS can't narrow the
            // `unknown` type of `value` through that typeof check, so this is
            // never actually stringifying a plain object.
            // eslint-disable-next-line @typescript-eslint/no-base-to-string
            String(e.detail.value);
    }
  };

  /**
   * Switches this row into live display mode. Direct-input rows show their own typed
   * value; datasource/range rows show the result of the enclosing app's batched
   * PrimeDataSources call (undefined only if that call itself failed outright).
   */
  enterLiveMode(result?: types.FieldSubscriptionResult) {
    if (this.inputType === "direct") {
      this.liveIdentifier = null;
      this.liveValue = this.directValue;
      return;
    }
    if (this.inputType === "schedule") {
      this.liveIdentifier = null;
      this.liveValue = `Schedule: ${this.range || "no range"}`;
      return;
    }
    this.liveIdentifier = result?.Identifier || null;
    if (result?.Error) {
      this.liveValue = result.Error;
    } else if (!this.liveIdentifier) {
      this.liveValue = "Offset out of range";
    } else {
      this.liveValue = String(result?.Value ?? "");
    }
  }

  exitLiveMode() {
    this.liveIdentifier = null;
  }

  /**
   * Returns null for DIRECT rows — they never resolve from a data source — and for
   * SCHEDULE rows, which have no backend resolution path yet.
   */
  buildSubscription(): types.FieldSubscription | null {
    if (this.inputType === "direct" || this.inputType === "schedule")
      return null;
    return types.FieldSubscription.createFrom({
      Source: this.source,
      Type: this.fieldType,
      InputType: this.inputType,
      Location: this.location,
      Range: this.range,
      Offset: this.offset,
    });
  }

  /** The current value to send to Go for a push/update, as raw text; Go coerces it. */
  currentRawValue(): string {
    return this.inputType === "direct" ? this.directValue : this.liveValue;
  }

  toConfig(): ui.FieldConfig {
    return ui.FieldConfig.createFrom({
      key: this.key,
      type: this.fieldType,
      inputType: this.inputType,
      location: this.inputType === "datasource" ? this.location : "",
      source: this.inputType === "direct" ? "" : this.source,
      value: this.inputType === "direct" ? this.directValue : "",
      range:
        this.inputType === "range" || this.inputType === "schedule"
          ? this.range
          : "",
      offset:
        this.inputType === "range" || this.inputType === "schedule"
          ? this.offset
          : 0,
    });
  }

  static fromConfig(config: ui.FieldConfig): CaspFieldRow {
    const row = document.createElement("casp-field-row") as CaspFieldRow;
    row.key = config.key || "";
    row.fieldType = (config.type as FieldType) || "string";
    row.inputType = (config.inputType as FieldInputType) || "datasource";
    row.location = config.location || "";
    row.source = config.source || "";
    row.directValue = config.value || "";
    row.range = config.range || "";
    row.offset = config.offset || 0;
    return row;
  }

  private onRemove = () => {
    this.dispatchEvent(new CustomEvent("casp-remove", { bubbles: true }));
    this.remove();
  };

  protected override render() {
    return html`
      <div class="field-row">
        <div class="edit-only field-row-edit">
          <input
            type="text"
            placeholder="Key"
            class="f-key"
            .value=${this.key}
            @input=${(e: Event) => (this.key = (e.target as HTMLInputElement).value)}
          />
          <select
            class="f-type"
            .value=${this.fieldType}
            @change=${(e: Event) => (this.fieldType = (e.target as HTMLSelectElement).value as FieldType)}
          >
            <option value="string">String</option>
            <option value="int">Int</option>
            <option value="float">Float</option>
          </select>
          <select
            class="f-input-type"
            .value=${this.inputType}
            @change=${(e: Event) => (this.inputType = (e.target as HTMLSelectElement).value as FieldInputType)}
          >
            <option value="datasource">Data Source</option>
            <option value="direct">Direct Input</option>
            <option value="range">Data Source Range</option>
            <option value="schedule">Schedule</option>
          </select>

          ${
            this.inputType === "datasource"
              ? html`
                  <div class="f-datasource-inputs">
                    <input
                      type="text"
                      placeholder="Location"
                      class="f-id"
                      .value=${this.location}
                      @input=${(e: Event) => (this.location = (e.target as HTMLInputElement).value)}
                    />
                    <select
                      class="f-source"
                      .value=${this.source}
                      @change=${(e: Event) => (this.source = (e.target as HTMLSelectElement).value)}
                    >
                      ${this.dataSources.map((s) => html`<option value=${s}>${s}</option>`)}
                    </select>
                  </div>
                `
              : ""
          }
          ${
            this.inputType === "direct"
              ? html`
                  <div class="f-direct-inputs">
                    <input
                      type="text"
                      placeholder="Value"
                      class="f-value"
                      .value=${this.directValue}
                      @input=${(e: Event) => (this.directValue = (e.target as HTMLInputElement).value)}
                    />
                  </div>
                `
              : ""
          }
          ${
            this.inputType === "range" || this.inputType === "schedule"
              ? html`
                  <div
                    class=${
                      this.inputType === "range"
                        ? "f-range-inputs"
                        : "f-schedule-inputs"
                    }
                  >
                    <input
                      type="text"
                      placeholder="Range e.g. Sheet1!A1:A10"
                      class="f-range"
                      .value=${this.range}
                      @input=${(e: Event) => (this.range = (e.target as HTMLInputElement).value)}
                    />
                    <select
                      class="f-source"
                      .value=${this.source}
                      @change=${(e: Event) => (this.source = (e.target as HTMLSelectElement).value)}
                    >
                      ${this.dataSources.map((s) => html`<option value=${s}>${s}</option>`)}
                    </select>
                    <input
                      type="number"
                      placeholder="Offset"
                      class="f-offset"
                      min="0"
                      .value=${String(this.offset)}
                      @input=${(e: Event) => (this.offset = parseInt((e.target as HTMLInputElement).value, 10) || 0)}
                    />
                  </div>
                `
              : ""
          }
          <button
            class="delete-row-btn"
            aria-label="Remove field"
            @click=${this.onRemove}
          >
            ❌
          </button>
        </div>
        <div class="live-only">
          <strong class="live-key-display">${this.key}</strong>:
          <span class="live-value-display">${this.liveValue}</span>
        </div>
      </div>
    `;
  }
}

customElements.define("casp-field-row", CaspFieldRow);
