# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of Orochi BitTorrent client
- Modern React-based web interface with Material-UI
- Real-time updates via WebSocket
- Drag & drop support for multiple torrent files
- Magnet link support
- File selection within torrents
- Download/upload speed monitoring
- Progress tracking with visual indicators
- Dark mode support
- Japanese/English language support
- VPN binding functionality with kill switch
- Bearer token authentication
- REST API with OpenAPI documentation
- Cross-platform support (Windows, macOS, Linux)
- Automatic CI/CD with GitHub Actions
- Comprehensive test coverage

### Security
- CORS protection for WebSocket connections
- File upload size limits
- Path traversal protection
- Authentication for API endpoints

### Technical Features
- Built with Go and React/TypeScript
- Uses anacrolix/torrent library for BitTorrent protocol
- SQLite database for persistence
- Embedded web UI and API documentation
- Structured logging with zap
- Metrics collection support
- Performance monitoring

[Unreleased]: https://github.com/ayutaz/orochi/compare/v0.1.0...HEAD