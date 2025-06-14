const EFFECT_ALLOWED = 'move';
const DROP_EFFECT = 'move';

export type DragContext = {
  dragTab: HTMLElement;
  initialIndex: number;
  newIndex?: number;
  dropped?: boolean;
};

export type DragTabsOptions = {
  selectors: {
    tabsContainer: string;
    tab: string;
    ignore: string;
    active?: string;
  };
  showPreview?: boolean;
};

export type DragTabsEvents = {
  start: DragContext;
  drag: { dragTab: HTMLElement; newIndex?: number; };
  end: { dragTab: HTMLElement; newIndex?: number; };
  cancel: { dragTab: HTMLElement; newIndex: number; };
};

class Emitter<Events extends Record<string, any>> {
  private _listeners: Map<keyof Events, Set<Function>> = new Map();
  on<K extends keyof Events>(type: K, listener: (event: Events[K]) => void) {
    if (!this._listeners.has(type)) this._listeners.set(type, new Set());
    (this._listeners.get(type) ?? new Set()).add(listener);
  }
  off<K extends keyof Events>(type: K, listener: (event: Events[K]) => void) {
    this._listeners.get(type)?.delete(listener);
  }
  emit<K extends keyof Events>(type: K, event: Events[K]) {
    for (const fn of this._listeners.get(type) ?? []) (fn as (e: Events[K]) => void)(event);
  }
}

export class DragTabs extends Emitter<DragTabsEvents> {
  $el: HTMLElement;
  options: DragTabsOptions;
  private _context: DragContext | null;
  private _moveTab: (event: Event) => boolean;
  private _onDragstart: (event: DragEvent) => void;
  private _onDragend: (event: Event) => void;
  private _onDrop: (event: Event) => void;

  constructor($el: HTMLElement, options: DragTabsOptions) {
    super();
    this.$el = $el;
    this.options = options;
    this._context = null;
    
    this._moveTab = this._moveTabImpl.bind(this);
    this._onDragstart = this._onDragstartImpl.bind(this);
    this._onDragend = this._onDragendImpl.bind(this);
    this._onDrop = this._onDropImpl.bind(this);
    this._init();
    this.update();
  }

  getActiveTabNode(): HTMLElement | null {
    const { active } = this.options.selectors;
    return active ? this.$el.querySelector<HTMLElement>(active) : null;
  }

  getTabsContainerNode(): HTMLElement | null {
    const { tabsContainer } = this.options.selectors;
    return this.$el.querySelector<HTMLElement>(tabsContainer);
  }

  getAllTabNodes(): HTMLElement[] {
    const { tab } = this.options.selectors;
    return Array.from(this.$el.querySelectorAll<HTMLElement>(tab));
  }

  private _setDraggable() {
    const allTabs = this.getAllTabNodes();
    const { ignore } = this.options.selectors;

    allTabs.forEach(tabNode => {
      if (ignore && tabNode.matches(ignore)) {
        tabNode.setAttribute('draggable', 'false');
      } else {
        tabNode.setAttribute('draggable', 'true');
      }
    });
  }

  update() {
    this._setDraggable();
  }

  private _init() {
    this.$el.addEventListener('dragstart', this._onDragstart);
  }

  private _bind(eventName: keyof HTMLElementEventMap, fn: EventListenerOrEventListenerObject, parent?: HTMLElement) {
    (parent || this.$el).addEventListener(eventName, fn);
  }

  private _unbind(eventName: keyof HTMLElementEventMap, fn: EventListenerOrEventListenerObject, parent?: HTMLElement) {
    (parent || this.$el).removeEventListener(eventName, fn);
  }

  private _moveTabImpl(event: Event): boolean {
    const dragEvent = event as DragEvent;
    const context = this._context;

    if (!context) {
      return cancelEvent(dragEvent);
    }

    const initialIndex = context.initialIndex;
    const tabContainer = this.getTabsContainerNode();

    if (!tabContainer) {
      return cancelEvent(dragEvent);
    }

    const currentIndex = typeof context.newIndex === 'number' ? context.newIndex : initialIndex;
    const target = dragEvent.target as HTMLElement;
    const dragTab = context.dragTab;

    if (!tabContainer.contains(target) || !target?.draggable || target === dragTab) {
      return cancelEvent(dragEvent);
    }

    const targetWidth = target.offsetWidth;
    const delta = targetWidth - dragTab.offsetWidth;

    let offset;
    if (delta > 0) {
      offset = delta / 2;
      if (offset > dragEvent.offsetX || (targetWidth - offset) < dragEvent.offsetX) {
        return cancelEvent(dragEvent);
      }
    }

    const children = Array.from(tabContainer.children) as HTMLElement[];
    const newIndex = children.indexOf(target);

    if (newIndex !== currentIndex) {
      context.newIndex = newIndex;
      this.emit('drag', { dragTab, newIndex });
    }

    return cancelEvent(dragEvent);
  }

  private _onDragstartImpl(event: DragEvent) {
    const tabContainer = this.getTabsContainerNode();
    const target = event.target as HTMLElement;
    if (!target.draggable || !tabContainer) {
      return cancelEvent(event);
    }

    const children = Array.from(tabContainer.children) as HTMLElement[];
    const initialIndex = children.indexOf(target);

    this._context = {
      dragTab: target,
      initialIndex
    };

    this.emit('start', this._context);
    this._bind('dragenter', cancelEvent as EventListener);
    this._bind('dragleave', cancelEvent as EventListener);
    this._bind('dragover', this._moveTab as EventListener);
    this._bind('dragend', this._onDragend as EventListener);
    this._bind('drop', this._onDrop as EventListener);
    this.$el.classList.add('dragging-active');

    const dataTransfer = event.dataTransfer;
    if (dataTransfer) {
      dataTransfer.dropEffect = DROP_EFFECT;
      dataTransfer.effectAllowed = EFFECT_ALLOWED;

      if (this.options.showPreview === false && 'setDragImage' in dataTransfer) {
        dataTransfer.setDragImage(emptyImage(), 0, 0);
      }

      dataTransfer.setData('text/plain', '');
    }
  }

  private _onDragendImpl(event: Event) {
    const $el = this.$el;
    const context = this._context;

    $el.classList.remove('dragging-active');

    this._unbind('dragenter', cancelEvent as EventListener);
    this._unbind('dragleave', cancelEvent as EventListener);
    this._unbind('dragover', this._moveTab as EventListener);
    this._unbind('dragend', this._onDragend as EventListener);
    this._unbind('drop', this._onDrop as EventListener);

    if (context) {
      this.emit('end', { dragTab: context.dragTab, newIndex: context.newIndex });

      if (!context.dropped) {
        this.emit('cancel', { dragTab: context.dragTab, newIndex: context.initialIndex });
      }

    }

    this._context = null;
  }

  private _onDropImpl(event: Event) {
    if (this._context) this._context.dropped = true;
  }
}

export default function create($el: HTMLElement & { __dragTabs?: DragTabs; }, options: DragTabsOptions | boolean): DragTabs | undefined {
  let dragTabs = get($el);

  if (!dragTabs && options !== false) {
    dragTabs = new DragTabs($el, options as DragTabsOptions);
    $el.__dragTabs = dragTabs;
  }

  return dragTabs;
}

function get($el: HTMLElement & { __dragTabs?: DragTabs; }): DragTabs | undefined {
  return $el.__dragTabs;
}

function cancelEvent(event: Event): boolean {
  event.preventDefault();
  event.stopPropagation();
  return false;
}

function emptyImage(): HTMLImageElement {
  const img = document.createElement('img');
  img.src = 'data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw==';
  return img;
}