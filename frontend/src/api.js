/**
 * APIService — all communication with the Wails Go backend.
 *
 * Every method logs its own errors and returns a safe default so callers
 * don't need to wrap individual calls in try/catch.
 */
export const APIService = {
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

  async stopCGData(template, layer = 1, channel = 1, delay = 0) {
    try {
      await window.go.ui.UIService.StopCasparCGData(template, layer, channel, delay);
    } catch (error) {
      console.error("Failed to stop CG data:", error);
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
