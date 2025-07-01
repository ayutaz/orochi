import React, { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  AppBar,
  Box,
  Drawer,
  IconButton,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Toolbar,
  Typography,
  useTheme as useMuiTheme,
  useMediaQuery,
} from '@mui/material'
import {
  Menu as MenuIcon,
  CloudDownload as CloudDownloadIcon,
  Settings as SettingsIcon,
  Brightness4 as Brightness4Icon,
  Brightness7 as Brightness7Icon,
} from '@mui/icons-material'
import { useTheme } from '../contexts/ThemeContext'

const drawerWidth = 240

interface LayoutProps {
  children: React.ReactNode
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const muiTheme = useMuiTheme()
  const isMobile = useMediaQuery(muiTheme.breakpoints.down('sm'))
  const [mobileOpen, setMobileOpen] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { t } = useTranslation()
  const { darkMode, toggleTheme } = useTheme()

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen)
  }

  const menuItems = [
    { text: t('navigation.torrents'), icon: <CloudDownloadIcon />, path: '/' },
    { text: t('navigation.settings'), icon: <SettingsIcon />, path: '/settings' },
  ]

  const drawer = (
    <Box>
      <Toolbar>
        <Typography variant="h6" noWrap component="div">
          Orochi
        </Typography>
      </Toolbar>
      <List>
        {menuItems.map((item) => (
          <ListItem key={item.path} disablePadding>
            <ListItemButton
              selected={location.pathname === item.path}
              onClick={() => {
                navigate(item.path)
                if (isMobile) {
                  setMobileOpen(false)
                }
              }}
            >
              <ListItemIcon>{item.icon}</ListItemIcon>
              <ListItemText primary={item.text} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
    </Box>
  )

  return (
    <Box sx={{ display: 'flex' }}>
      <AppBar
        position="fixed"
        sx={{
          width: { sm: `calc(100% - ${drawerWidth}px)` },
          ml: { sm: `${drawerWidth}px` },
        }}
      >
        <Toolbar>
          <IconButton
            color="inherit"
            aria-label="open drawer"
            edge="start"
            onClick={handleDrawerToggle}
            sx={{ mr: 2, display: { sm: 'none' } }}
          >
            <MenuIcon />
          </IconButton>
          <Typography variant="h6" noWrap component="div" sx={{ flexGrow: 1 }}>
            {menuItems.find((item) => item.path === location.pathname)?.text || 'Orochi'}
          </Typography>
          <IconButton sx={{ ml: 1 }} onClick={toggleTheme} color="inherit">
            {darkMode ? <Brightness7Icon /> : <Brightness4Icon />}
          </IconButton>
        </Toolbar>
      </AppBar>
      <Box
        component="nav"
        sx={{ width: { sm: drawerWidth }, flexShrink: { sm: 0 } }}
      >
        <Drawer
          variant={isMobile ? 'temporary' : 'permanent'}
          open={isMobile ? mobileOpen : true}
          onClose={handleDrawerToggle}
          ModalProps={{
            keepMounted: true,
          }}
          sx={{
            '& .MuiDrawer-paper': {
              boxSizing: 'border-box',
              width: drawerWidth,
            },
          }}
        >
          {drawer}
        </Drawer>
      </Box>
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 3,
          width: { sm: `calc(100% - ${drawerWidth}px)` },
          mt: 8,
        }}
      >
        {children}
      </Box>
    </Box>
  )
}

export default Layout