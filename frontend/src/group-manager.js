import { APIService } from "./api.js";
import {
  CSS_CLASSES,
  FIELD_TYPES,
  GROUP_CONTAINER_CLASS,
  INPUT_TYPES,
  SELECTORS,
} from "./constants.js";
import { DOMUtils } from "./dom-utils.js";
import { LayoutManager } from "./layout.js";
import { MediaWidgetManager } from "./media-widget-manager.js";
import { AppState } from "./state.js";
import { WidgetManager } from "./widget-manager.js";

/**
 * GroupManager — creates group container boxes in the main grid.
 *
 * A group is a special grid-stack-item that holds a list of widget cards.
 * In live mode, "Execute All" fires every widget in the group at once via
 * PushCasparCGDataGroup.
 */
export const GroupManager = {
  async create() {
    return this.createFromConfig(null);
  },

  async createFromConfig(config = null) {
    const groupId = config?.id || `group-${Date.now()}`;
    const groupName = config?.name || "New Group";

    const groupHTML = `
      <div class="grid-stack-item ${GROUP_CONTAINER_CLASS}" data-group-id="${groupId}">
        <div class="grid-stack-item-content group-card">
          <div class="group-header">
            <input type="text" class="group-name-input ${CSS_CLASSES.EDIT_ONLY}" value="${groupName}" placeholder="Group name">
            <span class="group-name-display ${CSS_CLASSES.LIVE_ONLY}">${groupName}</span>
            <div class="group-header-actions">
              <button class="${CSS_CLASSES.ACTION_BTN} ${CSS_CLASSES.LIVE_ONLY}" data-action="execute-group">▶ Execute All</button>
              <button class="${CSS_CLASSES.ACTION_BTN} ${CSS_CLASSES.LIVE_ONLY}" data-action="next-group">⏭ Next All</button>
              <button class="${CSS_CLASSES.ACTION_BTN} ${CSS_CLASSES.LIVE_ONLY}" data-action="stop-group">■ Stop All</button>
              <button class="add-element-btn ${CSS_CLASSES.EDIT_ONLY}">+ Add Dynamic Element</button>
              <button class="add-media-btn ${CSS_CLASSES.EDIT_ONLY}">+ Add Media Element</button>
              <button class="${CSS_CLASSES.DELETE_BTN} ${CSS_CLASSES.EDIT_ONLY}" data-action="remove-group">Remove Group</button>
            </div>
          </div>
          <div class="group-widgets-list"></div>
        </div>
      </div>
    `;

    const gridOptions = {
      w: config?.w || 4,
      h: config?.h || 3,
      minW: 3,
      minH: 3,
    };
    if (config) {
      gridOptions.x = config.x;
      gridOptions.y = config.y;
    }

    const gridItem = AppState.grid.addWidget(groupHTML, gridOptions);
    const groupCard = gridItem.querySelector(".group-card");
    const widgetsList = groupCard.querySelector(".group-widgets-list");

    if (config?.widgets?.length > 0) {
      for (const widgetConfig of config.widgets) {
        const entry = await WidgetManager.createForGroup(widgetConfig);
        widgetsList.appendChild(entry);
      }
    }

    if (config?.mediaWidgets?.length > 0) {
      for (const mediaConfig of config.mediaWidgets) {
        const entry = await MediaWidgetManager.createForGroup(mediaConfig);
        widgetsList.appendChild(entry);
      }
    }

    this._attachGroupListeners(gridItem, groupCard);
    return gridItem;
  },

  _attachGroupListeners(gridItem, groupCard) {
    const nameInput = groupCard.querySelector(".group-name-input");
    const nameDisplay = groupCard.querySelector(".group-name-display");
    const widgetsList = groupCard.querySelector(".group-widgets-list");

    nameInput?.addEventListener("input", () => {
      if (nameDisplay) nameDisplay.textContent = nameInput.value;
      LayoutManager.scheduleAutoSave();
    });

    groupCard
      .querySelector(".add-element-btn")
      ?.addEventListener("click", async () => {
        const entry = await WidgetManager.createForGroup(null);
        widgetsList.appendChild(entry);
        LayoutManager.scheduleAutoSave();
      });

    groupCard
      .querySelector(".add-media-btn")
      ?.addEventListener("click", async () => {
        const entry = await MediaWidgetManager.createForGroup(null);
        widgetsList.appendChild(entry);
        LayoutManager.scheduleAutoSave();
      });

    groupCard
      .querySelector(`[data-action="remove-group"]`)
      ?.addEventListener("click", () => {
        AppState.grid.removeWidget(gridItem);
        LayoutManager.scheduleAutoSave();
      });

    groupCard
      .querySelector(`[data-action="execute-group"]`)
      ?.addEventListener("click", () => {
        this.executeGroup(groupCard);
      });
    groupCard
      .querySelector(`[data-action="next-group"]`)
      ?.addEventListener("click", () => {
        this.nextGroup(groupCard);
      });
    groupCard
      .querySelector(`[data-action="stop-group"]`)
      ?.addEventListener("click", () => {
        this.stopGroup(groupCard);
      });
  },

  async executeGroup(groupCard) {
    const widgetEntries = [
      ...groupCard.querySelectorAll(".group-widgets-list .group-widget-entry"),
    ];

    const dynamicWidgetDataGroups = [];
    const mediaWidgetDataGroups = [];

    for (const entry of widgetEntries) {
      const widgetType = entry.getAttribute("data-widget-type") || "dynamic";

      if (widgetType === "media") {
        const mediaCard = entry.querySelector(
          `.${CSS_CLASSES.MEDIA_WIDGET_CARD}`,
        );
        if (mediaCard) {
          const mediaData = MediaWidgetManager.collectMediaData(mediaCard);
          if (mediaData) {
            mediaWidgetDataGroups.push(mediaData);
          }
        }
      } else {
        const widgetCard = entry.querySelector(`.${CSS_CLASSES.WIDGET_CARD}`);
        if (widgetCard) {
          const widgetData = await WidgetManager.collectWidgetData(widgetCard);
          if (widgetData) {
            dynamicWidgetDataGroups.push(widgetData);
          }
        }
      }
    }

    // Execute dynamic widgets if any
    if (dynamicWidgetDataGroups.length > 0) {
      console.log(
        `Executing group with ${dynamicWidgetDataGroups.length} dynamic elements`,
      );
      await APIService.pushCGDataGroup(dynamicWidgetDataGroups);
    }

    // Execute media widgets if any
    if (mediaWidgetDataGroups.length > 0) {
      console.log(
        `Executing group with ${mediaWidgetDataGroups.length} media elements`,
      );
      for (const mediaData of mediaWidgetDataGroups) {
        await APIService.playMedia(
          mediaData.filename,
          mediaData.layer,
          mediaData.channels,
          mediaData.loop,
          mediaData.delay,
        );
      }
    }
  },

  async stopGroup(groupCard) {
    const widgetEntries = [
      ...groupCard.querySelectorAll(".group-widgets-list .group-widget-entry"),
    ];

    const dynamicWidgetDataGroups = [];
    const mediaWidgetDataGroups = [];

    for (const entry of widgetEntries) {
      const widgetType = entry.getAttribute("data-widget-type") || "dynamic";

      if (widgetType === "media") {
        const mediaCard = entry.querySelector(
          `.${CSS_CLASSES.MEDIA_WIDGET_CARD}`,
        );
        if (mediaCard) {
          const mediaData = MediaWidgetManager.collectMediaData(mediaCard);
          if (mediaData) {
            mediaWidgetDataGroups.push(mediaData);
          }
        }
      } else {
        const widgetCard = entry.querySelector(`.${CSS_CLASSES.WIDGET_CARD}`);
        if (widgetCard) {
          const widgetData = await WidgetManager.collectWidgetData(widgetCard);
          if (widgetData) {
            dynamicWidgetDataGroups.push(widgetData);
          }
        }
      }
    }

    // Stop dynamic widgets if any
    if (dynamicWidgetDataGroups.length > 0) {
      console.log(
        `Stopping group with ${dynamicWidgetDataGroups.length} dynamic elements`,
      );
      await APIService.stopCGDataGroup(dynamicWidgetDataGroups);
    }

    // Stop media widgets if any
    if (mediaWidgetDataGroups.length > 0) {
      console.log(
        `Stopping group with ${mediaWidgetDataGroups.length} media elements`,
      );
      for (const mediaData of mediaWidgetDataGroups) {
        await APIService.stopMedia(
          mediaData.layer,
          mediaData.channels,
          mediaData.delay,
        );
      }
    }
  },

  async nextGroup(groupCard) {
    const widgetEntries = [
      ...groupCard.querySelectorAll(".group-widgets-list .group-widget-entry"),
    ];

    for (const entry of widgetEntries) {
      if ((entry.getAttribute("data-widget-type") || "dynamic") !== "dynamic")
        continue;

      const widgetCard = entry.querySelector(`.${CSS_CLASSES.WIDGET_CARD}`);
      if (!widgetCard) continue;

      const cgData = await WidgetManager.collectWidgetData(widgetCard);
      if (!cgData) continue;

      await APIService.nextCGData(
        cgData.template,
        cgData.layer,
        cgData.channels,
        cgData.delay,
      );
    }
  },

  serializeGroups() {
    const groups = [];

    AppState.grid.getGridItems().forEach((item) => {
      if (!item.classList.contains(GROUP_CONTAINER_CLASS)) return;

      const groupId =
        item.getAttribute("data-group-id") || `group-${Date.now()}`;
      const node = item.gridstackNode;
      const groupCard = item.querySelector(".group-card");
      if (!groupCard) return;

      const name =
        groupCard.querySelector(".group-name-input")?.value?.trim() || "Group";

      const widgets = [];
      const mediaWidgets = [];
      groupCard.querySelectorAll(".group-widget-entry").forEach((entry) => {
        const widgetId =
          entry.getAttribute("data-widget-id") || `w-${Date.now()}`;
        const widgetType = entry.getAttribute("data-widget-type") || "dynamic";

        if (widgetType === "media") {
          const card = entry.querySelector(`.${CSS_CLASSES.MEDIA_WIDGET_CARD}`);
          if (!card) return;

          const filename = DOMUtils.querySelector(
            ".media-dropdown",
            card,
          )?.value;
          const nameInput = DOMUtils.querySelector(".widget-name-input", card);
          const channelInput = DOMUtils.querySelector(".channel-input", card);
          const loop =
            DOMUtils.querySelector(".loop-input", card)?.checked ?? false;
          const delayVal = DOMUtils.querySelector(".delay-input", card)?.value;

          mediaWidgets.push({
            id: widgetId,
            x: 0,
            y: 0,
            w: 0,
            h: 0,
            name: nameInput?.value || "Media Element",
            filename: filename || "",
            layer:
              parseInt(
                DOMUtils.querySelector(".layer-input", card)?.value,
                10,
              ) || 1,
            channel: parseInt(channelInput?.value, 10) || 1,
            channelExpr: channelInput?.value || "1",
            loop: loop,
            delay: delayVal ? parseInt(delayVal, 10) : 0,
          });
        } else {
          const card = entry.querySelector(`.${CSS_CLASSES.WIDGET_CARD}`);
          if (!card) return;

          const fields = [];
          DOMUtils.querySelectorAll(`.${CSS_CLASSES.FIELD_ROW}`, card).forEach(
            (row) => {
              const keyInput = DOMUtils.querySelector(SELECTORS.FIELD_KEY, row);
              if (!keyInput?.value) return;
              const inputType =
                DOMUtils.querySelector(SELECTORS.FIELD_INPUT_TYPE, row)
                  ?.value || INPUT_TYPES.DATASOURCE;
              fields.push({
                key: keyInput.value,
                type:
                  DOMUtils.querySelector(SELECTORS.FIELD_TYPE, row)?.value ||
                  FIELD_TYPES.STRING,
                inputType,
                id:
                  inputType === INPUT_TYPES.DIRECT
                    ? ""
                    : DOMUtils.querySelector(SELECTORS.FIELD_ID, row)?.value ||
                      "",
                source:
                  inputType === INPUT_TYPES.DIRECT
                    ? ""
                    : DOMUtils.querySelector(SELECTORS.FIELD_SOURCE, row)
                        ?.value || "",
                value:
                  inputType === INPUT_TYPES.DIRECT
                    ? DOMUtils.querySelector(SELECTORS.FIELD_DIRECT_VALUE, row)
                        ?.value || ""
                    : "",
              });
            },
          );

          const posXVal = DOMUtils.querySelector(".pos-x-input", card)?.value;
          const posYVal = DOMUtils.querySelector(".pos-y-input", card)?.value;
          const sizeXVal = DOMUtils.querySelector(".size-x-input", card)?.value;
          const sizeYVal = DOMUtils.querySelector(".size-y-input", card)?.value;
          const delayVal = DOMUtils.querySelector(".delay-input", card)?.value;
          const nameInput = DOMUtils.querySelector(".widget-name-input", card);
          const channelInput = DOMUtils.querySelector(".channel-input", card);

          widgets.push({
            id: widgetId,
            x: 0,
            y: 0,
            w: 0,
            h: 0,
            name: nameInput?.value || "Dynamic Element",
            template:
              DOMUtils.querySelector(".api-dropdown", card)?.value || "",
            layer:
              parseInt(
                DOMUtils.querySelector(".layer-input", card)?.value,
                10,
              ) || 1,
            channel: parseInt(channelInput?.value, 10) || 1,
            channelExpr: channelInput?.value || "1",
            posX: posXVal ? parseInt(posXVal, 10) : null,
            posY: posYVal ? parseInt(posYVal, 10) : null,
            sizeX: sizeXVal ? parseFloat(sizeXVal) : null,
            sizeY: sizeYVal ? parseFloat(sizeYVal) : null,
            delay: delayVal ? parseInt(delayVal, 10) : 0,
            fields,
          });
        }
      });

      const groupData = {
        id: groupId,
        x: node.x,
        y: node.y,
        w: node.w,
        h: node.h,
        name,
        widgets,
      };

      // Only add mediaWidgets if there are any
      if (mediaWidgets.length > 0) {
        groupData.mediaWidgets = mediaWidgets;
      }

      groups.push(groupData);
    });

    return groups;
  },
};
