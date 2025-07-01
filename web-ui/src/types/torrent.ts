export interface Torrent {
  id: string
  info: {
    name: string
    infoHash: string
    length: number
    pieceLength: number
    announce: string
    trackers: string[]
    files: FileInfo[]
  }
  status: 'stopped' | 'downloading' | 'seeding' | 'error'
  progress: number
  downloaded: number
  uploaded: number
  downloadRate: number
  uploadRate: number
  addedAt: string
  error?: string
}

export interface FileInfo {
  path: string[]
  length: number
  selected?: boolean
  priority?: 'low' | 'normal' | 'high'
}

export interface TorrentStats {
  downloadSpeed: number
  uploadSpeed: number
  eta: number
  peers: number
  seeds: number
  ratio: number
}