# Orochi Web UI

Modern React-based user interface for the Orochi torrent client.

## Features

- **Modern Design**: Built with Material-UI for a clean, responsive interface
- **Real-time Updates**: WebSocket integration for live torrent status updates
- **Multi-language Support**: Japanese and English translations
- **Dark Mode**: Toggle between light and dark themes
- **Torrent Management**: 
  - Add torrents via file upload or magnet links
  - Start/stop/remove torrents
  - View detailed torrent information
  - File selection and priority management
  - Peer and tracker information
- **Settings**: Configure download paths, network settings, and UI preferences

## Development

### Prerequisites

- Node.js 18+ and npm
- The Orochi server running on port 3030

### Setup

1. Install dependencies:
   ```bash
   npm install
   ```

2. Start the development server:
   ```bash
   npm start
   ```

3. The app will open at http://localhost:5173 and proxy API requests to the Go server at http://localhost:3030

### Building for Production

The UI is built to static files that are embedded in the Go binary:

```bash
npm run build
```

This creates optimized production files in `../internal/web/dist/` which are automatically embedded when building the Go application.

## Project Structure

```
web-ui/
├── src/
│   ├── components/      # Reusable UI components
│   ├── contexts/        # React contexts (Theme, WebSocket)
│   ├── hooks/           # Custom React hooks
│   ├── locales/         # Translation files (en.json, ja.json)
│   ├── pages/           # Page components
│   ├── services/        # API service layer
│   ├── types/           # TypeScript type definitions
│   └── utils/           # Utility functions
├── public/              # Static assets
├── index.html           # HTML template
├── package.json         # Dependencies and scripts
├── tsconfig.json        # TypeScript configuration
└── vite.config.ts       # Vite build configuration
```

## Technologies

- **React 18**: UI framework
- **TypeScript**: Type safety
- **Material-UI v5**: Component library
- **Vite**: Build tool
- **React Router v6**: Client-side routing
- **i18next**: Internationalization
- **Axios**: HTTP client
- **WebSocket**: Real-time updates

## API Integration

The UI communicates with the Go server through:

- REST API endpoints under `/api/*`
- WebSocket connection at `/ws` for real-time updates

During development, Vite proxies these requests to `http://localhost:3030`.

## Localization

The UI supports multiple languages:

- Japanese (default)
- English

Translations are stored in `src/locales/` and can be switched in the settings page.