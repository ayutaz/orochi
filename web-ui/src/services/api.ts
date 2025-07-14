import axios from 'axios';
import { Torrent } from '../types/torrent';

const API_BASE = '/api';

export const api = {
  // Torrent operations
  getTorrents: async (): Promise<Torrent[]> => {
    const response = await axios.get(`${API_BASE}/torrents`);
    return response.data;
  },

  getTorrent: async (id: string): Promise<Torrent> => {
    const response = await axios.get(`${API_BASE}/torrents/${id}`);
    return response.data;
  },

  addTorrent: async (file: File): Promise<{ id: string }> => {
    const formData = new FormData();
    formData.append('torrent', file);
    const response = await axios.post(`${API_BASE}/torrents`, formData);
    return response.data;
  },

  addMagnet: async (magnetLink: string): Promise<{ id: string }> => {
    const response = await axios.post(`${API_BASE}/torrents/magnet`, { magnet: magnetLink });
    return response.data;
  },

  deleteTorrent: async (id: string): Promise<void> => {
    await axios.delete(`${API_BASE}/torrents/${id}`);
  },

  startTorrent: async (id: string): Promise<void> => {
    await axios.post(`${API_BASE}/torrents/${id}/start`);
  },

  stopTorrent: async (id: string): Promise<void> => {
    await axios.post(`${API_BASE}/torrents/${id}/stop`);
  },

  updateFiles: async (
    torrentId: string,
    files: Array<{ path: string; selected: boolean }>
  ): Promise<void> => {
    await axios.put(`${API_BASE}/torrents/${torrentId}/files`, { files });
  },

  // Settings operations
  getSettings: async (): Promise<any> => {
    const response = await axios.get(`${API_BASE}/settings`);
    return response.data;
  },

  updateSettings: async (settings: any): Promise<void> => {
    await axios.put(`${API_BASE}/settings`, settings);
  },
};
