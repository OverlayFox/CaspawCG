import { LitElement, html } from "lit";
import { property, state } from "lit/decorators.js";
import { ui } from "../../wailsjs/go/models";
import * as api from "../lib/api";
import { CaspMediaCard } from "./media-card";
import { CaspTemplateCard } from "./template-card";

/**
 * CaspGroupCard — a container holding multiple template/media cards, with Execute/Next/
 * Stop-All buttons that batch-call every child at once via the *Group Go endpoints.
 */
export class CaspGroupCard extends LitElement {
  @property({ type: String }) groupId = `group-${Date.now()}`;
  @property({ type: String }) groupName = "New Group";

  @state() private error = "";
  private queuedTemplates: ui.TemplateConfig[] = [];
  private queuedMediaWidgets: ui.MediaWidgetConfig[] = [];

  protected override createRenderRoot() {
    return this;
  }

  protected override firstUpdated() {
    const list = this.widgetsList();
    for (const t of this.queuedTemplates)
      list?.appendChild(CaspTemplateCard.fromConfig(t));
    for (const m of this.queuedMediaWidgets)
      list?.appendChild(CaspMediaCard.fromConfig(m));
    this.queuedTemplates = [];
    this.queuedMediaWidgets = [];
  }

  private widgetsList(): HTMLElement | null {
    return this.querySelector(".group-widgets-list");
  }

  private templateCards(): CaspTemplateCard[] {
    return Array.from(this.querySelectorAll("casp-template-card"));
  }

  private mediaCards(): CaspMediaCard[] {
    return Array.from(this.querySelectorAll("casp-media-card"));
  }

  toConfig(position: {
    x: number;
    y: number;
    w: number;
    h: number;
  }): ui.GroupConfig {
    const mediaWidgets = this.mediaCards().map((c) =>
      c.toConfig({ x: 0, y: 0, w: 0, h: 0 }),
    );
    return ui.GroupConfig.createFrom({
      id: this.groupId,
      x: position.x,
      y: position.y,
      w: position.w,
      h: position.h,
      name: this.groupName,
      widgets: this.templateCards().map((c) =>
        c.toConfig({ x: 0, y: 0, w: 0, h: 0 }),
      ),
      mediaWidgets: mediaWidgets.length ? mediaWidgets : undefined,
    });
  }

  static fromConfig(config: ui.GroupConfig): CaspGroupCard {
    const group = document.createElement("casp-group-card") as CaspGroupCard;
    group.groupId = config.id || `group-${Date.now()}`;
    group.groupName = config.name || "New Group";
    group.queuedTemplates = config.widgets || [];
    group.queuedMediaWidgets = config.mediaWidgets || [];
    return group;
  }

  private addTemplate = () => {
    this.widgetsList()?.appendChild(
      document.createElement("casp-template-card"),
    );
    this.dispatchEvent(new CustomEvent("casp-change", { bubbles: true }));
  };

  private addMedia = () => {
    this.widgetsList()?.appendChild(document.createElement("casp-media-card"));
    this.dispatchEvent(new CustomEvent("casp-change", { bubbles: true }));
  };

  private onRemove = () => {
    this.dispatchEvent(
      new CustomEvent("casp-remove", {
        bubbles: true,
        detail: { element: this },
      }),
    );
  };

  // Handles a child template/media card's own "Remove" button: it only detaches itself
  // from the DOM (it's not a GridStack item, just a plain nested element), unlike a
  // top-level card, which app-root removes via GridStack's removeWidget instead.
  override connectedCallback() {
    super.connectedCallback();
    this.addEventListener("casp-remove", this.onChildRemove as EventListener);
  }

  override disconnectedCallback() {
    this.removeEventListener(
      "casp-remove",
      this.onChildRemove as EventListener,
    );
    super.disconnectedCallback();
  }

  private onChildRemove = (e: CustomEvent<{ element: HTMLElement }>) => {
    if (e.target === this) return; // this group's own remove button; let it bubble to app-root
    e.stopPropagation();
    e.detail.element.remove();
    this.dispatchEvent(new CustomEvent("casp-change", { bubbles: true }));
  };

  async executeAll() {
    this.error = "";
    try {
      const templateGroups = this.templateCards()
        .map((c) => c.toGroupItemOrNull())
        .filter((g): g is ui.CGDataGroup => g !== null);
      if (templateGroups.length) await api.pushCGDataGroup(templateGroups);

      const mediaGroups = this.mediaCards().map((c) => c.toMediaGroupItem());
      if (mediaGroups.length) await api.playMediaGroup(mediaGroups);
    } catch (e) {
      this.error = String(e);
    }
  }

  async stopAll() {
    this.error = "";
    try {
      const templateGroups = this.templateCards()
        .map((c) => c.toGroupItemOrNull())
        .filter((g): g is ui.CGDataGroup => g !== null);
      if (templateGroups.length) await api.stopCGDataGroup(templateGroups);

      const mediaGroups = this.mediaCards().map((c) => c.toMediaGroupItem());
      if (mediaGroups.length) await api.stopMediaGroup(mediaGroups);
    } catch (e) {
      this.error = String(e);
    }
  }

  async nextAll() {
    this.error = "";
    try {
      const templateGroups = this.templateCards()
        .map((c) => c.toGroupItemOrNull())
        .filter((g): g is ui.CGDataGroup => g !== null);
      if (templateGroups.length) await api.nextCGDataGroup(templateGroups);
    } catch (e) {
      this.error = String(e);
    }
  }

  protected override render() {
    return html`
      <div class="group-header">
        <input
          type="text"
          class="group-name-input edit-only"
          .value=${this.groupName}
          placeholder="Group name"
          @input=${(e: Event) => {
            this.groupName = (e.target as HTMLInputElement).value;
          }}
          @change=${() =>
            this.dispatchEvent(
              new CustomEvent("casp-change", { bubbles: true }),
            )}
        />
        <span class="group-name-display live-only">${this.groupName}</span>
        <div class="group-header-actions">
          <button class="action-btn live-only" @click=${this.executeAll}>
            ▶ Execute All
          </button>
          <button class="action-btn live-only" @click=${this.nextAll}>
            ⏭ Next All
          </button>
          <button class="action-btn live-only" @click=${this.stopAll}>
            ■ Stop All
          </button>
          <button class="add-element-btn edit-only" @click=${this.addTemplate}>
            + Add Template
          </button>
          <button class="add-media-btn edit-only" @click=${this.addMedia}>
            + Add Media Element
          </button>
          <button class="delete-btn edit-only" @click=${this.onRemove}>
            Remove Group
          </button>
        </div>
      </div>
      ${this.error ? html`<div class="widget-error">${this.error}</div>` : ""}
      <div class="group-widgets-list"></div>
    `;
  }
}

customElements.define("casp-group-card", CaspGroupCard);
