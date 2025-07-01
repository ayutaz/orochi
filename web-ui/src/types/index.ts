export interface TorrentInfo {
  name: string
  length: number
  pieceLength: number
  pieces: string
  files?: FileInfo[]
}

export interface FileInfo {
  path: string[]
  length: number
  selected?: boolean
  priority?: 'low' | 'normal' | 'high'
}

export interface Torrent {
  id: string
  info: TorrentInfo
  status: 'downloading' | 'seeding' | 'stopped' | 'error' | 'checking'
  progress: number
  downloaded: number
  uploaded: number
  downloadSpeed: number
  uploadSpeed: number
  peers: number
  seeds: number
  ratio: number
  eta?: number
  error?: string
  addedAt: string
  completedAt?: string
}

export interface Peer {
  id: string
  address: string
  client: string
  progress: number
  downloadSpeed: number
  uploadSpeed: number
}

export interface Tracker {
  url: string
  status: 'working' | 'error' | 'disabled'
  peers: number
  lastAnnounce?: string
  nextAnnounce?: string
  error?: string
}

export interface Settings {
  language: 'en' | 'ja'
  theme: 'light' | 'dark'
  downloadPath: string
  maxConnections: number
  port: number
  uploadLimit?: number
  downloadLimit?: number
}

export interface WebSocketMessage {
  type: 'torrent_update' | 'torrent_added' | 'torrent_removed' | 'error'
  data: any
}