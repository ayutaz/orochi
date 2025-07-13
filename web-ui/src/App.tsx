import { Routes, Route } from 'react-router-dom';
import { CustomThemeProvider } from './contexts/ThemeContext';
import { WebSocketProvider } from './contexts/WebSocketContext';
import Layout from './components/Layout';
import TorrentList from './pages/TorrentList';
import TorrentDetail from './pages/TorrentDetail';
import Settings from './pages/Settings';

function App() {
  return (
    <CustomThemeProvider>
      <WebSocketProvider>
        <Layout>
          <Routes>
            <Route path="/" element={<TorrentList />} />
            <Route path="/torrent/:id" element={<TorrentDetail />} />
            <Route path="/settings" element={<Settings />} />
          </Routes>
        </Layout>
      </WebSocketProvider>
    </CustomThemeProvider>
  );
}

export default App;
