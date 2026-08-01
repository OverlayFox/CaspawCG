import { GridStack, type GridItemHTMLElement } from "gridstack";
import "gridstack/dist/gridstack.min.css";

export interface GridPosition {
  x: number;
  y: number;
  w: number;
  h: number;
}

export interface GridItemOptions {
  w: number;
  h: number;
  minW: number;
  minH: number;
  x?: number;
  y?: number;
}

/**
 * The only module that imports `gridstack` directly. Wraps a plain content element
 * (e.g. one Lit custom element) in the `.grid-stack-item > .grid-stack-item-content`
 * structure GridStack expects, so GridStack only ever manages outer position/size and
 * never touches the content element's own rendering.
 */
export class GridWrapper {
  private grid: GridStack | null = null;

  init(onChange: () => void): void {
    this.grid = GridStack.init({
      cellHeight: 100,
      margin: 10,
      float: true,
      column: 12,
      minRow: 1,
    });
    this.grid.on("change", onChange);
  }

  addItem(content: HTMLElement, options: GridItemOptions): GridItemHTMLElement {
    const wrapper = document.createElement("div");
    wrapper.className = "grid-stack-item";
    const contentWrapper = document.createElement("div");
    contentWrapper.className = "grid-stack-item-content";
    contentWrapper.appendChild(content);
    wrapper.appendChild(contentWrapper);

    return this.grid!.addWidget(wrapper, options);
  }

  removeItem(item: GridItemHTMLElement): void {
    this.grid!.removeWidget(item);
  }

  getItems(): GridItemHTMLElement[] {
    return this.grid!.getGridItems();
  }

  getPosition(item: GridItemHTMLElement): GridPosition {
    const node = item.gridstackNode!;
    return { x: node.x ?? 0, y: node.y ?? 0, w: node.w ?? 1, h: node.h ?? 1 };
  }

  enableEditing(enabled: boolean): void {
    this.grid!.enableMove(enabled);
    this.grid!.enableResize(enabled);
  }
}
