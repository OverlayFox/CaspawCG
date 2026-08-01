import { LitElement, html } from "lit";
import { state } from "lit/decorators.js";

/** CaspConfirmModal — a reusable yes/no dialog. Call `.confirm(message)` and await the result. */
export class CaspConfirmModal extends LitElement {
  @state() private message = "";
  @state() private visible = false;
  private resolve: ((result: boolean) => void) | null = null;

  protected override createRenderRoot() {
    return this;
  }

  confirm(message: string): Promise<boolean> {
    this.message = message;
    this.visible = true;
    return new Promise((resolve) => {
      this.resolve = resolve;
    });
  }

  private close(result: boolean) {
    this.visible = false;
    this.resolve?.(result);
    this.resolve = null;
  }

  protected override render() {
    return html`
      <div class="modal-overlay" ?hidden=${!this.visible}>
        <div class="modal-dialog">
          <p class="modal-message">${this.message}</p>
          <div class="modal-actions">
            <button id="confirm-modal-cancel" @click=${() => this.close(false)}>Cancel</button>
            <button class="delete-btn" @click=${() => this.close(true)}>Confirm</button>
          </div>
        </div>
      </div>
    `;
  }
}

customElements.define("casp-confirm-modal", CaspConfirmModal);
