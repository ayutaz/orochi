# Orochi - Modern BitTorrent Client

[![CI](https://github.com/ayutaz/orochi/actions/workflows/ci.yml/badge.svg)](https://github.com/ayutaz/orochi/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/ayutaz/orochi)](https://goreportcard.com/report/github.com/ayutaz/orochi)
[![Release](https://img.shields.io/github/release/ayutaz/orochi.svg)](https://github.com/ayutaz/orochi/releases/latest)

A modern, secure, and user-friendly BitTorrent client with a beautiful web interface, written in Go and React.

## Features

- 🚀 **Modern Web UI** - Beautiful, responsive interface built with React and Material-UI
- 🔄 **Real-time Updates** - Live progress tracking via WebSocket
- 🎯 **Smart File Selection** - Choose which files to download within torrents
- 🌍 **Cross-platform** - Works on Windows, macOS, and Linux
- 🔒 **Security First** - VPN binding, kill switch, and authentication support
- 📦 **Single Binary** - No dependencies, easy deployment
- 🌓 **Dark Mode** - Easy on the eyes during late-night downloads
- 🌐 **Multi-language** - Support for English and Japanese
- 📊 **Comprehensive API** - RESTful API with OpenAPI documentation
- ⚡ **High Performance** - Efficient concurrent downloads with piece prioritization

## Installation

### From Binary

Download the latest release for your platform from the [releases page](https://github.com/ayutaz/orochi/releases).

### From Source

Requirements:
- Go 1.23 or later
- Node.js 20 or later (for building the web UI)

```bash
git clone https://github.com/ayutaz/orochi.git
cd orochi
make build
```

## Quick Start

```bash
# Start with default settings
./orochi

# Start on a different port
./orochi --port 9000

# Specify download directory
./orochi --download-dir /path/to/downloads

# Show version information
./orochi --version
```

Then open http://localhost:8080 in your browser.

### Using Real BitTorrent Mode

By default, Orochi runs in stub mode for testing. To use actual BitTorrent functionality:

```bash
./orochi --real
```

⚠️ **Legal Notice**: Only download and share content you have the legal right to access.

## Configuration

Orochi can be configured through:
- Command-line flags
- Web interface settings
- Configuration file (auto-created on first run)

Key settings:
- **Download Directory**: Where to save downloaded files
- **Port**: BitTorrent listen port (default: 6881)
- **Max Connections**: Maximum peer connections
- **Speed Limits**: Upload/download speed restrictions
- **VPN Binding**: Restrict traffic to specific network interface

## Development

This project follows Test-Driven Development (TDD) principles.

### Prerequisites

- Go 1.23 or higher
- Make (optional)

### Running Tests

```bash
make test
```

### Running with Coverage

```bash
make coverage
```

## Documentation

### User Guide

For detailed usage instructions, please see the [User Guide](docs/USER_GUIDE.md) (日本語).

### API Documentation

Once the server is running, you can access the interactive API documentation at:

```
http://localhost:8080/api-docs
```

The API documentation is generated using OpenAPI 3.0 specification and provides:
- Complete API reference with all endpoints
- Request/response schemas
- Interactive API testing interface (Swagger UI)
- WebSocket endpoint documentation

## Legal Notice

This software is designed for downloading and sharing files using the BitTorrent protocol. The use of this software for downloading or distributing copyrighted material without permission is illegal. Users are responsible for complying with local laws and regulations.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Implement your changes
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Acknowledgments

- Inspired by BitThief's simplicity
- Built with [anacrolix/torrent](https://github.com/anacrolix/torrent)