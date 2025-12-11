# Frontend Layouts Extensibility and Persistence Plan

This document details the plan for restructuring and refactoring the frontend layout system to support dynamic view registration (for future plugin integration) and enable user layouts to be saved and loaded from both local IndexedDB and a remote backend API. This plan fully embraces Svelte 5's functional component model and runes.

> NOTE: This is mostly AI-generated slop and will be completely rewritten with clearer documentation. While the fundamental idea of what this document describes remains true, this document is garbage and implementation should NOT be based off of this document.
---

### 1. Core Architectural Principles

The frontend layout system will adhere to several core architectural principles:

*   **Separation of Concerns:** Distinct services will handle view registration, layout persistence, and UI rendering.
*   **Dynamic Extensibility:** The system will allow for new view types, including those from plugins, to be registered at runtime.
*   **Persistence:** The design will support serialization and deserialization of the entire layout state for saving and loading.
*   **Svelte 5 Idiomatic:** The implementation will leverage Svelte 5 runes (`$state`, `$derived`, `$effect`) and the functional component paradigm.
*   **Imperative Rendering for Plugins:** The plan includes provisions for `mount()`/`unmount()` functions to support truly dynamic and remote components.

---

### 2. New Components/Services

#### 2.1. `viz/src/lib/layouts/viewRegistry.ts` (View Registry Service)

This will be a singleton service responsible for centrally managing available Svelte component definitions that can be used as views in the layout. This new service replaces the previous static `views` array.

**`RegisteredViewDefinition` Interface:**


```typescript
import type { Component } from 'svelte'; // Svelte 5 component function type

export interface RegisteredViewDefinition {
    id: number;              // Unique numerical ID for the view type
    name: string;            // Display name (e.g., "Collections", "Clock")
    component: Component;    // The actual Svelte 5 component function
    path?: string;           // Optional path, useful for preloading data (e.g., "/collections")
    pluginMeta?: {           // Optional: Metadata for plugin-provided views
        pluginIdentifier: string;
        pluginViewIdentifier: string;
        // ... any other plugin-specific config/data for the host component
    };
}
```

**`ViewRegistry` Class:**


```typescript
import type { RegisteredViewDefinition } from './viewRegistry';

class ViewRegistry {
    private views: Map<number, RegisteredViewDefinition> = new Map();
    private nameToIdMap: Map<string, number> = new Map();
    private pathToIdMap: Map<string, number> = new Map();
    private nextId: number = 1;

    public registerView(definition: Omit<RegisteredViewDefinition, 'id'>): RegisteredViewDefinition {
        const id = this.nextId++;
        const registeredDef: RegisteredViewDefinition = { ...definition, id };
        this.views.set(id, registeredDef);
        this.nameToIdMap.set(definition.name, id);
        if (definition.path) {
            this.pathToIdMap.set(definition.path, id);
        }
        return registeredDef;
    }

    public getViewById(id: number): RegisteredViewDefinition | undefined {
        return this.views.get(id);
    }

    public getViewByName(name: string): RegisteredViewDefinition | undefined {
        const id = this.nameToIdMap.get(name);
        return id ? this.views.get(id) : undefined;
    }

    public getViewByPath(path: string): RegisteredViewDefinition | undefined {
        const id = this.pathToIdMap.get(path);
        return id ? this.views.get(id) : undefined;
    }

    public getAllRegisteredViews(): RegisteredViewDefinition[] {
        return Array.from(this.views.values());
    }
}
```

A singleton instance `export const viewRegistry = new ViewRegistry();` will be exported.

#### 2.2. `viz/src/lib/layouts/layoutPersistence.ts` (Layout Persistence Service)

This service will manage the saving and loading of layout configurations, handling both local (IndexedDB) and remote (backend API) storage.

**`PersistentLayoutConfig` Interface:**


```typescript
import type { SerializedVizView } from '$lib/views/views.svelte';

export interface PersistentLayoutConfig {
    name: string;
    description?: string;
    id?: string; // Identifier for a saved layout (local or remote)
    tree: SerializedVizSubPanelData[]; // The serialized layout tree structure
}
```

**`LayoutPersistence` Class:**


```typescript
import { writable, type Writable } from 'svelte/store';
import type { Component } from 'svelte';
import type { SerializedVizView } from '$lib/views/views.svelte';
import { hydrateVizView } from '$lib/views/hydrateVizView';
import VizView from '$lib/views/views.svelte'; // Assuming VizView is a generic class/type, e.g., VizView<ViewData>

// For now, VizSubPanelData mirrors SerializedVizSubPanelData for simplicity in this plan.
type VizSubPanelData = SerializedVizSubPanelData;

export type SerializedVizSubPanelData = {
    type: 'pane' | 'group';
    id: string;
    size: number;
    views?: SerializedVizView[]; // For 'pane' type
    children?: SerializedVizSubPanelData[]; // For 'group' type
    direction?: 'horizontal' | 'vertical';
};


export interface PersistentLayoutConfig {
    name: string;
    description?: string;
    id?: string; // Identifier for a saved layout (local or remote)
    tree: SerializedVizSubPanelData[]; // The serialized layout tree structure
}

class LayoutPersistence<ViewData = any> {
    public activeLayout: Writable<PersistentLayoutConfig | null> = writable<PersistentLayoutConfig | null>(null);
    public async saveLayout(
        name: string,
        currentLayoutTree: VizSubPanelData[],
        type: 'local' | 'remote',
        description?: string
    ): Promise<string> {
        const serializedTree = this.serializeLayoutTree(currentLayoutTree);
        const layoutConfig: PersistentLayoutConfig = { name, description, tree: serializedTree };

        if (type === 'local') {
            return this.saveToIndexedDB(layoutConfig);
        } else {
            return this.saveToBackend(layoutConfig);
        }
    }

    public async loadLayout(layoutId: string, type: 'local' | 'remote'): Promise<PersistentLayoutConfig | null> {
        let layoutConfig: PersistentLayoutConfig | null = null;
        if (type === 'local') {
            layoutConfig = await this.loadFromIndexedDB(layoutId);
        } else {
            layoutConfig = await this.loadFromBackend(layoutId);
        }

        if (layoutConfig) {
            // Assuming layoutState.tree needs to be updated with live VizSubPanelData
            // This part requires access to the actual layoutState or a mechanism to update it.
            // For now, we'll just return the hydrated tree.
            // const liveLayoutTree = this.deserializeLayoutTree(layoutConfig.tree);
            this.activeLayout.set(layoutConfig); // Update the Svelte store
        }
        return layoutConfig;
    }

    public async getAvailableLayouts(
        type: 'local' | 'remote'
    ): Promise<Array<{ id: string; name: string; description?: string }>> {
        if (type === 'local') {
            return this.getLayoutsFromIndexedDB();
        } else {
            return this.getLayoutsFromBackend();
        }
    }

    private serializeLayoutTree(tree: VizSubPanelData[]): SerializedVizSubPanelData[] {
        // This is a conceptual implementation. Real implementation would recursively
        // traverse the VizSubPanelData and convert VizView instances to SerializedVizView.
        // Assuming VizSubPanelData has a structure that can be directly serialized or
        // contains VizView instances that have a .toJSON() method.
        return tree.map(item => {
            if (item.views) {
                return {
                    ...item,
                    views: item.views.map((view: VizView<ViewData>) => view.toJSON())
                } as SerializedVizSubPanelData;
            }
            if (item.children) {
                return {
                    ...item,
                    children: this.serializeLayoutTree(item.children)
                } as SerializedVizSubPanelData;
            }
            return item as SerializedVizSubPanelData;
        });
    }

    private deserializeLayoutTree(serializedTree: SerializedVizSubPanelData[]): VizSubPanelData[] {
        // This is a conceptual implementation. Real implementation would recursively
        // traverse the SerializedVizSubPanelData and reconstruct VizView instances.
        return serializedTree.map(item => {
            if (item.views) {
                return {
                    ...item,
                    views: item.views.map(serializedView => hydrateVizView<ViewData>(serializedView))
                } as VizSubPanelData;
            }
            if (item.children) {
                return {
                    ...item,
                    children: this.deserializeLayoutTree(item.children)
                } as VizSubPanelData;
            }
            return item as VizSubPanelData;
        });
    }

    private async saveToIndexedDB(layout: PersistentLayoutConfig): Promise<string> {
        console.log('Saving layout to IndexedDB (conceptual):', layout.name);
        // Simulate IndexedDB save and return a generated ID
        return Promise.resolve(`local-${Date.now()}`);
    }

    private async loadFromIndexedDB(id: string): Promise<PersistentLayoutConfig | null> {
        console.log('Loading layout from IndexedDB (conceptual):', id);
        // Simulate IndexedDB load
        return Promise.resolve(null);
    }

    private async getLayoutsFromIndexedDB(): Promise<Array<{ id: string; name: string; description?: string }>> {
        console.log('Getting layouts from IndexedDB (conceptual)');
        // Simulate IndexedDB list
        return Promise.resolve([]);
    }

    private async saveToBackend(layout: PersistentLayoutConfig): Promise<string> {
        console.log('Saving layout to backend (conceptual):', layout.name);
        // Simulate API call to save layout
        return Promise.resolve(`remote-${Date.now()}`);
    }

    private async loadFromBackend(id: string): Promise<PersistentLayoutConfig | null> {
        console.log('Loading layout from backend (conceptual):', id);
        // Simulate API call to load layout
        return Promise.resolve(null);
    }

    private async getLayoutsFromBackend(): Promise<Array<{ id: string; name: string; description?: string }>> {
        console.log('Getting layouts from backend (conceptual)');
        // Simulate API call to list layouts
        return Promise.resolve([]);
    }
}
```

A singleton instance `export const layoutPersistence = new LayoutPersistence();` will be exported.

#### 2.3. `viz/src/lib/views/hydrateVizView.ts` (View Hydration Utility)

This is a dedicated function designed to reconstruct a `VizView` instance from its serialized data, making use of the `viewRegistry`.

**Function Signature:**


```typescript
import type { SerializedVizView } from './views.svelte';
import VizView from './views.svelte';
import { viewRegistry } from '$lib/layouts/viewRegistry';
import type { Component } from 'svelte';

export function hydrateVizView<ViewData = any>(serialized: SerializedVizView): VizView<ViewData> {
    const registeredDef = viewRegistry.getViewById(serialized.id); // Lookup by registered view ID
    if (!registeredDef) {
        console.warn(`Component for view ID ${serialized.id} not found in registry. Returning fallback.`);
        // Return a fallback/error component if the definition is not found
        // This would be a basic Svelte component (e.g., ErrorView.svelte) registered beforehand
        const errorViewDef = viewRegistry.getViewById(/* ID of pre-registered ErrorView */);
        return new VizView({
            name: `Error: ${serialized.name}`,
            component: (errorViewDef ? errorViewDef.component : (() => { /* default empty component */ }) as Component),
            id: serialized.id,
            path: serialized.path,
            // ... other properties, possibly an error message prop
        });
    }
    // Use the original VizView.fromJSON, passing the resolved component
    return VizView.fromJSON(serialized, registeredDef.component);
}
```

#### 2.4. `viz/src/lib/components/panels/PluginHostView.svelte` (Host Component for Plugin UI)

**Purpose:** This generic Svelte 5 component will be registered in `viewRegistry.ts` and act as a container for dynamically loaded and rendered plugin content, including imperative mounting of Svelte components.

**Component Structure:**


```svelte
<script lang="ts">
    import { mount, unmount } from 'svelte'; // Assuming these are available or mocked for documentation

    // Props
    const {
        pluginIdentifier,
        pluginViewIdentifier,
        pluginType,
        initialPluginData
    } = $props<{
        pluginIdentifier: string;
        pluginViewIdentifier: string;
        pluginType: 'svelte-component' | 'html-content' | 'data-driven' | 'custom-element';
        initialPluginData?: any;
    }>();

    // Internal State (`$state`)
    let mountTarget: HTMLElement | undefined = $state();
    let mountedComponentInstance: ReturnType<typeof mount> | undefined = $state();

    // The rest of the rendering logic and effects would go here.
    // (e.g., $effect(() => { ... }))
</script>

<!-- Markup would be here -->
```
**Rendering Logic (`$effect` and lifecycle hooks):**

The rendering logic will vary based on the `pluginType`:

*   For `pluginType === 'html-content'`, the component will use `{@html pluginContent}` where `pluginContent` is fetched from the plugin logic.
*   For `pluginType === 'data-driven'`, it will render its own Svelte template based on `pluginData` fetched from the plugin.
*   For `pluginType === 'custom-element'`:
    *   Within `$effect` or `onMount`, `document.createElement(pluginCustomElementTagName)` will be created.
    *   It will then be appended to `mountTarget`.
    *   Properties will be set on the custom element (e.g., `el.prop = value`).
    *   In `onDestroy`, the custom element will be removed.
*   For `pluginType === 'svelte-component'` (Imperative Svelte 5 Component Loading):
    *   **Loading Remote Component Code:** This is the most complex part. The component will be responsible for orchestrating the loading of the actual Svelte component *function* from a remote source (e.g., via dynamic `import()` of a known URL, or by receiving a compiled module from Extism and instantiating it). This assumes the remote component is already compiled to a Svelte 5 functional component.
    *   **Mounting:** In an `$effect` reacting to `mountTarget` and the `remoteComponentFunction` becoming available, the component will mount: `mountedComponentInstance = mount(remoteComponentFunction, { target: mountTarget, props: { ...initialPluginData } });`. `flushSync()` can be considered if immediate DOM update is needed.
    *   **Unmounting:** In an `$effect` or `onDestroy` callback reacting to `mountedComponentInstance` needing to be removed, the component will unmount: `unmount(mountedComponentInstance);`. Any plugin-specific cleanup (e.g., calling `plugin.unmount(pluginViewIdentifier)`) will also be performed.

**Security:** This component must implement strong security measures, especially for `html-content` and remote `svelte-component` types (e.g., DOMPurify for HTML, strict CSP).

---

### 3. Updates to Existing Files

#### 3.1. `viz/src/lib/layouts/views.ts`

The static `export const views: VizView[] = [...]` array will be **removed**. All these views will be registered via `viewRegistry.registerView()` at application startup.

#### 3.2. `viz/src/lib/views/views.svelte.ts` (`VizView` Class)

The `component` property will explicitly type as `Component` (Svelte 5 functional component). The static `fromJSON` method will be simplified to `static fromJSON(serialized: SerializedVizView, component: Component): VizView<any>`. Its main responsibility will be to construct a new `VizView` instance with the provided component, rather than performing the component lookup itself. The lookup will be handled by `hydrateVizView`.

#### 3.3. `viz/src/lib/components/panels/SubPanel.svelte`



**Actions:**

The `import { views } from "$lib/layouts/views";` statement will be removed, and `viewRegistry` and `layoutPersistence` services will be imported instead.



**Initial View Hydration:** The current hydration loop for `initialViews` will be refactored. It will receive `SerializedVizView[]` data (from `layoutPersistence`). For each `SerializedVizView`, it will use `hydrateVizView(serializedView)` to reconstruct the live `VizView` instances.



**Rendering `activeView.component` (Svelte 5 Functional Component Invocation):**

```svelte

{#if activeView?.component}

    {activeView.component({

        // Pass props explicitly as arguments

        keyId, tabDropper, panelData, panelViews, activeView, subPanelContentFocused, onFocus: () => (subPanelContentFocused = true)

    })}

{/if}

```

**Panel Operations (`closeTab`, `splitRight`, etc.):**

These functions will manipulate the `panelViews` array, which holds live `VizView` instances. An `$effect` will be introduced to react to changes in `panelViews`. This `$effect` will:

1.  Call `layoutPersistence.serializeLayoutTree()` with the current `panelViews` (or the relevant part of the tree).

2.  Update the corresponding part of `layoutState.tree` with the serialized data.

3.  Potentially trigger `layoutPersistence.saveLayout()` (e.g., debounced autosave).



When creating new views (e.g., duplicating), `viewRegistry` will be used to get `RegisteredViewDefinition` and construct new `VizView` instances.

#### 3.4. `viz/src/lib/views/tabs.svelte.ts`

**Actions:**
The `import { views } from "$lib/layouts/views";` statement will be removed, and `viewRegistry` and `layoutPersistence` will be imported.

**View Creation/Duplication:** When creating new `VizView` instances (e.g., for split panels or dropped tabs), they must be constructed using a component obtained from `viewRegistry`. All manipulations of the layout state (`layoutState.tree`) must involve `SerializedVizView` objects or `VizView` instances that are immediately serialized before updating `layoutState.tree`. `handleTabDropInside` and other drag-and-drop related functions will ensure that new `Content` groups are added to `layoutState.tree` with `SerializedVizView` objects, utilizing `hydrateVizView` if starting from a serialized form, or `VizView.toJSON()` if starting from a live `VizView`.

#### 3.5. `viz/src/lib/third-party/svelte-splitpanes/state.svelte.ts`

**Action:** The `layoutState.tree` will now strictly contain `VizSubPanelData` where the `views` arrays within its `Content` objects hold `SerializedVizView` objects, not live `VizView` instances. This ensures `layoutState.tree` is always ready for persistence.
**Type Update:** The type of `layoutState.tree` will be updated to reflect `SerializedVizSubPanelData[]`.

---

### 4. Svelte 5 Specific Considerations

*   **Functional Components:** All Svelte components developed will adhere to the Svelte 5 functional component paradigm, utilizing direct function invocation (`{Component({...})}`) for dynamic rendering.
*   **Imperative `mount()`/`unmount()`:** The `PluginHostView.svelte` component will specifically utilize Svelte 5's `mount()` and `unmount()` functions to facilitate the dynamic loading and rendering of Svelte components sourced from remote locations.
*   **Runes:** The reactive primitives `$state`, `$derived`, and `$effect` will be consistently applied for managing reactive state, computed values, and side effects throughout all affected Svelte components.
*   **Event Handling:** The new event handling mechanisms introduced in Svelte 5, such as `onclick={handler}` and the use of callback props for component events, will be adopted.

---

### 5. Key Challenges/Considerations for Plugin System

*   **Security for Remote Components:** Implementing robust security measures is crucial, particularly for handling remote components. This includes enforcing strict Content Security Policy (CSP) rules to mitigate risks associated with loading and executing remote JavaScript, establishing a clear strategy for trusting plugin sources (e.g., signed modules, curated marketplaces), and leveraging Extism's WebAssembly (Wasm) sandboxing for plugin logic. Additional sandboxing may be necessary for UI-specific code running in the main JavaScript thread.
*   **Module Loading for Remote Svelte Components:** A primary challenge lies in determining how compiled Svelte 5 component *functions* will be delivered and made available as JavaScript modules at runtime. Potential options include dynamic `import()` of known URLs, custom module loaders, or specific build strategies such as Webpack/Vite module federation, if compatible with Svelte 5.
*   **Plugin API/Contract:** A well-defined Application Programming Interface (API) is essential for plugins to expose (e.g., via Wasm host functions or a manifest). This API will enable `PluginHostView.svelte` to query available views, provide initial data, and manage lifecycle events effectively.
*   **Type Safety:** Ensuring type safety across plugin boundaries is critical, especially when passing properties (props) to dynamically loaded components to maintain code integrity and reduce runtime errors.
---

### 6. Alternative Plugin Approach: Iframes with Message Passing (Figma-like)

Given the desire for robust isolation and technology agnosticism for external plugins, an approach leveraging iframes with message passing will be used. This is similar to how applications like Figma integrate third-party UI, but with a strong recommendation and SDK support for building plugin UIs specifically in Svelte.

#### 6.1. Core Principles of the Iframe Approach

The iframe approach is built upon several core principles to ensure robust and flexible plugin integration:

*   **Process Isolation:** Each plugin UI operates within its own browsing context (iframe), completely isolated from the main application's DOM, CSS, and JavaScript. This isolation is crucial for preventing style clashes, mitigating script errors, and blocking unauthorized DOM access.
*   **Technology Agnostic (with Svelte recommendation):** While iframes naturally support any web technology, plugins are *recommended* to be built using Svelte. A dedicated SDK will be provided to expose core classes and utilities, ensuring seamless interoperability and strong type safety for Svelte-based plugin UIs.
*   **Secure & Type-Safe Communication:** All interactions between the main application and a plugin are conducted through a well-defined `window.postMessage` API. The accompanying SDK will provide essential type definitions and helper functions to ensure that this communication is both secure (through strict origin validation) and type-safe.
*   **Decoupled Deployment:** Plugins can be hosted and deployed independently from the main application. This architectural separation offers greater flexibility and scalability for plugin development and maintenance.

#### 6.2. Architectural Changes for Iframe Integration

To support this approach, the following components and modifications would be necessary:

##### 6.2.1. `viz/src/lib/components/panels/IframePluginHost.svelte` (New Host Component)

This Svelte component will serve as the container for embedding external plugin UIs within an `<iframe>`. Its responsibilities include managing the iframe's lifecycle, handling communication, and basic rendering.

**Props (`$props`):**


```typescript
interface IframePluginHostProps {
    pluginUrl: string; // The URL of the external plugin application to load in the iframe
    pluginIdentifier: string; // Unique identifier for the plugin
    pluginViewIdentifier: string; // Identifier for the specific view within the plugin
    initialData?: any; // Initial data to send to the plugin upon loading
}
```

**Internal State (`$state`):**


```typescript
let iframeRef: HTMLIFrameElement | undefined = $state();
let messageCounter: number = $state(0); // For tracking messages, if needed
```

**Logic (`$effect` and Event Listeners):**

**Loading:** The `pluginUrl` prop will be used to set the `src` attribute of the `<iframe>`.

**Outgoing Messages (to plugin):** A method will be provided to send messages to the iframe. This method will internally leverage the SDK's messaging utilities.
```typescript
import { sendPluginMessage } from '$lib/sdk/messaging'; // Assuming an SDK utility

function sendMessageToPlugin(type: string, payload: any): void {
    if (iframeRef?.contentWindow && pluginUrl) {
        // sendPluginMessage handles targetOrigin and type-checking via SDK
        sendPluginMessage(iframeRef.contentWindow, pluginUrl, { type, payload });
    }
}
```

**Incoming Messages (from plugin):** The component will listen for `message` events from the iframe, leveraging the SDK for secure parsing and type-checking.
```typescript
import { onPluginMessage } from '$lib/sdk/messaging'; // Assuming an SDK utility

$effect(() => {
    const unsubscribe = onPluginMessage(window, (message) => {
        // message is already type-checked and origin-validated by SDK
        console.log('Message from plugin:', message);
        // Dispatch custom events or update stores in the main application
    }, pluginUrl); // SDK uses pluginUrl for origin validation

    return () => unsubscribe();
});
```

**Markup:**


```svelte
<iframe
    bind:this={iframeRef}
    src={pluginUrl}
    title={`Plugin: ${pluginIdentifier}-${pluginViewIdentifier}`}
    sandbox="allow-scripts allow-same-origin allow-popups allow-forms"
    allow="clipboard-write;"
    style="width: 100%; height: 100%; border: none;"
></iframe>
```

The `sandbox` attribute is crucial for security, limiting what the iframe can do. Permissions (e.g., `allow-forms`, `allow-popups`, `allow-downloads`) should be adjusted based on specific plugin requirements.

##### 6.2.2. Plugin Application (within the iframe, built with Svelte and SDK)

Each external plugin will be a self-contained Svelte web application served within its iframe. It will import and utilize the SDK for structured communication with the main application.

##### 6.2.3. Adopting JSON-RPC for the Messaging Protocol

Using a standardized protocol like JSON-RPC for the `postMessage` communication is an excellent idea and would bring significant benefits over a custom message structure.

**Why JSON-RPC is a Good Fit:**

*   **Standardization:** It provides a clear, battle-tested specification for remote procedure calls, eliminating ambiguity in message formats.
*   **Structured Communication:** It formalizes the concepts of requests (with IDs), responses (with results or errors), and notifications (fire-and-forget messages), which map perfectly to the needs of a host-plugin architecture.
*   **Bidirectional RPC:** Both the host and the plugin can expose an "API" to each other, allowing for true bidirectional communication (e.g., host calls `plugin.updateState()` and plugin calls `host.getTheme()`).
*   **Existing Libraries:** Leveraging existing JSON-RPC libraries can simplify the SDK development and ensure robust parsing and validation.

**Integration with `postMessage`:**

The SDK would abstract the protocol. Instead of manually creating message objects, developers would use SDK functions:

```typescript
// SDK in Main App (Host)
import { createRpcClient } from '@viz/sdk/rpc'; // Assuming SDK export
const pluginRpc = createRpcClient(iframeRef.contentWindow, pluginUrl);
const pluginApi = pluginRpc.getProxy<PluginApi>(); // Typed proxy object

// Call a method on the plugin
const result = await pluginApi.somePluginMethod({ data: 'foo' });

// The SDK internally constructs and sends a JSON-RPC request via postMessage
// { "jsonrpc": "2.0", "method": "somePluginMethod", "params": { "data": "foo" }, "id": 1 }
```

**Handling Complex Data Types (JSON-RPC Extension):**

The `postMessage` API has a powerful feature that JSON doesn't: the ability to transfer "Transferable" objects like `ArrayBuffer`, `MessagePort`, or `Blob` with near-zero copy cost, avoiding expensive serialization or Base64 encoding. We can create a lightweight extension to JSON-RPC to support this.

The process, managed by the SDK, would be:

1.  **Serialization:** Before sending a JSON-RPC message, the SDK scans the `params` (for requests) or `result` (for responses) for Transferable objects.
2.  **Substitution:** It replaces each Transferable object with a JSON-compatible placeholder and collects the actual objects in a separate `transferList`.

    ```javascript
    // Original params
    const params = {
        name: 'my-image.png',
        data: new ArrayBuffer(1024) // Not valid in JSON
    };

    // Substituted params for the JSON-RPC message
    const substitutedParams = {
        name: 'my-image.png',
        data: { "$type": "ArrayBuffer", "placeholderId": 0 }
    };
    const transferList = [params.data]; // The actual ArrayBuffer
    ```
4.  **Transmission:** The SDK sends the message using `postMessage`.

    ```javascript
    iframe.contentWindow.postMessage({
        protocol: 'json-rpc',
        payload: jsonRpcMessage, // contains substitutedParams
    }, targetOrigin, transferList);
    ```
5.  **Deserialization:** The receiving SDK parses the JSON-RPC payload. When it encounters a placeholder, it retrieves the corresponding object from the `MessageEvent`'s transferred items, reconstructing the original data structure before passing it to the target method.

This approach combines the structural benefits of JSON-RPC with the performance advantages of `postMessage`'s transfer mechanism, providing a robust and efficient communication channel for all sorts of data types.

**Outgoing Messages (to main app):**


```typescript
import { sendHostMessage } from '@viz/sdk/messaging'; // Assuming SDK export

function sendDataToMainApp(data: any): void {
    // sendHostMessage handles targetOrigin and type-checking via SDK
    sendHostMessage({ type: 'plugin-data-update', payload: data });
}
```

**Incoming Messages (from main app):**


```typescript
import { onHostMessage } from '@viz/sdk/messaging'; // Assuming SDK export
import { writable } from 'svelte/store';

export const pluginConfig = writable<any>({});

onHostMessage(window, (message) => {
    // message is already type-checked and origin-validated by SDK
    if (message.type === 'init-config') {
        pluginConfig.set(message.payload);
    } else if (message.type === 'theme-change') {
        // ... update theme
    }
    console.log('Message from main app:', message);
});
```

**Leveraging SDK for Views/Metadata:** The plugin's Svelte application will use SDK-defined types to describe its views and metadata, sending this information to the main application during an initial handshake.
```typescript
import { registerPluginView } from '@viz/sdk/registration'; // SDK for plugin registration

// Example of registering a view from within the plugin iframe app
registerPluginView({
    id: 'my-plugin-view-1',
    name: 'My Custom Plugin View',
    icon: 'path/to/icon.svg',
    // ... other metadata defined by SDK types
});
```

##### 6.2.3. Messaging Protocol (Defined and enforced by SDK)

The SDK will define and expose the interfaces for messages exchanged, thereby ensuring type safety.

**Main App to Plugin (SDK types):**


```typescript
interface InitConfigMessage { type: 'init-config'; payload: { initialData: any, theme: 'dark' | 'light' }; }
interface UpdatePropsMessage { type: 'update-props'; payload: { newProp: any }; }
// ... other message types
```

**Plugin to Main App (SDK types):**


```typescript
interface PluginReadyMessage { type: 'plugin-ready'; }
interface StateChangeMessage { type: 'state-change'; payload: { key: string, value: any }; }
interface RequestActionMessage { type: 'request-action'; payload: { actionName: string, args: any[] }; }
// ... other message types
```

#### 6.3. Integration with Existing Plan

**`RegisteredViewDefinition`:** The `component` field within `RegisteredViewDefinition` will now point to `IframePluginHost.svelte` for plugins that are iframe-based. The `pluginMeta` property will be used to store the `pluginUrl` and other iframe-specific details. Additionally, it will contain metadata, such as the plugin's exposed views and their properties, which will be communicated via the SDK.
```typescript
export interface RegisteredViewDefinition {
    // ... existing fields ...
    pluginMeta?: {
        pluginIdentifier: string;
        pluginViewIdentifier: string;
        type: 'iframe'; // Indicate an iframe-based plugin
        url: string;    // The URL where the iframe content is hosted
        // additional metadata about the view from plugin (e.g., via SDK handshake)
        pluginProvidedMetadata?: any;
    };
}
```

**`LayoutPersistence`:** Serialization will save the `pluginMeta.url`, `pluginIdentifier`, `pluginViewIdentifier`, and any `pluginProvidedMetadata`. This will enable the `IframePluginHost.svelte` component to be correctly instantiated and loaded when deserializing the layout.

#### 6.4. Pros and Cons of the Iframe Approach (with SDK)

**Pros:**

*   **Highest Isolation:** This approach provides the strongest security and environmental separation, isolating CSS, JavaScript, and the DOM to prevent conflicts and unauthorized access.
*   **Structured Development (Svelte + SDK):** Plugin developers are guided towards a consistent and type-safe development experience through the use of Svelte and the provided SDK.
*   **Type-Safe Communication:** The SDK ensures that all messages exchanged between the main application and plugins conform to defined interfaces, significantly reducing runtime errors.
*   **Independent Deployment:** Plugins can be developed, built, and deployed independently from the main application, offering greater flexibility.
*   **Crash Resilience:** A crash within a plugin's iframe typically does not affect the main application, enhancing overall stability.

**Cons:**

*   **Styling and Layout Challenges:** Achieving full integration in terms of theming, dynamic sizing, and context menus for iframes requires careful design and robust SDK support for coordination.
*   **Accessibility (A11y):** Managing focus, keyboard navigation, and ARIA attributes across iframe boundaries is more complex and necessitates dedicated SDK support.
*   **User Experience (UX):** Creating seamless cross-iframe interactions, such as drag and drop or shared tooltips, demands significant effort and effective SDK abstraction.
*   **Local Development Complexity:** Setting up a separate development server for each plugin can add to the complexity of the local development environment.

#### 6.5. Iframe Performance & Communication: Engineering for Seamless Integration

While iframes introduce inherent architectural considerations, particularly regarding performance and inter-context communication, these are engineering challenges to be actively managed rather than insurmountable blockers. Successful applications like Figma demonstrate that these aspects can be mitigated to provide a seamless user experience.

##### Performance Overhead: The Nature of Isolation

The primary source of performance overhead stems from the fundamental isolation iframes provide. Each `<iframe>` creates its own dedicated resources, leading to:

*   **Resource Duplication:** Every iframe instance has its own JavaScript engine, DOM tree, CSSOM, and often its own set of runtime libraries. In modern browsers, cross-origin iframes frequently operate in separate OS processes, which incurs significant system-level overhead for process management.
*   **Browser Context Setup:** Loading an iframe involves a complete document lifecycle, including network requests for its assets, HTML/CSS parsing, JavaScript execution, layout, and painting. This "mini-webpage" loading cost applies to each individual iframe.

**Mitigation Strategies (Figma as Example):**

*   **Strictly On-Demand Loading & Unloading:** The most impactful strategy involves loading iframe plugins *only* when explicitly invoked by the user. When a plugin is closed or becomes inactive, its iframe resources should be aggressively freed, for instance, by removing the iframe element from the DOM or setting its `src` to `about:blank`. Figma effectively uses this approach by loading plugins solely upon user activation and releasing resources upon closure.
*   **Lightweight Content:** The content within the iframe itself must be exceptionally lean and optimized. This includes minimizing JavaScript bundle size, employing efficient rendering techniques, and avoiding unnecessary dependencies.
*   **Strategic Pre-warming:** For a select few highly critical or frequently accessed plugins, consider "pre-warming" their iframes by loading them in the background while hidden during periods of application inactivity. This strategy requires caution to avoid negatively impacting the main application's initial load performance.

##### Communication Latency: Bridging Isolated Contexts

Communication between the main application and an iframe relies on `window.postMessage`, an inherently asynchronous mechanism. This process requires data serialization by the sender and deserialization by the receiver, consuming CPU cycles and introducing latency, making it slower than direct function calls or object passing within a single JavaScript context.

**Mitigation Strategies (Figma as Example):**

*   **Batching & Debouncing:** To reduce communication overhead, avoid "chatty" interactions. Aggregate multiple small data updates into a single, larger message sent periodically, often debounced or throttled.
*   **Send Only Deltas:** Instead of transmitting entire application or plugin states, communicate only the minimal necessary changes or "deltas."
*   **Optimized Protocol & Payloads:** Design a communication protocol that utilizes compact, structured data formats. Figma's plugin API likely employs highly optimized, well-defined data structures for efficient data exchange.
*   **Data Transfer (for large binary data):** For significant binary data, such as images, `postMessage` supports transferring (not just copying) `ArrayBuffer` objects, which can significantly reduce overhead, although this adds complexity.
*   **Asynchronous by Design:** The entire application and plugin architecture must be built with asynchronous communication in mind, anticipating and gracefully handling inherent latencies.

These strategies collectively enable the powerful isolation benefits of iframes to be realized without sacrificing critical application performance, a balance successfully struck by sophisticated web applications.

This refined approach combines the strong isolation benefits of iframes with the structured development and type safety advantages of an SDK, specifically for Svelte-based plugins.

