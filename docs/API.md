# Orochi API Documentation

## Overview

Orochi provides a RESTful API for managing BitTorrent downloads. All API endpoints are prefixed with `/api`.

## Base URL

```
http://localhost:8080
```

## Authentication

Currently, the API does not require authentication. This may change in future versions.

## Endpoints

### Health Check

Check if the server is running.

```
GET /health
```

**Response**
- Status: 200 OK
- Body: `OK`

### List Torrents

Get a list of all torrents.

```
GET /api/torrents
```

**Response**
- Status: 200 OK
- Body: Array of torrent objects

```json
[
  {
    "id": "1234567890abcdef1234567890abcdef12345678",
    "info": {
      "name": "Example.txt",
      "info_hash": "1234567890abcdef1234567890abcdef12345678",
      "length": 1024,
      "piece_length": 16384,
      "announce": "http://tracker.example.com:8080/announce",
      "trackers": ["http://tracker.example.com:8080/announce"],
      "files": []
    },
    "status": "stopped",
    "progress": 0,
    "downloaded": 0,
    "uploaded": 0,
    "added_at": "2024-01-31T12:00:00Z",
    "error": ""
  }
]
```

### Add Torrent File

Add a new torrent from a .torrent file.

```
POST /api/torrents
```

**Request**
- Content-Type: multipart/form-data
- Form field: `torrent` - The torrent file

**Response**
- Status: 201 Created
- Body:
```json
{
  "id": "1234567890abcdef1234567890abcdef12345678"
}
```

**Error Responses**
- 400 Bad Request - Invalid torrent file
- 500 Internal Server Error - Server error

### Add Magnet Link

Add a new torrent from a magnet link.

```
POST /api/torrents/magnet
```

**Request**
- Content-Type: application/json
- Body:
```json
{
  "magnet": "magnet:?xt=urn:btih:1234567890abcdef1234567890abcdef12345678&dn=Example.txt"
}
```

**Response**
- Status: 201 Created
- Body:
```json
{
  "id": "1234567890abcdef1234567890abcdef12345678"
}
```

**Error Responses**
- 400 Bad Request - Invalid magnet link or missing magnet field
- 500 Internal Server Error - Server error

### Get Torrent

Get details of a specific torrent.

```
GET /api/torrents/:id
```

**Parameters**
- `id` - The torrent ID (info hash)

**Response**
- Status: 200 OK
- Body: Torrent object (same as in List Torrents)

**Error Responses**
- 404 Not Found - Torrent not found

### Delete Torrent

Remove a torrent.

```
DELETE /api/torrents/:id
```

**Parameters**
- `id` - The torrent ID (info hash)

**Response**
- Status: 204 No Content

**Error Responses**
- 404 Not Found - Torrent not found
- 500 Internal Server Error - Server error

### Start Torrent

Start downloading a torrent.

```
POST /api/torrents/:id/start
```

**Parameters**
- `id` - The torrent ID (info hash)

**Response**
- Status: 200 OK
- Body:
```json
{
  "status": "started"
}
```

**Error Responses**
- 404 Not Found - Torrent not found
- 500 Internal Server Error - Server error

### Stop Torrent

Stop downloading a torrent.

```
POST /api/torrents/:id/stop
```

**Parameters**
- `id` - The torrent ID (info hash)

**Response**
- Status: 200 OK
- Body:
```json
{
  "status": "stopped"
}
```

**Error Responses**
- 404 Not Found - Torrent not found
- 500 Internal Server Error - Server error

## Error Format

All error responses follow this format:

```json
{
  "error": "Error message"
}
```

## Status Codes

- 200 OK - Request successful
- 201 Created - Resource created successfully
- 204 No Content - Request successful, no content to return
- 400 Bad Request - Invalid request data
- 404 Not Found - Resource not found
- 500 Internal Server Error - Server error

## Rate Limiting

Currently, there is no rate limiting. This may be added in future versions.

## Versioning

The API is currently at version 1.0. Future versions may introduce breaking changes.