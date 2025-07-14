import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Button,
  Checkbox,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  Select,
  MenuItem,
  FormControl,
} from '@mui/material';
import {
  Folder as FolderIcon,
  InsertDriveFile as FileIcon,
  Save as SaveIcon,
} from '@mui/icons-material';
import { Torrent, FileInfo } from '../types/torrent';
import { formatBytes } from '../utils/format';
import { api } from '../services/api';

interface TorrentFilesProps {
  torrent: Torrent;
}

const TorrentFiles: React.FC<TorrentFilesProps> = ({ torrent }) => {
  const { t } = useTranslation();
  const [files, setFiles] = useState<FileInfo[]>(torrent.info.files || []);
  const [hasChanges, setHasChanges] = useState(false);

  const handleSelectAll = (checked: boolean) => {
    setFiles(files.map((file) => ({ ...file, selected: checked })));
    setHasChanges(true);
  };

  const handleSelectFile = (index: number, checked: boolean) => {
    const newFiles = [...files];
    newFiles[index] = { ...newFiles[index], selected: checked };
    setFiles(newFiles);
    setHasChanges(true);
  };

  const handlePriorityChange = (index: number, priority: FileInfo['priority']) => {
    const newFiles = [...files];
    newFiles[index] = { ...newFiles[index], priority };
    setFiles(newFiles);
    setHasChanges(true);
  };

  const handleSave = async () => {
    try {
      await api.updateFiles(
        torrent.id,
        files.map((file) => ({
          path: file.path.join('/'),
          selected: file.selected !== false,
        }))
      );
      setHasChanges(false);
    } catch (error) {
      console.error('Failed to update files:', error);
    }
  };

  const allSelected = files.every((file) => file.selected !== false);
  const someSelected = files.some((file) => file.selected !== false) && !allSelected;

  if (!files || files.length === 0) {
    return (
      <Box sx={{ textAlign: 'center', py: 4 }}>
        <Typography variant="body1" color="text.secondary">
          Single file torrent: {torrent.info.name}
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
          {formatBytes(torrent.info.length)}
        </Typography>
      </Box>
    );
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Box>
          <Button size="small" onClick={() => handleSelectAll(true)} sx={{ mr: 1 }}>
            {t('torrentDetail.selectAll')}
          </Button>
          <Button size="small" onClick={() => handleSelectAll(false)}>
            {t('torrentDetail.deselectAll')}
          </Button>
        </Box>
        {hasChanges && (
          <Button variant="contained" size="small" startIcon={<SaveIcon />} onClick={handleSave}>
            Save Changes
          </Button>
        )}
      </Box>

      <TableContainer>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell padding="checkbox">
                <Checkbox
                  indeterminate={someSelected}
                  checked={allSelected}
                  onChange={(e) => handleSelectAll(e.target.checked)}
                />
              </TableCell>
              <TableCell>Name</TableCell>
              <TableCell align="right">Size</TableCell>
              <TableCell align="center">Priority</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {files.map((file, index) => (
              <TableRow key={index}>
                <TableCell padding="checkbox">
                  <Checkbox
                    checked={file.selected !== false}
                    onChange={(e) => handleSelectFile(index, e.target.checked)}
                  />
                </TableCell>
                <TableCell>
                  <Box sx={{ display: 'flex', alignItems: 'center' }}>
                    {file.path.length > 1 ? (
                      <FolderIcon sx={{ mr: 1 }} />
                    ) : (
                      <FileIcon sx={{ mr: 1 }} />
                    )}
                    <Typography variant="body2" noWrap>
                      {file.path.join('/')}
                    </Typography>
                  </Box>
                </TableCell>
                <TableCell align="right">
                  <Typography variant="body2">{formatBytes(file.length)}</Typography>
                </TableCell>
                <TableCell align="center">
                  <FormControl size="small" variant="standard">
                    <Select
                      value={file.priority || 'normal'}
                      onChange={(e) =>
                        handlePriorityChange(index, e.target.value as FileInfo['priority'])
                      }
                      disabled={file.selected === false}
                    >
                      <MenuItem value="low">{t('torrentDetail.priority.low')}</MenuItem>
                      <MenuItem value="normal">{t('torrentDetail.priority.normal')}</MenuItem>
                      <MenuItem value="high">{t('torrentDetail.priority.high')}</MenuItem>
                    </Select>
                  </FormControl>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
};

export default TorrentFiles;
