import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Paper,
  Typography,
  IconButton,
  Tabs,
  Tab,
  CircularProgress,
  Alert,
  Grid,
  Chip,
  LinearProgress,
  Button,
} from '@mui/material';
import {
  ArrowBack as ArrowBackIcon,
  PlayArrow as PlayArrowIcon,
  Stop as StopIcon,
  Delete as DeleteIcon,
  CloudDownload as CloudDownloadIcon,
  CloudUpload as CloudUploadIcon,
} from '@mui/icons-material';
import { api } from '../services/api';
import { Torrent } from '../types/torrent';
import { formatBytes, formatSpeed } from '../utils/format';
import TorrentFiles from '../components/TorrentFiles';
import TorrentPeers from '../components/TorrentPeers';
import TorrentTrackers from '../components/TorrentTrackers';
import { useWebSocket } from '../contexts/WebSocketContext';

const TorrentDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { lastMessage } = useWebSocket();
  const [tab, setTab] = useState(0);
  const [torrent, setTorrent] = useState<Torrent | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadTorrent();
  }, [id]);

  useEffect(() => {
    if (lastMessage?.type === 'torrents' && lastMessage.data) {
      const updatedTorrent = lastMessage.data.find((t: Torrent) => t.id === id);
      if (updatedTorrent) {
        setTorrent(updatedTorrent);
      }
    }
  }, [lastMessage, id]);

  const loadTorrent = async () => {
    if (!id) return;

    try {
      setLoading(true);
      const data = await api.getTorrent(id);
      setTorrent(data);
    } catch (err) {
      setError('Failed to load torrent details');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleStart = async () => {
    if (!torrent) return;
    try {
      await api.startTorrent(torrent.id);
      await loadTorrent();
    } catch (error) {
      console.error('Failed to start torrent:', error);
    }
  };

  const handleStop = async () => {
    if (!torrent) return;
    try {
      await api.stopTorrent(torrent.id);
      await loadTorrent();
    } catch (error) {
      console.error('Failed to stop torrent:', error);
    }
  };

  const handleDelete = async () => {
    if (!torrent || !window.confirm(t('messages.confirmDelete'))) return;
    try {
      await api.deleteTorrent(torrent.id);
      navigate('/');
    } catch (error) {
      console.error('Failed to delete torrent:', error);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'downloading':
        return 'primary';
      case 'seeding':
        return 'success';
      case 'stopped':
        return 'default';
      case 'error':
        return 'error';
      default:
        return 'default';
    }
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="60vh">
        <CircularProgress />
      </Box>
    );
  }

  if (error || !torrent) {
    return (
      <Box>
        <Box display="flex" alignItems="center" mb={2}>
          <IconButton onClick={() => navigate('/')} sx={{ mr: 2 }}>
            <ArrowBackIcon />
          </IconButton>
          <Typography variant="h5">{t('torrentDetail.title')}</Typography>
        </Box>
        <Alert severity="error">{error || 'Torrent not found'}</Alert>
      </Box>
    );
  }

  return (
    <Box>
      <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
        <Box display="flex" alignItems="center">
          <IconButton onClick={() => navigate('/')} sx={{ mr: 2 }}>
            <ArrowBackIcon />
          </IconButton>
          <Typography variant="h5">{torrent.info.name}</Typography>
        </Box>
        <Box>
          {torrent.status === 'stopped' ? (
            <Button
              variant="contained"
              startIcon={<PlayArrowIcon />}
              onClick={handleStart}
              sx={{ mr: 1 }}
            >
              {t('common.start')}
            </Button>
          ) : (
            <Button variant="outlined" startIcon={<StopIcon />} onClick={handleStop} sx={{ mr: 1 }}>
              {t('common.stop')}
            </Button>
          )}
          <Button
            variant="outlined"
            color="error"
            startIcon={<DeleteIcon />}
            onClick={handleDelete}
          >
            {t('common.delete')}
          </Button>
        </Box>
      </Box>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Grid container spacing={2}>
          <Grid item xs={12} md={6}>
            <Typography variant="subtitle2" color="text.secondary">
              {t('torrentDetail.status')}
            </Typography>
            <Box display="flex" alignItems="center" gap={1} mb={2}>
              <Chip
                label={t(`status.${torrent.status}`)}
                color={getStatusColor(torrent.status)}
                size="small"
              />
              {torrent.error && (
                <Typography variant="body2" color="error">
                  {torrent.error}
                </Typography>
              )}
            </Box>
          </Grid>

          <Grid item xs={12} md={6}>
            <Typography variant="subtitle2" color="text.secondary">
              {t('torrentDetail.size')}
            </Typography>
            <Typography variant="body1" mb={2}>
              {formatBytes(torrent.info.length)}
            </Typography>
          </Grid>

          <Grid item xs={12}>
            <Typography variant="subtitle2" color="text.secondary">
              {t('torrentDetail.progress')}
            </Typography>
            <Box mb={2}>
              <LinearProgress
                variant="determinate"
                value={torrent.progress}
                sx={{ height: 8, borderRadius: 4, mb: 1 }}
              />
              <Box display="flex" justifyContent="space-between">
                <Typography variant="body2">
                  {formatBytes(torrent.downloaded)} / {formatBytes(torrent.info.length)}(
                  {torrent.progress.toFixed(1)}%)
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {t('torrentDetail.uploaded')}: {formatBytes(torrent.uploaded)}
                </Typography>
              </Box>
            </Box>
          </Grid>

          {(torrent.status === 'downloading' || torrent.status === 'seeding') && (
            <>
              <Grid item xs={6} md={3}>
                <Box display="flex" alignItems="center" gap={0.5}>
                  <CloudDownloadIcon fontSize="small" />
                  <Typography variant="subtitle2" color="text.secondary">
                    {t('torrentDetail.downloadSpeed')}
                  </Typography>
                </Box>
                <Typography variant="h6">{formatSpeed(torrent.downloadRate || 0)}</Typography>
              </Grid>

              <Grid item xs={6} md={3}>
                <Box display="flex" alignItems="center" gap={0.5}>
                  <CloudUploadIcon fontSize="small" />
                  <Typography variant="subtitle2" color="text.secondary">
                    {t('torrentDetail.uploadSpeed')}
                  </Typography>
                </Box>
                <Typography variant="h6">{formatSpeed(torrent.uploadRate || 0)}</Typography>
              </Grid>
            </>
          )}

          <Grid item xs={12}>
            <Typography variant="subtitle2" color="text.secondary">
              {t('torrentDetail.infoHash')}
            </Typography>
            <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
              {torrent.info.infoHash}
            </Typography>
          </Grid>
        </Grid>
      </Paper>

      <Paper sx={{ p: 3 }}>
        <Tabs
          value={tab}
          onChange={(_, newValue) => setTab(newValue)}
          sx={{ borderBottom: 1, borderColor: 'divider' }}
        >
          <Tab label={t('torrentDetail.tabs.files')} />
          <Tab label={t('torrentDetail.tabs.peers')} />
          <Tab label={t('torrentDetail.tabs.trackers')} />
        </Tabs>

        <Box sx={{ mt: 3 }}>
          {tab === 0 && <TorrentFiles torrent={torrent} />}
          {tab === 1 && <TorrentPeers torrentId={torrent.id} />}
          {tab === 2 && <TorrentTrackers torrentId={torrent.id} />}
        </Box>
      </Paper>
    </Box>
  );
};

export default TorrentDetail;
