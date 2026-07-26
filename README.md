# CY_BORGER ⚡

A lightning-fast, cyberpunk-themed character generator and campaign manager built with Go, HTMX, and SQLite.

CY_BORGER is designed for the [CY_BORG](https://cy-borg.com/) tabletop roleplaying game. 

**[🔗 Try the Live Demo!](https://cy-borger.panther-bleak.ts.net/game/bb1605f7cb87)**

## Features

- **Character Generation & Management**: Instantly roll fully fleshed-out cyberpunk operators, edit their stats, track HP/glitches, and manage their inventory.
- **Campaign Management**: Create shared game lobbies where GMs can oversee active party members and manage their flatlined characters in the graveyard.
- **Real-Time Updates**: WebSockets provide instantaneous syncing for all players in a game session—when you take damage, the GM sees it instantly.
- **Blazing Fast**: Server-rendered HTMX architecture eliminates the need for heavy client-side JavaScript frameworks.
- **Persistent Data**: Lightweight, CGO-free `modernc.org/sqlite` database backend.
- **Docker Ready**: Fully containerized multi-stage build makes deployment a breeze.
- **Cyberpunk Aesthetic**: Highly thematic dark mode UI with neon accents and glitch elements.

## Getting Started

### Local Development

Prerequisites:
- Go 1.26+
- [Air](https://github.com/cosmtrek/air) (for live reloading)

```bash
# Clone the repository
git clone https://github.com/mrpoundsign/cy_borger.git
cd cy_borger

# Run with Air
air
```

The application will be available at `http://localhost:8080`.

### Docker Deployment

You can quickly spin up the full stack using Docker Compose:

```bash
docker compose up -d --build
```

This will run the CY_BORGER server on port `8080` (or `9090` depending on your `docker-compose.yml` config) and mount a local volume to persist the SQLite database.

## Architecture

- **Backend**: Go `net/http` standard library.
- **Database**: SQLite (via `modernc.org/sqlite`).
- **Frontend**: Go HTML templates with HTMX for interactivity.
- **Tests**: Playwright is used for end-to-end E2E testing.

## License

CY_BORGER is released under the GNU General Public License v3.0 (GPLv3). See the `LICENSE` file for details.
