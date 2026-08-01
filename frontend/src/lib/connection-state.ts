import * as api from "./api";

export interface ConnectionData {
  templates: string[];
  media: string[];
}

type Subscriber = (data: ConnectionData) => void;

/**
 * Tracks CasparCG connection state and caches the last known templates/media list.
 * When the server reconnects, refreshes both lists and notifies subscribers so open
 * dropdowns can update themselves.
 */
class ConnectionState {
  private isConnected = false;
  private cachedTemplates: string[] = [];
  private cachedMedia: string[] = [];
  private subscribers: Subscriber[] = [];

  subscribe(callback: Subscriber): void {
    this.subscribers.push(callback);
  }

  unsubscribe(callback: Subscriber): void {
    this.subscribers = this.subscribers.filter((cb) => cb !== callback);
  }

  private notify(): void {
    const data: ConnectionData = { templates: this.cachedTemplates, media: this.cachedMedia };
    for (const callback of this.subscribers) {
      try {
        callback(data);
      } catch (error) {
        console.error("Error in connection-state subscriber:", error);
      }
    }
  }

  async handleConnectionChange(isConnected: boolean): Promise<void> {
    const wasConnected = this.isConnected;
    this.isConnected = isConnected;

    if (!wasConnected && isConnected) {
      console.log("CasparCG reconnected - refreshing templates and media");
      await this.refreshAllData();
    }
  }

  async refreshAllData(): Promise<void> {
    try {
      const [templates, media] = await Promise.all([api.getCasparCGTemplates(), api.getCasparCGMedia()]);
      if (templates?.length) this.cachedTemplates = templates;
      if (media?.length) this.cachedMedia = media;
      if (templates?.length || media?.length) this.notify();
    } catch (error) {
      console.error("Failed to refresh CasparCG data:", error);
    }
  }

  getCachedTemplates(): string[] {
    return this.cachedTemplates;
  }

  getCachedMedia(): string[] {
    return this.cachedMedia;
  }

  /** Fetches templates, falling back to the cache if the call fails or returns nothing. */
  async getTemplateOptions(): Promise<string[]> {
    try {
      const templates = await api.getCasparCGTemplates();
      if (templates?.length) this.cachedTemplates = templates;
      return this.cachedTemplates.length ? this.cachedTemplates : templates;
    } catch (error) {
      console.error("Failed to fetch template options:", error);
      return this.cachedTemplates;
    }
  }

  /** Fetches media filenames, falling back to the cache if the call fails or returns nothing. */
  async getMediaOptions(): Promise<string[]> {
    try {
      const media = await api.getCasparCGMedia();
      if (media?.length) this.cachedMedia = media;
      return this.cachedMedia.length ? this.cachedMedia : media;
    } catch (error) {
      console.error("Failed to fetch media options:", error);
      return this.cachedMedia;
    }
  }
}

export const connectionState = new ConnectionState();
