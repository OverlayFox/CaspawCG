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

  async getTemplateOptions() {
    try {
      return await window.go.ui.UIService.GetCasparCGTemplates();
    } catch (error) {
      console.error("Failed to fetch template options:", error);
      return [];
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
    channel = 1,
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
        channel,
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

  async stopCGData(template, layer = 1, channel = 1, delay = 0) {
    try {
      await window.go.ui.UIService.StopCasparCGData(
        template,
        layer,
        channel,
        delay,
      );
    } catch (error) {
      console.error("Failed to stop CG data:", error);
    }
  },

  async getMediaOptions() {
    try {
      return await window.go.ui.UIService.GetCasparCGMedia();
    } catch (error) {
      console.error("Failed to fetch media options:", error);
      return [];
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

  async playMedia(filename, layer = 1, channel = 1, loop = false, delay = 0) {
    try {
      await window.go.ui.UIService.PlayCasparCGMedia(filename, layer, channel, loop, delay);
    } catch (error) {
      console.error("Failed to play media:", error);
    }
  },

  async stopMedia(layer = 1, channel = 1, delay = 0) {
    try {
      await window.go.ui.UIService.StopCasparCGMedia(layer, channel, delay);
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
