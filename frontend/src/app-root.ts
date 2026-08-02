import type { GridItemHTMLElement } from "gridstack";
import { LitElement, html } from "lit";
import { state } from "lit/decorators.js";
import { types, ui } from "../wailsjs/go/models";
import "./components/casparcg-status-bar";
import { CaspConfirmModal } from "./components/confirm-modal";
import { CaspFieldRow } from "./components/field-row";
import { CaspGroupCard } from "./components/group-card";
import { CaspMediaCard } from "./components/media-card";
import { CaspTemplateCard } from "./components/template-card";
import * as api from "./lib/api";
import { initLiveEvents } from "./lib/events";
import { GridWrapper } from "./lib/gridstack-wrapper";

const AUTOSAVE_DELAY_MS = 1000;

/**
 * CaspApp — the top-level component: toolbar, GridStack instance, and load/save of
 * the whole layout. The only component that talks to GridStack (via GridWrapper) and
 * the only one that assembles the full LayoutConfig sent to Go.
 */
export class CaspApp extends LitElement {
  @state() private liveMode = false;

  private grid = new GridWrapper();
  private saveTimeout: number | null = null;

  protected override createRenderRoot() {
    return this;
  }

  override async firstUpdated() {
    this.grid.init(() => this.scheduleAutoSave());
    initLiveEvents();
    await this.loadLayout();
    this.addEventListener("casp-change", () => this.scheduleAutoSave());
    this.addEventListener("casp-remove", this.onRemove as EventListener);
  }

  private confirmModal(): CaspConfirmModal {
    return this.querySelector("casp-confirm-modal") as CaspConfirmModal;
  }

  // A top-level card/group asked to be removed. Nested removals (a template inside a
  // group) are handled by the group itself and stop propagation before reaching here.
  private onRemove = (e: CustomEvent<{ element: HTMLElement }>) => {
    const item = e.detail.element.closest(
      ".grid-stack-item",
    ) as GridItemHTMLElement | null;
    if (item) this.grid.removeItem(item);
    this.scheduleAutoSave();
  };

  private scheduleAutoSave() {
    if (this.saveTimeout !== null) window.clearTimeout(this.saveTimeout);
    this.saveTimeout = window.setTimeout(
      () => this.saveLayout(),
      AUTOSAVE_DELAY_MS,
    );
  }

  private async saveLayout() {
    const templates: ui.TemplateConfig[] = [];
    const mediaWidgets: ui.MediaWidgetConfig[] = [];
    const groups: ui.GroupConfig[] = [];

    for (const item of this.grid.getItems()) {
      const content = item.querySelector(".grid-stack-item-content > *");
      if (!content) continue;
      const position = this.grid.getPosition(item);

      if (content instanceof CaspTemplateCard)
        templates.push(content.toConfig(position));
      else if (content instanceof CaspMediaCard)
        mediaWidgets.push(content.toConfig(position));
      else if (content instanceof CaspGroupCard)
        groups.push(content.toConfig(position));
    }

    const layout = ui.LayoutConfig.createFrom({
      version: 2,
      widgets: templates,
      mediaWidgets: mediaWidgets.length ? mediaWidgets : undefined,
      groups: groups.length ? groups : undefined,
    });

    try {
      await api.saveLayout(layout);
    } catch (e) {
      console.error("Failed to save layout:", e);
    }
  }

  private async loadLayout() {
    let layout: ui.LayoutConfig;
    try {
      layout = await api.loadLayout();
    } catch (e) {
      console.error("Failed to load layout:", e);
      return;
    }

    for (const config of layout.widgets || []) {
      this.grid.addItem(CaspTemplateCard.fromConfig(config), {
        w: config.w || 4,
        h: config.h || 3,
        minW: 3,
        minH: 2,
        x: config.x,
        y: config.y,
      });
    }
    for (const config of layout.mediaWidgets || []) {
      this.grid.addItem(CaspMediaCard.fromConfig(config), {
        w: config.w || 3,
        h: config.h || 3,
        minW: 3,
        minH: 2,
        x: config.x,
        y: config.y,
      });
    }
    for (const config of layout.groups || []) {
      this.grid.addItem(CaspGroupCard.fromConfig(config), {
        w: config.w || 4,
        h: config.h || 3,
        minW: 3,
        minH: 3,
        x: config.x,
        y: config.y,
      });
    }
  }

  private addTemplate = () => {
    this.grid.addItem(document.createElement("casp-template-card"), {
      w: 4,
      h: 3,
      minW: 3,
      minH: 2,
    });
    this.scheduleAutoSave();
  };

  private addMedia = () => {
    this.grid.addItem(document.createElement("casp-media-card"), {
      w: 3,
      h: 3,
      minW: 3,
      minH: 2,
    });
    this.scheduleAutoSave();
  };

  private addGroup = () => {
    this.grid.addItem(document.createElement("casp-group-card"), {
      w: 4,
      h: 3,
      minW: 3,
      minH: 3,
    });
    this.scheduleAutoSave();
  };

  private toggleMode = async () => {
    if (this.liveMode) this.exitLiveMode();
    else await this.enterLiveMode();
  };

  private async enterLiveMode() {
    const rows = Array.from(
      document.querySelectorAll("casp-field-row"),
    ) as CaspFieldRow[];
    const subs: types.FieldSubscription[] = [];
    const subIndexByRow = new Map<CaspFieldRow, number>();
    for (const row of rows) {
      const sub = row.buildSubscription();
      if (sub) {
        subIndexByRow.set(row, subs.length);
        subs.push(sub);
      }
    }

    let results: types.FieldSubscriptionResult[] = [];
    if (subs.length) {
      try {
        results = await api.primeDataSources(subs);
      } catch (e) {
        console.error("Failed to prime data sources:", e);
      }
    }
    for (const row of rows) {
      const idx = subIndexByRow.get(row);
      row.enterLiveMode(idx !== undefined ? results[idx] : undefined);
    }

    this.liveMode = true;
    document.body.classList.add("is-live");
  }

  private exitLiveMode() {
    api.removeDataSources();

    this.liveMode = false;
    document.body.classList.remove("is-live");
    for (const row of document.querySelectorAll("casp-field-row"))
      (row as CaspFieldRow).exitLiveMode();
  }

  private clearChannels = async () => {
    const input = this.querySelector<HTMLInputElement>("#clear-channels-input");
    const channelExpr = input?.value ?? "";
    const message = channelExpr.trim()
      ? `Are you sure you want to clear everything on channel(s) ${channelExpr}?`
      : "Are you sure you want to clear all Video Layers in CasparCG?";

    if (!(await this.confirmModal().confirm(message))) return;

    try {
      if (channelExpr.trim()) await api.clearChannels(channelExpr);
      else await api.clearAll();
    } catch (e) {
      console.error("Failed to clear:", e);
      alert(`Failed to clear: ${e}`);
    }
  };

  protected override render() {
    return html`
      <div id="toolbar">
        <button
          id="add-template-btn"
          class="edit-only"
          @click=${this.addTemplate}
        >
          Add New Template
        </button>
        <button id="add-media-btn" class="edit-only" @click=${this.addMedia}>
          Add New Media
        </button>
        <button id="add-group-btn" class="edit-only" @click=${this.addGroup}>
          Add Group
        </button>
        <button
          id="toggle-mode-btn"
          class=${this.liveMode ? "mode-live" : "mode-edit"}
          @click=${this.toggleMode}
        >
          Current: ${this.liveMode ? "LIVE MODE" : "EDIT MODE"}
        </button>
        <input
          id="clear-channels-input"
          type="text"
          placeholder="e.g. 1, 1-3, 1,3-5"
          title="Channels to clear (leave blank for all)"
          style="margin-left: auto; width: 160px"
        />
        <button id="clear-all-btn" @click=${this.clearChannels}>Clear</button>
      </div>

      <div class="grid-stack"></div>

      <casp-confirm-modal></casp-confirm-modal>
      <div id="caspar-status-bar" class="status-bar">
        <casparcg-status-bar></casparcg-status-bar>
      </div>
    `;
  }
}

customElements.define("casp-app", CaspApp);
