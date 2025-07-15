import { lazy, Suspense } from 'react';
import { Routes, Route } from 'react-router-dom';
import { CircularProgress, Box } from '@mui/material';
import { CustomThemeProvider } from './contexts/ThemeContext';
import { WebSocketProvider } from './contexts/WebSocketContext';
import Layout from './components/Layout';

// Lazy load pages
const TorrentList = lazy(() => import('./pages/TorrentList'));
const TorrentDetail = lazy(() => import('./pages/TorrentDetail'));
const Settings = lazy(() => import('./pages/Settings'));

// Loading component
const PageLoader = () => (
  <Box display="flex" justifyContent="center" alignItems="center" minHeight="60vh">
    <CircularProgress />
  </Box>
);

function App() {
  return (
    <CustomThemeProvider>
      <WebSocketProvider>
        <Layout>
          <Suspense fallback={<PageLoader />}>
            <Routes>
              <Route path="/" element={<TorrentList />} />
              <Route path="/torrent/:id" element={<TorrentDetail />} />
              <Route path="/settings" element={<Settings />} />
            </Routes>
          </Suspense>
        </Layout>
      </WebSocketProvider>
    </CustomThemeProvider>
  );
}

export default App;
