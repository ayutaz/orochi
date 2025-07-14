import React, { useState, useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Tabs,
  Tab,
  Box,
  Alert,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  IconButton,
  LinearProgress,
  Typography,
  Paper,
} from '@mui/material';
import {
  CloudUpload as CloudUploadIcon,
  Close as CloseIcon,
  InsertDriveFile as FileIcon,
} from '@mui/icons-material';
import { api } from '../services/api';

interface AddTorrentDialogProps {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
  initialFiles?: File[];
}

const AddTorrentDialog: React.FC<AddTorrentDialogProps> = ({
  open,
  onClose,
  onSuccess,
  initialFiles = [],
}) => {
  const { t } = useTranslation();
  const [tab, setTab] = useState(0);
  const [files, setFiles] = useState<File[]>([]);
  const [magnetLink, setMagnetLink] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [uploadProgress, setUploadProgress] = useState<{ [key: string]: number }>({});
  const [isDragging, setIsDragging] = useState(false);

  // initialFilesが渡された場合、ファイルリストを初期化
  useEffect(() => {
    if (open && initialFiles.length > 0) {
      setFiles(initialFiles);
      setTab(0); // ファイルタブに切り替え
    }
  }, [open, initialFiles]);

  const handleClose = () => {
    setFiles([]);
    setMagnetLink('');
    setError('');
    setUploadProgress({});
    setTab(0);
    onClose();
  };

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files) {
      const newFiles = Array.from(event.target.files).filter((file) =>
        file.name.endsWith('.torrent')
      );
      setFiles((prev) => [...prev, ...newFiles]);
      setError('');
    }
  };

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    const droppedFiles = Array.from(e.dataTransfer.files).filter((file) =>
      file.name.endsWith('.torrent')
    );

    if (droppedFiles.length === 0) {
      setError('Only .torrent files are allowed');
      return;
    }

    setFiles((prev) => [...prev, ...droppedFiles]);
    setError('');
  }, []);

  const removeFile = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index));
  };

  const handleSubmit = async () => {
    setLoading(true);
    setError('');

    try {
      if (tab === 0) {
        // File upload
        if (files.length === 0) {
          setError('Please select at least one file');
          setLoading(false);
          return;
        }

        // Upload files sequentially to show progress
        for (let i = 0; i < files.length; i++) {
          const file = files[i];
          try {
            setUploadProgress((prev) => ({ ...prev, [file.name]: 0 }));

            // Simulate progress (in real app, use XMLHttpRequest for progress tracking)
            await api.addTorrent(file);

            setUploadProgress((prev) => ({ ...prev, [file.name]: 100 }));
          } catch (error) {
            console.error(`Failed to upload ${file.name}:`, error);
            setUploadProgress((prev) => ({ ...prev, [file.name]: -1 }));
          }
        }
      } else {
        // Magnet link
        if (!magnetLink.trim()) {
          setError('Please enter a magnet link');
          setLoading(false);
          return;
        }
        await api.addMagnet(magnetLink.trim());
      }

      // Wait a bit to show completion
      setTimeout(() => {
        onSuccess();
        handleClose();
      }, 500);
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to add torrent');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('torrent.addTorrent')}</DialogTitle>
      <DialogContent>
        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <Tabs value={tab} onChange={(_, newValue) => setTab(newValue)}>
            <Tab label={t('torrent.selectFile')} />
            <Tab label={t('torrent.magnetLink')} />
          </Tabs>
        </Box>
        <Box sx={{ pt: 3 }}>
          {tab === 0 ? (
            <Box>
              <Paper
                sx={{
                  p: 3,
                  border: 2,
                  borderStyle: 'dashed',
                  borderColor: isDragging ? 'primary.main' : 'divider',
                  bgcolor: isDragging ? 'action.hover' : 'background.paper',
                  cursor: 'pointer',
                  transition: 'all 0.2s',
                  textAlign: 'center',
                }}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
                onClick={() => document.getElementById('torrent-file-input')?.click()}
              >
                <input
                  accept=".torrent"
                  style={{ display: 'none' }}
                  id="torrent-file-input"
                  type="file"
                  multiple
                  onChange={handleFileChange}
                />
                <CloudUploadIcon sx={{ fontSize: 48, mb: 2, opacity: 0.6 }} />
                <Typography variant="h6" gutterBottom>
                  {t('torrent.dropFileHere')}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  または クリックしてファイルを選択
                </Typography>
                <Typography variant="caption" display="block" sx={{ mt: 1 }}>
                  複数のファイルを選択できます
                </Typography>
              </Paper>

              {files.length > 0 && (
                <Box sx={{ mt: 2 }}>
                  <Typography variant="subtitle2" gutterBottom>
                    選択されたファイル ({files.length})
                  </Typography>
                  <List dense>
                    {files.map((file, index) => (
                      <ListItem key={`${file.name}-${index}`}>
                        <FileIcon sx={{ mr: 1, opacity: 0.6 }} />
                        <ListItemText
                          primary={file.name}
                          secondary={`${(file.size / 1024).toFixed(1)} KB`}
                        />
                        {uploadProgress[file.name] !== undefined && (
                          <Box sx={{ width: 100, mr: 2 }}>
                            {uploadProgress[file.name] === -1 ? (
                              <Typography variant="caption" color="error">
                                Failed
                              </Typography>
                            ) : (
                              <LinearProgress
                                variant="determinate"
                                value={uploadProgress[file.name]}
                                color={uploadProgress[file.name] === 100 ? 'success' : 'primary'}
                              />
                            )}
                          </Box>
                        )}
                        <ListItemSecondaryAction>
                          <IconButton
                            edge="end"
                            size="small"
                            onClick={() => removeFile(index)}
                            disabled={loading}
                          >
                            <CloseIcon />
                          </IconButton>
                        </ListItemSecondaryAction>
                      </ListItem>
                    ))}
                  </List>
                </Box>
              )}
            </Box>
          ) : (
            <TextField
              fullWidth
              multiline
              rows={4}
              variant="outlined"
              placeholder={t('torrent.pastemagnetLink')}
              value={magnetLink}
              onChange={(e) => setMagnetLink(e.target.value)}
            />
          )}
          {error && (
            <Alert severity="error" sx={{ mt: 2 }}>
              {error}
            </Alert>
          )}
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>{t('common.cancel')}</Button>
        <Button
          onClick={handleSubmit}
          variant="contained"
          disabled={loading || (tab === 0 ? files.length === 0 : !magnetLink.trim())}
        >
          {loading ? (
            <>
              {tab === 0 && files.length > 1
                ? `Uploading (${Object.values(uploadProgress).filter((p) => p === 100).length}/${files.length})...`
                : 'Uploading...'}
            </>
          ) : (
            t('common.add')
          )}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default AddTorrentDialog;
