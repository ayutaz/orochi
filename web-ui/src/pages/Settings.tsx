import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Paper,
  Typography,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Switch,
  FormControlLabel,
  Button,
  Divider,
  Alert,
} from '@mui/material';
import { Save as SaveIcon } from '@mui/icons-material';
import i18n from '../i18n';
import { api } from '../services/api';
import { useTheme } from '../contexts/ThemeContext';

const Settings: React.FC = () => {
  const { t } = useTranslation();
  const { darkMode, toggleTheme } = useTheme();
  const [settings, setSettings] = useState({
    language: i18n.language,
    downloadPath: '',
    maxDownloadSpeed: 0,
    maxUploadSpeed: 0,
    maxConnections: 200,
    port: 6881,
    dht: true,
    peerExchange: true,
    localPeerDiscovery: true,
  });
  const [saved, setSaved] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      setLoading(true);
      const data = await api.getSettings();
      setSettings({
        language: data.language || i18n.language,
        downloadPath: data.downloadPath || '',
        maxDownloadSpeed: data.maxDownloadSpeed || 0,
        maxUploadSpeed: data.maxUploadSpeed || 0,
        maxConnections: data.maxConnections || 200,
        port: data.port || 6881,
        dht: data.dht !== false,
        peerExchange: data.peerExchange !== false,
        localPeerDiscovery: data.localPeerDiscovery !== false,
      });
    } catch (err) {
      setError('Failed to load settings');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      setError(null);
      await api.updateSettings(settings);

      // Save language preference
      localStorage.setItem('language', settings.language);
      i18n.changeLanguage(settings.language);

      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } catch (err) {
      setError('Failed to save settings');
      console.error(err);
    }
  };

  const handleChange = (field: string, value: any) => {
    setSettings((prev) => ({ ...prev, [field]: value }));
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="60vh">
        <Typography>Loading...</Typography>
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h5" gutterBottom>
        {t('navigation.settings')}
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          {t('settings.general')}
        </Typography>
        <FormControl fullWidth sx={{ mb: 2 }}>
          <InputLabel>{t('settings.language')}</InputLabel>
          <Select
            value={settings.language}
            label={t('settings.language')}
            onChange={(e) => handleChange('language', e.target.value)}
          >
            <MenuItem value="ja">日本語</MenuItem>
            <MenuItem value="en">English</MenuItem>
          </Select>
        </FormControl>
        <FormControlLabel
          control={<Switch checked={darkMode} onChange={toggleTheme} />}
          label={t('settings.darkMode')}
        />
      </Paper>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          {t('settings.downloads')}
        </Typography>
        <TextField
          fullWidth
          label={t('settings.downloadPath')}
          value={settings.downloadPath}
          onChange={(e) => handleChange('downloadPath', e.target.value)}
          sx={{ mb: 2 }}
        />
        <TextField
          fullWidth
          type="number"
          label={t('settings.maxDownloadSpeed')}
          value={settings.maxDownloadSpeed}
          onChange={(e) => handleChange('maxDownloadSpeed', parseInt(e.target.value))}
          helperText="0 = unlimited (KB/s)"
          sx={{ mb: 2 }}
        />
        <TextField
          fullWidth
          type="number"
          label={t('settings.maxUploadSpeed')}
          value={settings.maxUploadSpeed}
          onChange={(e) => handleChange('maxUploadSpeed', parseInt(e.target.value))}
          helperText="0 = unlimited (KB/s)"
        />
      </Paper>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          {t('settings.network')}
        </Typography>
        <TextField
          fullWidth
          type="number"
          label={t('settings.port')}
          value={settings.port}
          onChange={(e) => handleChange('port', parseInt(e.target.value))}
          sx={{ mb: 2 }}
        />
        <TextField
          fullWidth
          type="number"
          label={t('settings.maxConnections')}
          value={settings.maxConnections}
          onChange={(e) => handleChange('maxConnections', parseInt(e.target.value))}
          sx={{ mb: 2 }}
        />
        <Divider sx={{ my: 2 }} />
        <FormControlLabel
          control={
            <Switch
              checked={settings.dht}
              onChange={(e) => handleChange('dht', e.target.checked)}
            />
          }
          label={t('settings.dht')}
        />
        <FormControlLabel
          control={
            <Switch
              checked={settings.peerExchange}
              onChange={(e) => handleChange('peerExchange', e.target.checked)}
            />
          }
          label={t('settings.peerExchange')}
        />
        <FormControlLabel
          control={
            <Switch
              checked={settings.localPeerDiscovery}
              onChange={(e) => handleChange('localPeerDiscovery', e.target.checked)}
            />
          }
          label={t('settings.localPeerDiscovery')}
        />
      </Paper>

      {saved && (
        <Alert severity="success" sx={{ mb: 3 }}>
          {t('messages.settingsSaved')}
        </Alert>
      )}

      <Button variant="contained" startIcon={<SaveIcon />} onClick={handleSave} size="large">
        {t('common.save')}
      </Button>
    </Box>
  );
};

export default Settings;
