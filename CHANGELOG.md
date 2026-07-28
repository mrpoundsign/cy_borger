# Changelog

All notable changes to this project will be documented in this file.


## [0.1.1] - 2026-07-28

### Added
- **REAL-TIME OFF-PAGE WEBSOCKET UPDATES**: Standalone character sheets (`/character/:id`) now connect to the game WebSocket topic when in an active game, automatically updating HP, stats, and flatline/revive banners live when modified off-page (e.g. by the GM).
- **WEBSOCKET AUTO-RECONNECT & STATUS BANNER**: Added automatic reconnect with exponential backoff and a persistent status banner (`⚡ DISCONNECTED - RECONNECTING TO TERMINAL...`) when WebSockets drop on the game screen.
- **WEBSOCKET CONNECTION REUSE**: Prevented redundant WebSocket connections by reusing existing game sockets (`window.cyGameSocket`) when inspecting character sheets inside modals on the GM screen.
- **MULTI-USER E2E PLAYWRIGHT INTEGRATION TESTS**: Added `tests/test_ws_offpage_updates.js` to verify real-time cross-browser WebSocket event streaming and UI updates between GM and Player contexts.
- **SQLITE JSON DATA MIGRATION**: Created migration `000003_migrate_abilities_to_stats.up.sql` to cleanly transform legacy JSON `abilities` maps into explicit integer `stats`.

### Changed
- **REFACTORED CHARACTER STATS**: Converted character stats from arbitrary string maps to a strongly-typed `Character.Stats` struct with explicit `int` fields (`Strength`, `Agility`, `Presence`, `Toughness`, `Knowledge`) and method API (`List()`, `Get()`, `Set()`).
- **SIMPLIFIED STAT FORMATTING**: Standardized stat modifier display in templates using `fmt.Sprintf("%+d", value)`.
- **CSS HEADER CONSOLIDATION**: Extracted inline `style="..."` attributes on section titles into a centralized, reusable `.box-header-title` class in `style.css`.
- **PREVENTED STALE BROWSER CACHING**: Configured `Cache-Control: no-cache, no-store, must-revalidate` on character sheet handlers to prevent browsers from serving stale 304 Not Modified responses.

### Fixed
- **DETERMINISTIC STAT ORDER**: Replaced random map iteration order with a fixed canonical ordering (`Strength`, `Agility`, `Presence`, `Toughness`, `Knowledge`).
- **HIGH-CONTRAST "YOU" BADGE**: Added missing `.bg-accent` CSS class to ensure the `⚡ YOU!` badge renders high-contrast black text on neon yellow.
- **HTMX REDIRECT AFTER GAME JOIN**: Updated `handleJoinGame` to return `HX-Redirect` headers on HTMX requests.
- **GM STAT BROADCAST**: Fixed GM stat update broadcasts to include formatted stat values in the activity log.

## [0.1.0] - 2026-07-27

### Added
- **EMBEDDED NEON CARD HEADERS**: Embedded section headers (`YOU ARE <NAME>`, `STATS`, `CLASS`, `GEAR & ARSENAL`, `CYBERTECH & APPS`) directly into card top-border lines for an authentic CY_BORG terminal aesthetic.
- **ZERO-OVERLAP CONTROLS**: Complete overhaul of section controls (`EDIT`, `DELETE`) with smart internal padding and flex headers, eliminating text and border overlaps across all display sizes.
- **RESPONSIVE MOBILE UI**: Eradicated side-margin bloat and horizontal scroll overflow on mobile phone screens (375px+ viewports), featuring auto-scrolling Arsenal tables and responsive grid column stacking.
- **INTEGRATED VITALS & GLITCHES CARD**: Dedicated Vitals management block with real-time inline inputs (`HP`, `Glitches`, `Credits`) and edit overlays for max stat customization.
- **STREAMLINED OPERATOR TOOLBAR**: Pruned cluttered action buttons (`COPY EDIT LINK`, `[EDIT MODE]`, `REROLL NEW`) for a hyper-focused operator UI.

## [0.0.2-dev] - 2026-07-26

### Added
- Real-time Activity Log in the Game View that automatically streams and appends new events (HP updates, inventory changes, operator deaths) without refreshing or losing scroll position.

## [0.0.1] - 2026-07-26

### Added
- Initial release of CY_BORGER.
