# Orochi - Simple Torrent Client

A simple, cross-platform torrent client written in Go, inspired by BitThief's simplicity.

## Features

- 🚀 Simple and intuitive web-based UI
- 🔒 Security-first design with VPN binding support
- 📦 Single binary distribution (no dependencies)
- 🌍 Cross-platform: Windows, macOS, Linux
- 🎯 Focused on legal torrent usage (Linux ISOs, open-source software)

## Installation

### From Binary

Download the latest release for your platform from the [releases page](https://github.com/ayutaz/orochi/releases).

### From Source

```bash
git clone https://github.com/ayutaz/orochi.git
cd orochi
make build
```

## Usage

Simply run the binary:

```bash
./orochi
```

Then open your browser and navigate to `http://localhost:8080`

## Development

This project follows Test-Driven Development (TDD) principles.

### Prerequisites

- Go 1.21 or higher
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