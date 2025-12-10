# Image Server Protocol

| Field            | Value                                                |
| ---------------- | ---------------------------------------------------- |
| **RFC**          | 001                                                  |
| **Title**        | Server-to-Server Interoperability (Imagine Protocol) |
| **Author**       | Les                                                  |
| **Status**       | Draft                                                |
| **Created**      | 2025-12-10                                           |
| **Last Updated** | 2025-12-10                                           |

## Overview
This RFC proposes a standardized protocol for enabling communication and interoperability between different image server instances. The goal is to allow users on one server to interact with content and users on another, similar to federation in the Fediverse but tailored for high-fidelity image management.

This is a somewhat optimistic yet simple draft document meant outline sketch ideas that can be built upon. These ideas are based on long-established principles of designing open-protocols and is meant to be a modern and extensible protocol for the future. This protocol will likely NOT be called `imagine`, that is just the current name of this project, which in itself is likely to change to `viz` in the near-future.

## Core Concepts

### 1. The `imagine://` Protocol
A custom URL scheme to identify resources across servers.
- **Format:** `imagine://<server-domain>/<resource-type>/<resource-id>`
- **Example:** `imagine://photos.example.com/images/12345`
- **Resolution:** Clients (web/mobile) and servers can resolve these URLs to API endpoints.

### 2. Identity & Authentication
- **Remote User Identity:** Users are identified by `username@domain`.
- **Server Trust:** Servers perform a handshake (likely via shared keys or public key cryptography) to establish trust.
- **Access Control:** When a user from Server A accesses Server B, Server B authenticates the request via Server A (OIDC or signed requests).

### 3. Key Capabilities
- **Remote Viewing:** Browse collections on another server without downloading everything.
- **Cross-Server Search:** Queries can be federated to trusted peers.
- **Copy/Import:** "Forking" an image from a remote server to your local library with attribution preservation.
- **Syncing:** Keeping a collection in sync across two servers (e.g., a "Backup" server and a "Live" server).

### 4. Guest & Client Access
A critical use case is a Client working with multiple Photographers/Agencies (each with their own protocol-compliant server). The Client needs a unified view without managing multiple accounts.

- **The "Client Portal" Mode:** A lightweight, unbranded (or co-branded) view where a guest can see collections shared with them from *multiple* sources.
- **Federated Guest Identity:** 
    - A guest can be invited via email. That email acts as a "soft identity" across servers.
    - **Example:** Photographer A shares a collection with `client@example.com`. Photographer B shares a collection with `client@example.com`.
    - When the client logs into *any* protocol-compliant instance (or a standalone Client Portal app), they see a unified dashboard of assets from Photographer A and B.
- **Unified Lightbox:** The client can select photos from Server A and Server B and download them in a single batch operation (behind the scenes, the browser or portal fetches from respective servers).

### 5. Metadata & Standards Compliance
To prevent data loss and ensure robust interoperability, the protocol treats embedded metadata as the **Source of Truth** for asset information. This reduces reliance on proprietary API JSON structures for core data.

- **Embedded First:** When transferring or syncing images, the recipient server MUST prefer XMP/IPTC data embedded in the file over sidecar data, unless specific overrides are requested.
- **Standardized Mapping:** The protocol defines a strict mapping between API fields and XMP/IPTC fields (e.g., `api.creator` <-> `XMP:dc:creator`).
- **Preservation:** Servers acting as "relays" or "caches" MUST NOT strip XMP/EXIF/IPTC metadata unless explicitly configured for privacy (e.g., stripping GPS data for public guests).
- **Rights Management:** Usage Rights and Copyright notices must be read from `XMP:Rights` and `IPTC:CopyrightNotice` to enforce federation policies (e.g., "Do not distribute").

#### Schema Negotiation & Custom Mappings
Agencies and long-standing archives often have legacy metadata schemas (e.g., mapping `Photographer` to `IPTC:SpecialInstructions` instead of `IPTC:By-line`). The protocol supports this via negotiation:

1.  **Capability Discovery:** Servers expose their metadata profiles via `/.well-known/imagine/metadata-profile`. This JSON/YAML/XML document describes how the server reads/writes fields.
2.  **Respecting Authority:** When importing an asset from a remote server, the importing server should respect the *origin's* metadata mapping intention where possible, or at minimum, preserve the raw XMP packet to ensure no data is lost during "translation" to the local schema.
3.  **"Pass-Through" Handling:** If a server acts as a gateway or portal for another agency's server, it must present the metadata *as defined by the source*, rather than re-interpreting it through its own default mapping rules.

#### Client/Language Libraries for Resolution
Libraries for different programming languages will have to have their own spec-complaint libraries to help handle and resolve things. Using the [atproto](https://atproto.com/) and [Bluesky's various packages](https://github.com/bluesky-social/indigo) as an example, this greatly helps reduce friction in making legacy systems interop with the new protocol and can help encourage adoption among systems. This is to better help organisations with legacy code adapt quickly to modernising and standardising their code with wrappers essentially and expose those as APIs.

To achieve this, these standard libraries must implement the following core capabilities:

-   **Parsing:** Validating and decomposing `imagine://` URIs into their constituent parts (Authority, Collection, Asset ID).
-   **Discovery:** Automatically fetching `/.well-known/imagine` from the target domain to locate API endpoints and metadata capabilities.
-   **Security:** Verifying TLS certificates and handling signature verification for "trusted" server-to-server communication.
-   **Caching:** Implementing sensible caching strategies for resolved metadata to reduce network chatter and improve responsiveness.
-   **Adapters (Optional):** Providing interfaces for common frameworks (e.g., Express/Next.js middleware, Go `net/http` handlers, Django views) to easily "mount" the protocol handler.

### 6. Protocol Extensions
To ensure the protocol remains lightweight while supporting specialized use cases (e.g., AI tagging pipelines, print-specific color workflows, rights management documentation), it supports an extension mechanism.

**CRITICAL MANDATE:** Extensions MUST NOT be used to create "walled gardens" or proprietary lock-in.

-   **Public Declaration:** Any server implementing an extension MUST declare it in its public capability manifest (e.g., `/.well-known/imagine/capabilities`).
-   **Additive Only:** Extensions must be strictly additive. A client that does not understand an extension must still be able to perform core actions: View, Search, and Download the primary asset.
-   **Open Specification:** If an extension is used for interoperability between different organizations, its specification MUST be publicly documented. "Private" extensions are only permitted for internal-only workflows that do not cross organizational boundaries. An organisation using certain extensions for internal tools **MAY NOT** use those extensions to prevent external interoperability.
-   **Namespacing:** Extensions must use reverse-domain notation (e.g., `com.example.features.ai-tagging`) to prevent collisions. The `org.imagine.*` namespace is reserved for official protocol standards.

### 7. Event Federation (Proposal)
To maintain synchronization across servers (e.g., when an agency updates caption data or revokes usage rights), the protocol suggests an active event notification system. This section outlines a *proposed* mechanism for real-time updates, but specific event types and delivery methods are open for discussion. Events don't have to necessarily be standardised, events should ideally be documented so that servers can opt into which ones they wish to support.

-   **Mechanism:** Compliant servers SHOULD consider implementing a webhook subscription model (WebSub maybe) or similar push-based notification system.
-   **Example Events:**
    -   `asset.updated`: Metadata or file content has changed.
    -   `asset.deleted`: The asset has been removed or unshared.
    -   `rights.changed`: Critical event. Usage rights have been updated (e.g., expiration).
    -   `collection.items_added` / `collection.items_removed`: For keeping shared folders in sync.
-   **Security:** To ensure trust, it is RECOMMENDED that webhook payloads be signed (HMAC) using a shared secret.
-   **Resilience:** Senders SHOULD implement retry logic for failed deliveries.

### 8. Native Client (Idea, not a spec thing)
The protocol is designed not just for server-to-server communication, but to enable a rich ecosystem of native applications (desktop, mobile, CLI) that can interact with *any* compliant server. This decouples the "viewing/editing experience" from the "hosting provider."

Organisations/Agencies could adjust their current workflows to adapt to this protocol. 

For example:

`Media agency -> Propriatery CMS API -> InDesign -> Photoshop/Lightroom - Back to CMS` 

could be:

`Media agency -> Protocol API -> External Editor (Photoshop) with Custom Internal Plugin -> Protocol API (Update)`

In this improved workflow:
1.  The **Custom Plugin** (inside Photoshop/InDesign) connects directly to the Protocol API, allowing the user to search and browse the remote library without leaving the editor.
2.  The user selects an asset, which the plugin fetches (handling high-res vs. proxy logic transparently).
3.  Edits are performed natively in **Photoshop**.
4.  On save, the **Plugin** pushes the updated file (or version) back to the Protocol API, ensuring the central library is immediately current.

-   **Bring Your Own Client (BYOC):** Users can choose their preferred photo manager (e.g., a lightweight native viewer, a RAW editor, or a mobile gallery app) and simply point it at their data source. The protocol ensures consistent access to albums, search, and metadata regardless of the client implementation.
-   **Offline-First & Syncing:** Native apps can leverage the protocol's sync capabilities to cache high-performance thumbnails and metadata locally. This allows for browsing and searching gigabyte-sized libraries with zero latency, even when offline, syncing changes back when connectivity is restored.
-   **Direct Editing:** Advanced editors can open high-resolution assets directly from the server (via signed, short-lived URLs) and save edits back as new versions or sidecar files, preserving the non-destructive workflow central to professional photography.

## Technical Requirements
- **API Extension:** New endpoints for federation (e.g., `/.well-known/imagine`).
- **Metadata Standardization:** strict adherence to IPTC/XMP to ensure metadata survives transfer.

### Leveraging OpenAPI
Beyond simple documentation, the OpenAPI specification is a functional component of the protocol ecosystem:

1.  **Dynamic Client Generation:** "Client Portal" applications can fetch the schema from a target server at runtime to generate type-safe bindings. This allows a client to interact with servers running different versions of the protocol without breaking.
2.  **Validation Proxies:** Intermediary relays can use the spec to validate payloads (e.g., metadata structure) before forwarding requests, protecting destination servers from malformed data.
3.  **Extension Discovery:** A server's capability manifest can link to partial OpenAPI specs for its active extensions. Smart clients can ingest these partial specs to "learn" new features (e.g., a "3D Model Viewer" extension) dynamically.
4.  **Mocking & Compliance:** The official protocol spec provides a "reference OpenAPI document" that developers can use to spin up mock servers, enabling them to test their implementation's compliance without needing a live peer.

> **NOTE:** This is more of a "what's possible idea" and absolutely not gauranteed to make it into the spec at all. 

## Before-End
The world and industry around digital/image asset management is one with a long history of established practices, wide-scale adoption of different standards across companies from camera manufacture's to media agencies. There are decades long-established practices of organisations (corporate/commercial or otherwise internal) coming up with proprietary solutions that are either closed-off in its own ecosystem or require money somewhere in the chain of operability to have successful interoperability.

As the world of photography has become very accessible through affordable beginner cameras, mobile phone cameras with good-to-professional features and the average semi-professional photographer having to deal with "EXIF" from time-to-time, the gap in digital asset management needs to be bridged with long-term, open-source solutions that allow anyone from long-established professionals to beginner photographers to organise and share their media without hassle.

On a personal note, my frustrations with the need to currently have at least 3-4 different bits of software, from heavy and bloated native software taking up loads of resources just to view, rate and transcode some photos, to the disjointed experience of sharing those photos with clients or syncing them across devices with various cloud providers, is the primary motivator for this protocol. I want a more unified, lightweight, and standard way to handle images that respects the data, workflow and privacy of those doing the work. I am a prideful person but I want to rid the world of "not-invented-here-iritis" that fragmented my ability to label and edit and send a couple of photos in a few minutes.

## Next Steps
1. Define the `imagine://` resolution spec.
2. Establish a vocabulary and definitions
3. Finalise protocol name (must NOT be tied to any specific project. e.g., [atproto](https://atproto.com/) is designed by Bluesky but is not name associated)
4. Prototype a simple "remote view" component in the frontend.
5. Implement basic server handshake.
6. Implement basic client-to-server communication.
