import { LitElement, html } from "lit";
import { state } from "lit/decorators.js";
import { connectionState } from "../lib/connection-state";
import { CASPAR_STATUS_EVENT, type CasparStatusEvent } from "../lib/events";

interface ClientStatus {
  host: string;
  port: number;
  isAlive: boolean;
}

/** CasparcgStatusBar — shows a connection chip per CasparCG client, from keep-alive events. */
export class CasparcgStatusBar extends LitElement {
  @state() private clients = new Map<string, ClientStatus>();

  protected override createRenderRoot() {
    return this;
  }

  override connectedCallback() {
    super.connectedCallback();
    window.addEventListener(
      CASPAR_STATUS_EVENT,
      this.handleStatus as EventListener,
    );
  }

  override disconnectedCallback() {
    window.removeEventListener(
      CASPAR_STATUS_EVENT,
      this.handleStatus as EventListener,
    );
    super.disconnectedCallback();
  }

  private handleStatus = (e: CasparStatusEvent) => {
    const { host, port, isAlive } = e.detail;
    const clientId = `caspar-${host}-${port}`.replace(/[^a-zA-Z0-9-]/g, "-");
    const previous = this.clients.get(clientId);

    const clients = new Map(this.clients);
    clients.set(clientId, { host, port, isAlive });
    this.clients = clients;

    if (previous?.isAlive !== isAlive) {
      connectionState.handleConnectionChange(isAlive);
    }
  };

  protected override render() {
    return html`
      <span class="status-title">CasparCG Server Status:</span>
      <div id="caspar-clients-container" class="status-clients">
        ${Array.from(this.clients.entries()).map(
          ([id, client]) => html`
            <div id=${id} class="client-chip">
              <div
                class="status-dot ${client.isAlive
                  ? "status-online"
                  : "status-offline"}"
              ></div>
              <span>${client.host}:${client.port}</span>
            </div>
          `,
        )}
      </div>
    `;
  }
}

customElements.define("casparcg-status-bar", CasparcgStatusBar);
