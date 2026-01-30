# RFC 001: Server-to-Server Interoperability (Viz Protocol)

| Field            | Value                                            |
| ---------------- | ------------------------------------------------ |
| **RFC**          | 001                                              |
| **Title**        | Server-to-Server Interoperability (Viz Protocol) |
| **Author**       | Les                                              |
| **Status**       | Draft                                            |
| **Created**      | 10-01-2026                                     |
| **Last Updated** | 30-01-2026                                       |

## Abstract

This document specifies a standardized protocol for communication and interoperability between image server instances. It defines a URI scheme (`viz://`) for resource identification, a federated identity model, and a set of core capabilities including remote viewing, cross-server search, and metadata synchronization.

The protocol aims to bridge the gap between enterprise Digital Asset Management (DAM) systems, professional photography workflows, and personal archiving solutions, enabling a decentralized yet unified user experience.

## 1. Introduction

The world and industry around digital/image asset management is one with a long history of established practices, wide-scale adoption of different standards across companies from camera manufacturers to media agencies. There are decades-long established practices of organizations (corporate/commercial or otherwise internal) coming up with proprietary solutions that are either closed-off in their own ecosystem or require money somewhere in the chain of operability to have successful interoperability.

As the world of photography has become very accessible through affordable beginner cameras, mobile phone cameras with good-to-professional features, and the average semi-professional photographer having to deal with "EXIF" from time-to-time, the gap in digital asset management needs to be bridged with long-term, open-source solutions. These solutions should allow anyone—from long-established professionals to beginner photographers—to organize and share their media without hassle.

This protocol proposes a solution tailored for high-fidelity image management, inspired by federation in the Fediverse but optimized for the specific requirements of visual assets.

### 1.1 Requirements Language

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in BCP 14 [RFC2119] [RFC8174] when, and only when, they appear in all capitals, as shown here.

### 1.2 Motivation and Philosophy

On a personal note, my frustrations with the need to currently have at least 3-4 different bits of software, from heavy and bloated native software taking up loads of resources just to view, rate, and transcode some photos, to the disjointed experience of sharing those photos with clients or syncing them across devices with various cloud providers, is the primary motivator for this protocol. I want a more unified, lightweight, and standard way to handle images that respects the data, workflow, and privacy of those doing the work. I am a prideful person but I want to rid the world of "not-invented-here-iritis" that fragmented my ability to label and edit and send a couple of photos in a few minutes.

One of my ideas based off this protocol is to bridge a "Enterprise/Commercial-to-Consumer" divide. There is a future where a freelance photographer running a personal instance (e.g., Immich or a small `viz` setup) can seamlessly connect with a massive Media Agency's archival system. With minimal configuration, they could authorize the agency to pull high-res assets directly from their "rogue" server, or conversely, mirror a specific "Portfolio" collection from their heavy-duty work server to a lightweight, consumer-friendly instance for sharing with family on mobile, "quick and dirty" sharing to social media or sharing solo missing files with clients without having to reconnect to an enterprise system. This interoperability shouldn't be the privilege of expensive enterprise contracts; it's also about giving user choice.

## 2. Resource Identification

### 2.1 The `viz` URI Scheme

This protocol defines the `viz` Uniform Resource Identifier (URI) scheme to locate resources across the federated network.

> **Note:** The scheme name `viz` is currently a placeholder. As this protocol is designed to be project-agnostic and not tied to any specific application, a final, neutral scheme name will be selected in a future revision.

*   **Syntax:** `viz://<authority>/<resource-type>/<resource-id>`
*   **Authority:** The domain name of the hosting server (e.g., `photos.example.com`).
*   **Resource Type:** Identifies the collection type. Standard types include `images`, `collections`, and `albums`.
*   **Resource ID:** The unique identifier of the asset on the host server.

**Example:** `viz://photos.example.com/images/12345`

### 2.2 Resolution
Clients and servers MUST implement resolution logic to translate these URIs into actionable API endpoints. Resolution SHOULD rely on standard discovery mechanisms (see Section 4).

## 3. Identity and Authentication

### 3.1 Remote User Identity
Users SHALL be identified by a handle in the format `username@domain`. This allows for a decentralized identity model.

### 3.2 Server Trust
Servers MUST establish a trust relationship before exchanging protected data.
*   **Handshake:** Servers SHOULD perform a handshake using public key cryptography or shared secrets to establish mutual trust.
*   **Verification:** Servers MUST verify the identity of peer servers during this handshake.

### 3.3 Access Control
*   **Federated Authentication:** When a user from Server A accesses Server B, Server B MUST authenticate the request via Server A using standard mechanisms (e.g., OIDC or signed requests).
*   **Guest Access:** To support client portals, servers SHOULD support federated guest identities (e.g., email-based). This allows a guest (e.g., `client@example.com`) to authenticate once and access collections shared by multiple independent servers (Photographer A and Photographer B) in a unified view.

## 4. Metadata and Standards Compliance

To prevent data loss and ensure robust interoperability, this protocol mandates that embedded metadata is the **Source of Truth** for asset information.

### 4.1 Embedded First
When transferring or syncing images, the recipient server MUST prefer XMP/IPTC data embedded in the file over sidecar data or API properties, unless specific overrides are strictly requested.

### 4.2 Metadata Preservation
Servers acting as relays, proxies, or caches MUST NOT strip XMP/EXIF/IPTC metadata unless explicitly configured for privacy (e.g., stripping GPS data for public guests).

### 4.3 Mapping and Negotiation
Servers MUST standardize mapping between API fields and XMP/IPTC fields (e.g., `api.creator` <-> `XMP:dc:creator`).
*   **Discovery:** Servers MUST expose their metadata profiles via `/.well-known/viz/metadata-profile`.
*   **Respecting Authority:** Importing servers SHOULD respect the origin's metadata mapping intention where possible.

### 4.4 Rights Management
Usage Rights and Copyright notices MUST be read from `XMP:Rights` and `IPTC:CopyrightNotice`. Servers MUST enforce federation policies (e.g., "Do not distribute") based on these fields.

## 5. Core Capabilities

### 5.1 Remote Viewing
Servers SHOULD support remote viewing, allowing clients to browse collections on another server without full ingestion.
*   **Implementation:** This SHOULD be achieved via efficient thumbnail/preview fetching and on-demand streaming of high-resolution assets.

### 5.2 Cross-Server Search
Queries MAY be federated to trusted peers. Results from multiple servers SHOULD be aggregated by the client or the requesting server.

### 5.3 Copy, Import, and Sync
*   **Forking:** The protocol MUST support "forking" an image from a remote server to a local library while preserving attribution.
*   **Syncing:** The protocol SHOULD support keeping a collection in sync across two servers (e.g., "Backup" vs. "Live").

## 6. Data Transport and Delivery

To ensure high performance and reliability when transferring large media assets (e.g., RAW photos, high-res TIFFs), the protocol recommends a layered transport strategy.

### 6.1 Baseline Transport
Servers MUST support standard HTTP/HTTPS range requests. This ensures compatibility with all clients and allows for basic resumable downloads and seeking within large files.

### 6.2 Resumable Uploads
For large file ingest (e.g., from a photographer to an agency), servers SHOULD implement a resumable upload protocol, such as **TUS**. This prevents data loss during unstable network conditions and allows transfers to span multiple sessions.

### 6.3 Modern Transport
To minimize latency and improve congestion control, especially for mobile clients or high-latency links:
*   Servers MAY support **HTTP/3 (QUIC)** to reduce connection setup times and improve performance on unreliable networks.
*   For real-time scenarios (e.g., "Remote View" streaming), servers MAY evaluate **WebTransport** or similar low-latency streams to deliver tile data or progressive JPEGs without head-of-line blocking.

### 6.4 Optimized Delivery
Inspired by modern streaming architectures, servers SHOULD optimize delivery for "Remote Viewing":
*   **Segmentation:** Large assets MAY be virtually segmented (e.g., using IIIF tiling or byte-range chunks) to allow clients to fetch only the visible portion of an image.
*   **On-the-Fly Encoding:** Servers MAY implement just-in-time transcoding to deliver the optimal format (AVIF/WebP) and resolution based on the client's viewport and network conditions, caching the result for subsequent requests.

## 7. Client Ecosystem and Native Clients

The protocol allows for a rich ecosystem of native applications (desktop, mobile, CLI) that can interact with *any* compliant server. This decouples the "viewing/editing experience" from the "hosting provider."

### 7.1 Bring Your Own Client (BYOC)
Users SHOULD be able to choose their preferred photo manager (e.g., a lightweight native viewer, a RAW editor, or a mobile gallery app) and simply point it at their data source. The protocol ensures consistent access to albums, search, and metadata regardless of the client implementation.

### 7.2 Offline-First and Syncing
Native apps SHOULD leverage the protocol's sync capabilities to cache high-performance thumbnails and metadata locally. This allows for browsing and searching gigabyte-sized libraries with zero latency, even when offline, syncing changes back when connectivity is restored.

### 7.3 Direct Editing Workflow
Advanced editors SHOULD support opening high-resolution assets directly from the server (via signed, short-lived URLs) and saving edits back as new versions or sidecar files.

**Example Workflow:**
1.  **Context:** A Media Agency uses a Proprietary CMS. A Photographer uses Photoshop.
2.  **Implementation:** A Custom Plugin inside Photoshop connects directly to the Agency's Protocol API.
3.  **Action:** The user searches and browses the remote library without leaving the editor.
4.  **Edit:** The user selects an asset; the plugin fetches it (handling proxy vs. high-res logic transparently).
5.  **Save:** On save, the plugin pushes the updated file (or version) back to the Protocol API.
6.  **Result:** The central library is immediately current, bypassing manual file transfers.

## 8. Event Federation

To maintain synchronization across servers, servers MAY implement an active event notification system.

*   **Mechanism:** Servers MAY implement a webhook subscription model (e.g., WebSub) or similar push-based notification system.
*   **Payload:** Webhook payloads RECOMMENDED to be signed (HMAC) using a shared secret.
*   **Standard Events (Proposal):**
    *   `asset.updated`: Metadata or file content has changed.
    *   `asset.deleted`: The asset has been removed or unshared.
    *   `rights.changed`: Usage rights have been updated
    *   `collection.items_added` / `collection.items_removed`.

## 9. Protocol Extensions

To ensure the protocol remains lightweight while supporting specialized use cases, an extension mechanism is defined.

*   **Constraint:** Extensions MUST NOT be used to create "walled gardens" or proprietary lock-in.
*   **Discovery:** Extensions MUST be declared in the server's public capability manifest (e.g., `/.well-known/viz/capabilities`).
*   **Additive Only:** Extensions MUST be strictly additive. A client that does not understand an extension MUST still be able to perform core actions: View, Search, and Download.
*   **Namespacing:** Extensions MUST use reverse-domain notation (e.g., `com.example.features.ai-tagging`).

## 10. Client Libraries and OpenAPI

### 10.1 Standard Libraries
Libraries SHOULD be developed for common programming languages to reduce friction for adoption. These libraries MUST implement:
*   **Parsing:** Validating and decomposing `viz://` URIs.
*   **Discovery:** Automatically fetching `/.well-known` configurations.
*   **Security:** Verifying TLS certificates and signatures.

### 10.2 OpenAPI Usage
The protocol SHOULD utilize OpenAPI specifications for:
1.  **Dynamic Client Generation:** allowing clients to generate type-safe bindings at runtime.
2.  **Validation:** allowing intermediaries to validate payloads.
3.  **Mocking:** providing a reference document for compliance testing.

## Appendix A: Example Use-Cases

### A.1 Manifold Release 9 Integration
A GIS image server like Manifold Release 9 could integrate by exposing resources via `viz://` URIs, mapping internal identifiers to protocol IDs. It could participate in federated search, allowing GIS imagery to be discoverable alongside standard photography, while acting as a gateway for other WMS servers.

### A.2 Omeka S Image Server Integration
Omeka S (IIIF-compliant) aligns well with the "Remote Viewing" capabilities. Integration would involve mapping Dublin Core metadata to the protocol's XMP/IPTC standards, ensuring archival data is preserved when accessed via the federated network.

### A.3 The Editorial Supply Chain (Photographer -> Agency -> Media House)
This scenario demonstrates a multi-hop asset lifecycle enabled by the protocol:

1.  **The Photographer (Source):** A freelance photographer ingests RAW photos into their personal protocol-compliant server (e.g., a self-hosted instance). They tag assets and apply "Exclusive" usage rights in the embedded XMP.
2.  **The Agency (Aggregator):** The photographer grants the Agency's server access to a specific "Assignment" collection. The Agency's server "forks" (imports) the high-res assets. The protocol ensures the photographer's original IPTC credit and rights data are preserved during this transfer.
3.  **The Media House (Consumer):** A newspaper editor, using their organization's CMS (which has a protocol-compliant plugin), browses the Agency's catalog. They license an image. The Agency's server delivers the asset to the Media House's server.
4.  **Updates & Revocation:** Later, the photographer realizes a caption error. They update it on their server. An `asset.updated` event propagates to the Agency, which reviews and accepts the change, subsequently notifying the Media House. Conversely, if rights expire, a `rights.changed` event ensures the Media House's system flags the asset as "Do Not Publish."

### A.4 The Guardian's "The Grid" (Enterprise DAM Integration)
[The Grid](https://github.com/guardian/grid) is The Guardian's open-source image management system, capable of handling millions of images with sophisticated rights management and usage tracking. Integration with the Viz Protocol would enhance its capabilities in the following ways:

*   **Federated Ingestion:** Currently, The Grid ingests images from wire agencies and internal uploads. By implementing the protocol, The Grid could treat freelance photographers' personal nodes as "remote agencies." Instead of manually uploading files, a photographer simply shares an `viz://` collection link. The Grid then crawls and indexes these remote assets without immediate heavy transfer, pulling high-res files only when an image is selected for publication.
*   **Dynamic Rights Synchronization:** The Grid excels at tracking usage rights (e.g., "expires in 30 days"). The protocol's Event Federation (`rights.changed`) would allow The Grid to automatically update an asset's status if the rights holder (photographer or agency) modifies terms on their own server, reducing the risk of accidental copyright infringement.
*   **Unified Search:** The Grid's powerful search could federate queries to trusted partner archives (e.g., historical societies running protocol-compliant servers), presenting a unified search result page to picture editors that mixes internal holdings with external, licensable content.

## Next Steps
1.  Define the URI resolution spec (currently using the `viz` placeholder).
2.  Establish a controlled vocabulary and definitions.
3.  Finalize the formal protocol name (MUST NOT be tied to any specific project or application).
4.  Evaluate transport protocols (QUIC, TUS) for performance and resilience.
5.  Prototype a simple "remote view" component in the frontend.
6.  Implement basic server handshake.
7.  Implement basic client-to-server communication.
