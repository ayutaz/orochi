import React from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  Chip,
} from '@mui/material';

// Mock data for now - this would come from the API
const mockTrackers = [
  {
    url: 'udp://tracker.opentrackr.org:1337/announce',
    status: 'working' as const,
    peers: 150,
    lastAnnounce: '2 minutes ago',
    nextAnnounce: 'in 28 minutes',
  },
  {
    url: 'udp://tracker.openbittorrent.com:6969/announce',
    status: 'working' as const,
    peers: 89,
    lastAnnounce: '5 minutes ago',
    nextAnnounce: 'in 25 minutes',
  },
  {
    url: 'udp://tracker.example.com:8080/announce',
    status: 'error' as const,
    peers: 0,
    error: 'Connection timeout',
    lastAnnounce: '10 minutes ago',
    nextAnnounce: 'in 5 minutes',
  },
];

interface TorrentTrackersProps {
  torrentId: string;
}

const TorrentTrackers: React.FC<TorrentTrackersProps> = ({}) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'working':
        return 'success';
      case 'error':
        return 'error';
      case 'disabled':
        return 'default';
      default:
        return 'default';
    }
  };

  return (
    <TableContainer>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Tracker</TableCell>
            <TableCell align="center">Status</TableCell>
            <TableCell align="center">Peers</TableCell>
            <TableCell>Last Announce</TableCell>
            <TableCell>Next Announce</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {mockTrackers.map((tracker, index) => (
            <TableRow key={index}>
              <TableCell>
                <Typography variant="body2" sx={{ wordBreak: 'break-all' }}>
                  {tracker.url}
                </Typography>
                {tracker.error && (
                  <Typography variant="caption" color="error">
                    {tracker.error}
                  </Typography>
                )}
              </TableCell>
              <TableCell align="center">
                <Chip label={tracker.status} color={getStatusColor(tracker.status)} size="small" />
              </TableCell>
              <TableCell align="center">
                <Typography variant="body2">{tracker.peers}</Typography>
              </TableCell>
              <TableCell>
                <Typography variant="body2">{tracker.lastAnnounce}</Typography>
              </TableCell>
              <TableCell>
                <Typography variant="body2">{tracker.nextAnnounce}</Typography>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default TorrentTrackers;
