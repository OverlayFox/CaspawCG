/**
 * ConnectionStateManager — tracks CasparCG connection state and caches data.
 * Automatically refreshes templates/media when server reconnects.
 */
export const ConnectionStateManager = {
  _isConnected: false,
  _cachedTemplates: [],
  _cachedMedia: [],
  _subscribers: [],

  /**
   * Subscribe to data refresh events.
   * Callback receives { templates: [], media: [] } when data is updated.
   */
  subscribe(callback) {
    this._subscribers.push(callback);
  },

  unsubscribe(callback) {
    this._subscribers = this._subscribers.filter((cb) => cb !== callback);
  },

  _notifySubscribers() {
    const data = {
      templates: this._cachedTemplates,
      media: this._cachedMedia,
    };
    this._subscribers.forEach((callback) => {
      try {
        callback(data);
      } catch (error) {
        console.error("Error in ConnectionStateManager subscriber:", error);
      }
    });
  },

  /**
   * Called when CasparCG connection state changes.
   * If transitioning from offline to online, refreshes cached data.
   */
  async handleConnectionChange(isConnected) {
    const wasConnected = this._isConnected;
    this._isConnected = isConnected;

    // Server came back online - refresh data
    if (!wasConnected && isConnected) {
      console.log("CasparCG reconnected - refreshing templates and media");
      await this.refreshAllData();
    }
  },

  /**
   * Fetches latest templates and media, updates cache, and notifies subscribers.
   */
  async refreshAllData() {
    try {
      const [templates, media] = await Promise.all([
        window.go.ui.UIService.GetCasparCGTemplates(),
        window.go.ui.UIService.GetCasparCGMedia(),
      ]);

      // Only update and notify if we got actual data
      if (templates && templates.length > 0) {
        this._cachedTemplates = templates;
      }
      if (media && media.length > 0) {
        this._cachedMedia = media;
      }

      // Always notify subscribers even if one is empty
      // (the other might have data)
      if ((templates && templates.length > 0) || (media && media.length > 0)) {
        this._notifySubscribers();
      }
    } catch (error) {
      console.error("Failed to refresh CasparCG data:", error);
    }
  },

  getCachedTemplates() {
    return this._cachedTemplates;
  },

  getCachedMedia() {
    return this._cachedMedia;
  },
};

/**
 * APIService — all communication with the Wails Go backend.
 *
 * Every method logs its own errors and returns a safe default so callers
 * don't need to wrap individual calls in try/catch.
 */
export const APIService = {
  async clearAll() {
    try {
      await window.go.ui.UIService.ClearAll();
    } catch (error) {
      console.error("Failed to clear all data:", error);
    }
  },

  async clearChannels(channels) {
    try {
      await window.go.ui.UIService.ClearChannels(channels);
    } catch (error) {
      console.error("Failed to clear channels:", error);
    }
  },

  async getTemplateOptions() {
    try {
      const templates = await window.go.ui.UIService.GetCasparCGTemplates();
      // Update cache if we got data
      if (templates && templates.length > 0) {
        ConnectionStateManager._cachedTemplates = templates;
      }
      // Return cached data if available, otherwise return what we got (might be empty)
      return ConnectionStateManager._cachedTemplates.length > 0
        ? ConnectionStateManager._cachedTemplates
        : templates;
    } catch (error) {
      console.error("Failed to fetch template options:", error);
      // Return cached data if available, otherwise empty array
      return ConnectionStateManager._cachedTemplates.length > 0
        ? ConnectionStateManager._cachedTemplates
        : [];
    }
  },

  async getDataSources() {
    try {
      return await window.go.ui.UIService.GetDataSources();
    } catch (error) {
      console.error("Failed to fetch data sources:", error);
      return [];
    }
  },

  async pushCGData(
    template,
    layer = 1,
    channels = [1],
    data,
    posX = null,
    posY = null,
    sizeX = null,
    sizeY = null,
    delay = 0, // delay in nanoseconds as time.Duration is represented in Go as nanoseconds
  ) {
    try {
      const sizing = {
        posX: posX !== null ? parseInt(posX, 10) : 0,
        posY: posY !== null ? parseInt(posY, 10) : 0,
        sizeX: sizeX !== null ? parseFloat(sizeX) : 100,
        sizeY: sizeY !== null ? parseFloat(sizeY) : 100,
      };

      await window.go.ui.UIService.PushCasparCGData(
        template,
        layer,
        channels,
        data,
        sizing,
        delay,
      );
    } catch (error) {
      console.error("Failed to push CG data:", error);
    }
  },

  async pushCGDataGroup(dataGroups) {
    try {
      await window.go.ui.UIService.PushCasparCGDataGroup(dataGroups);
    } catch (error) {
      console.error("Failed to push CG data group:", error);
    }
  },

  async stopCGDataGroup(dataGroups) {
    try {
      await window.go.ui.UIService.StopCasparCGDataGroup(dataGroups);
    } catch (error) {
      console.error("Failed to stop CG data group:", error);
    }
  },

  async stopCGData(template, layer = 1, channels = [1], delay = 0) {
    try {
      await window.go.ui.UIService.StopCasparCGData(
        template,
        layer,
        channels,
        delay,
      );
    } catch (error) {
      console.error("Failed to stop CG data:", error);
    }
  },

  async getMediaOptions() {
    try {
      const media = await window.go.ui.UIService.GetCasparCGMedia();
      // Update cache if we got data
      if (media && media.length > 0) {
        ConnectionStateManager._cachedMedia = media;
      }
      // Return cached data if available, otherwise return what we got (might be empty)
      return ConnectionStateManager._cachedMedia.length > 0
        ? ConnectionStateManager._cachedMedia
        : media;
    } catch (error) {
      console.error("Failed to fetch media options:", error);
      // Return cached data if available, otherwise empty array
      return ConnectionStateManager._cachedMedia.length > 0
        ? ConnectionStateManager._cachedMedia
        : [];
    }
  },

  async getMediaInfo(filename) {
    try {
      return await window.go.ui.UIService.GetCasparCGMediaInfo(filename);
    } catch (error) {
      console.error("Failed to fetch media info:", error);
      return null;
    }
  },

  async playMedia(
    filename,
    layer = 1,
    channels = [1],
    loop = false,
    delay = 0,
  ) {
    try {
      await window.go.ui.UIService.PlayCasparCGMedia(
        filename,
        layer,
        channels,
        loop,
        delay,
      );
    } catch (error) {
      console.error("Failed to play media:", error);
    }
  },

  async stopMedia(layer = 1, channels = [1], delay = 0) {
    try {
      await window.go.ui.UIService.StopCasparCGMedia(layer, channels, delay);
    } catch (error) {
      console.error("Failed to stop media:", error);
    }
  },

  async fetchLiveData(identifier, type, source) {
    const result = await window.go.ui.UIService.GetDataSourceValue(source, {
      Key: identifier,
      Type: type,
    });
    return result.Value;
  },
};
