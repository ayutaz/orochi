import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Alert,
} from '@mui/material';
import { api } from '../services/api';

interface AddMagnetDialogProps {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

const AddMagnetDialog: React.FC<AddMagnetDialogProps> = ({ open, onClose, onSuccess }) => {
  const { t } = useTranslation();
  const [magnet, setMagnet] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async () => {
    if (!magnet.trim()) {
      setError('Please enter a magnet link');
      return;
    }

    if (!magnet.startsWith('magnet:?')) {
      setError('Invalid magnet link format');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      await api.addMagnet(magnet);
      onSuccess();
      handleClose();
    } catch (err: any) {
      setError(err.response?.data?.error || t('errors.addFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setMagnet('');
    setError(null);
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('addMagnet.title')}</DialogTitle>
      <DialogContent>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}
        <TextField
          autoFocus
          margin="dense"
          label={t('addMagnet.title')}
          type="text"
          fullWidth
          multiline
          rows={3}
          variant="outlined"
          placeholder={t('addMagnet.placeholder')}
          value={magnet}
          onChange={(e) => setMagnet(e.target.value)}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={loading}>
          {t('addMagnet.cancel')}
        </Button>
        <Button onClick={handleSubmit} variant="contained" disabled={loading || !magnet.trim()}>
          {t('addMagnet.add')}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default AddMagnetDialog;
