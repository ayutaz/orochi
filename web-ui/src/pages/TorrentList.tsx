import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Fab,
  Grid,
  IconButton,
  LinearProgress,
  Typography,
  Menu,
  MenuItem,
  Backdrop,
  Paper,
} from '@mui/material';
import {
  Add as AddIcon,
  PlayArrow as PlayArrowIcon,
  Stop as StopIcon,
  Delete as DeleteIcon,
  MoreVert as MoreVertIcon,
  CloudDownload as CloudDownloadIcon,
  CloudUpload as CloudUploadIcon,
} from '@mui/icons-material';
import { api } from '../services/api';
import { Torrent } from '../types/torrent';
import { useWebSocket } from '../contexts/WebSocketContext';
import AddTorrentDialog from '../components/AddTorrentDialog';
import { formatBytes, formatSpeed } from '../utils/format';

const TorrentList: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { lastMessage } = useWebSocket();
  const [torrents, setTorrents] = useState<Torrent[]>([]);
  const [loading, setLoading] = useState(true);
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedTorrent, setSelectedTorrent] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [droppedFiles, setDroppedFiles] = useState<File[]>([]);

  useEffect(() => {
    loadTorrents();
  }, []);

  useEffect(() => {
    if (lastMessage) {
      if (lastMessage.type === 'torrents' && lastMessage.data) {
        // 直接トレントデータを更新
        setTorrents(lastMessage.data);
      } else if (lastMessage.type === 'torrent_update') {
        // 従来の更新通知の場合はAPIから取得
        loadTorrents();
      }
    }
  }, [lastMessage]);

  const loadTorrents = async () => {
    try {
      const data = await api.getTorrents();
      setTorrents(data);
    } catch (error) {
      console.error('Failed to load torrents:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleStartTorrent = async (id: string) => {
    try {
      await api.startTorrent(id);
      await loadTorrents();
    } catch (error) {
      console.error('Failed to start torrent:', error);
    }
  };

  const handleStopTorrent = async (id: string) => {
    try {
      await api.stopTorrent(id);
      await loadTorrents();
    } catch (error) {
      console.error('Failed to stop torrent:', error);
    }
  };

  const handleDeleteTorrent = async (id: string) => {
    if (window.confirm(t('messages.confirmDelete'))) {
      try {
        await api.deleteTorrent(id);
        await loadTorrents();
      } catch (error) {
        console.error('Failed to delete torrent:', error);
      }
    }
    handleCloseMenu();
  };

  const handleMenuClick = (event: React.MouseEvent<HTMLElement>, torrentId: string) => {
    setAnchorEl(event.currentTarget);
    setSelectedTorrent(torrentId);
  };

  const handleCloseMenu = () => {
    setAnchorEl(null);
    setSelectedTorrent(null);
  };

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();

    // Check if files are torrent files
    const items = Array.from(e.dataTransfer.items);
    const hasTorrentFiles = items.some((item) => {
      if (item.kind === 'file') {
        const file = item.getAsFile();
        return file && file.name.endsWith('.torrent');
      }
      return false;
    });

    if (hasTorrentFiles) {
      setIsDragging(true);
    }
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();

    // Only set dragging to false if we're leaving the main container
    const rect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX;
    const y = e.clientY;

    if (x <= rect.left || x >= rect.right || y <= rect.top || y >= rect.bottom) {
      setIsDragging(false);
    }
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    const files = Array.from(e.dataTransfer.files).filter((file) => file.name.endsWith('.torrent'));

    if (files.length > 0) {
      // AddTorrentDialogに渡すためにファイルを保存
      setDroppedFiles(files);
      setAddDialogOpen(true);
    }
  }, []);

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

  const handleAddDialogClose = useCallback(() => {
    setAddDialogOpen(false);
    // ダイアログが閉じたらドロップされたファイルをクリア
    setDroppedFiles([]);
  }, []);

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="80vh">
        <CircularProgress />
      </Box>
    );
  }

  if (torrents.length === 0) {
    return (
      <Box
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        sx={{ minHeight: '100vh', position: 'relative' }}
      >
        <Box
          display="flex"
          flexDirection="column"
          alignItems="center"
          justifyContent="center"
          height="60vh"
        >
          <CloudDownloadIcon sx={{ fontSize: 80, mb: 2, opacity: 0.3 }} />
          <Typography variant="h5" gutterBottom>
            {t('torrent.noTorrents')}
          </Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setAddDialogOpen(true)}
            sx={{ mt: 2 }}
          >
            {t('torrent.addTorrent')}
          </Button>
        </Box>
        <AddTorrentDialog
          open={addDialogOpen}
          onClose={handleAddDialogClose}
          onSuccess={loadTorrents}
          initialFiles={droppedFiles}
        />

        {/* Drag & Drop Overlay */}
        <Backdrop
          open={isDragging}
          sx={{
            position: 'absolute',
            zIndex: 9999,
            backgroundColor: 'rgba(0, 0, 0, 0.8)',
          }}
        >
          <Paper
            sx={{
              p: 6,
              textAlign: 'center',
              backgroundColor: 'background.paper',
              border: 3,
              borderStyle: 'dashed',
              borderColor: 'primary.main',
            }}
          >
            <CloudUploadIcon sx={{ fontSize: 80, mb: 2, color: 'primary.main' }} />
            <Typography variant="h4" gutterBottom>
              {t('torrent.dropFileHere')}
            </Typography>
            <Typography variant="body1" color="text.secondary">
              複数のトレントファイルをドロップできます
            </Typography>
          </Paper>
        </Backdrop>
      </Box>
    );
  }

  return (
    <Box
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      sx={{ minHeight: '100vh', position: 'relative' }}
    >
      <Grid container spacing={2}>
        {torrents.map((torrent) => (
          <Grid item xs={12} key={torrent.id}>
            <Card>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box
                    flex={1}
                    onClick={() => navigate(`/torrent/${torrent.id}`)}
                    sx={{ cursor: 'pointer' }}
                  >
                    <Typography variant="h6" gutterBottom>
                      {torrent.info.name}
                    </Typography>
                    <Box display="flex" alignItems="center" gap={2} mb={1}>
                      <Chip
                        label={t(`status.${torrent.status}`)}
                        color={getStatusColor(torrent.status)}
                        size="small"
                      />
                      <Typography variant="body2" color="text.secondary">
                        {formatBytes(torrent.downloaded)} / {formatBytes(torrent.info.length)}
                      </Typography>
                      {(torrent.status === 'downloading' || torrent.status === 'seeding') && (
                        <>
                          <Box display="flex" alignItems="center" gap={0.5}>
                            <CloudDownloadIcon fontSize="small" />
                            <Typography variant="body2">
                              {formatSpeed(torrent.downloadRate || 0)}
                            </Typography>
                          </Box>
                          <Box display="flex" alignItems="center" gap={0.5}>
                            <CloudUploadIcon fontSize="small" />
                            <Typography variant="body2">
                              {formatSpeed(torrent.uploadRate || 0)}
                            </Typography>
                          </Box>
                        </>
                      )}
                    </Box>
                    <LinearProgress
                      variant="determinate"
                      value={torrent.progress}
                      sx={{ height: 6, borderRadius: 3 }}
                    />
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                      {torrent.progress.toFixed(1)}%
                    </Typography>
                  </Box>
                  <Box display="flex" alignItems="center">
                    {torrent.status === 'stopped' ? (
                      <IconButton color="primary" onClick={() => handleStartTorrent(torrent.id)}>
                        <PlayArrowIcon />
                      </IconButton>
                    ) : (
                      <IconButton color="default" onClick={() => handleStopTorrent(torrent.id)}>
                        <StopIcon />
                      </IconButton>
                    )}
                    <IconButton onClick={(e) => handleMenuClick(e, torrent.id)}>
                      <MoreVertIcon />
                    </IconButton>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleCloseMenu}>
        <MenuItem onClick={() => selectedTorrent && navigate(`/torrent/${selectedTorrent}`)}>
          {t('common.details')}
        </MenuItem>
        <MenuItem onClick={() => selectedTorrent && handleDeleteTorrent(selectedTorrent)}>
          <DeleteIcon fontSize="small" sx={{ mr: 1 }} />
          {t('common.delete')}
        </MenuItem>
      </Menu>

      <Fab
        color="primary"
        aria-label="add"
        sx={{ position: 'fixed', bottom: 16, right: 16 }}
        onClick={() => setAddDialogOpen(true)}
      >
        <AddIcon />
      </Fab>

      <AddTorrentDialog
        open={addDialogOpen}
        onClose={handleAddDialogClose}
        onSuccess={loadTorrents}
        initialFiles={droppedFiles}
      />

      {/* Drag & Drop Overlay */}
      <Backdrop
        open={isDragging}
        sx={{
          position: 'absolute',
          zIndex: 9999,
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
        }}
      >
        <Paper
          sx={{
            p: 6,
            textAlign: 'center',
            backgroundColor: 'background.paper',
            border: 3,
            borderStyle: 'dashed',
            borderColor: 'primary.main',
          }}
        >
          <CloudUploadIcon sx={{ fontSize: 80, mb: 2, color: 'primary.main' }} />
          <Typography variant="h4" gutterBottom>
            {t('torrent.dropFileHere')}
          </Typography>
          <Typography variant="body1" color="text.secondary">
            複数のトレントファイルをドロップできます
          </Typography>
        </Paper>
      </Backdrop>
    </Box>
  );
};

export default TorrentList;
