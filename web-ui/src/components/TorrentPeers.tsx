import React from 'react';
import {
  Box,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  LinearProgress,
} from '@mui/material';
import { formatSpeed } from '../utils/format';

// Mock data for now - this would come from the API
const mockPeers = [
  {
    id: '1',
    address: '192.168.1.100:51234',
    client: 'qBittorrent/4.5.0',
    progress: 45.2,
    downloadSpeed: 125000,
    uploadSpeed: 45000,
  },
  {
    id: '2',
    address: '10.0.0.50:6881',
    client: 'Transmission/3.00',
    progress: 78.9,
    downloadSpeed: 0,
    uploadSpeed: 250000,
  },
];

interface TorrentPeersProps {
  torrentId: string;
}

const TorrentPeers: React.FC<TorrentPeersProps> = ({}) => {
  if (mockPeers.length === 0) {
    return (
      <Box sx={{ textAlign: 'center', py: 4 }}>
        <Typography variant="body1" color="text.secondary">
          No connected peers
        </Typography>
      </Box>
    );
  }

  return (
    <TableContainer>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Address</TableCell>
            <TableCell>Client</TableCell>
            <TableCell align="center">Progress</TableCell>
            <TableCell align="right">Download</TableCell>
            <TableCell align="right">Upload</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {mockPeers.map((peer) => (
            <TableRow key={peer.id}>
              <TableCell>
                <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                  {peer.address}
                </Typography>
              </TableCell>
              <TableCell>
                <Typography variant="body2">{peer.client}</Typography>
              </TableCell>
              <TableCell align="center">
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Box sx={{ width: '100%', mr: 1 }}>
                    <LinearProgress variant="determinate" value={peer.progress} />
                  </Box>
                  <Box sx={{ minWidth: 35 }}>
                    <Typography variant="body2" color="text.secondary">
                      {`${Math.round(peer.progress)}%`}
                    </Typography>
                  </Box>
                </Box>
              </TableCell>
              <TableCell align="right">
                <Typography variant="body2">{formatSpeed(peer.downloadSpeed)}</Typography>
              </TableCell>
              <TableCell align="right">
                <Typography variant="body2">{formatSpeed(peer.uploadSpeed)}</Typography>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default TorrentPeers;
