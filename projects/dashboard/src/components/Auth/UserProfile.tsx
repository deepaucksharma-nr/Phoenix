import React, { useState } from 'react'
import {
  Box,
  Avatar,
  Button,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  Divider,
  Typography,
  IconButton,
  Chip,
} from '@mui/material'
import {
  AccountCircle,
  Settings,
  Logout,
  Security,
  Person,
  Business,
  Email,
} from '@mui/icons-material'
import { useAppSelector, useAppDispatch } from '@hooks/redux'
import { logout } from '@store/slices/authSlice'
import { useNavigate } from 'react-router-dom'

interface UserProfileProps {
  variant?: 'menu' | 'card'
  showFullProfile?: boolean
}

export const UserProfile: React.FC<UserProfileProps> = ({ 
  variant = 'menu',
  showFullProfile = false 
}) => {
  const navigate = useNavigate()
  const dispatch = useAppDispatch()
  const user = useAppSelector(state => state.auth.user)
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)
  const open = Boolean(anchorEl)

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget)
  }

  const handleClose = () => {
    setAnchorEl(null)
  }

  const handleLogout = async () => {
    await dispatch(logout())
    navigate('/login')
  }

  const handleSettings = () => {
    navigate('/settings')
    handleClose()
  }

  const handleProfile = () => {
    navigate('/profile')
    handleClose()
  }

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'admin':
        return 'error'
      case 'user':
        return 'primary'
      case 'viewer':
        return 'secondary'
      default:
        return 'default'
    }
  }

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map(n => n[0])
      .join('')
      .toUpperCase()
  }

  if (!user) {
    return null
  }

  if (variant === 'card') {
    return (
      <Box sx={{ p: 3, maxWidth: 400 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
          <Avatar sx={{ width: 64, height: 64, mr: 2, bgcolor: 'primary.main' }}>
            {getInitials(user.name)}
          </Avatar>
          <Box sx={{ flex: 1 }}>
            <Typography variant="h6" gutterBottom>
              {user.name}
            </Typography>
            <Chip 
              label={user.role.toUpperCase()} 
              size="small" 
              color={getRoleColor(user.role) as any}
              sx={{ mb: 1 }}
            />
            <Typography variant="body2" color="text.secondary">
              {user.email}
            </Typography>
          </Box>
        </Box>

        {showFullProfile && (
          <Box sx={{ mb: 3 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
              <Person sx={{ mr: 1, fontSize: 16, color: 'text.secondary' }} />
              <Typography variant="body2" color="text.secondary">
                Full Name
              </Typography>
            </Box>
            <Typography variant="body1" sx={{ mb: 2, ml: 3 }}>
              {user.name}
            </Typography>

            <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
              <Email sx={{ mr: 1, fontSize: 16, color: 'text.secondary' }} />
              <Typography variant="body2" color="text.secondary">
                Email Address
              </Typography>
            </Box>
            <Typography variant="body1" sx={{ mb: 2, ml: 3 }}>
              {user.email}
            </Typography>

            <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
              <Security sx={{ mr: 1, fontSize: 16, color: 'text.secondary' }} />
              <Typography variant="body2" color="text.secondary">
                Role
              </Typography>
            </Box>
            <Typography variant="body1" sx={{ mb: 2, ml: 3 }}>
              {user.role.charAt(0).toUpperCase() + user.role.slice(1)}
            </Typography>
          </Box>
        )}

        <Box sx={{ display: 'flex', gap: 1 }}>
          <Button
            variant="outlined"
            size="small"
            startIcon={<Settings />}
            onClick={handleSettings}
            sx={{ flex: 1 }}
          >
            Settings
          </Button>
          <Button
            variant="outlined"
            size="small"
            startIcon={<Logout />}
            onClick={handleLogout}
            color="error"
            sx={{ flex: 1 }}
          >
            Logout
          </Button>
        </Box>
      </Box>
    )
  }

  return (
    <>
      <IconButton
        onClick={handleClick}
        size="small"
        sx={{ ml: 2 }}
        aria-controls={open ? 'account-menu' : undefined}
        aria-haspopup="true"
        aria-expanded={open ? 'true' : undefined}
      >
        <Avatar sx={{ width: 32, height: 32, bgcolor: 'primary.main' }}>
          {getInitials(user.name)}
        </Avatar>
      </IconButton>

      <Menu
        anchorEl={anchorEl}
        id="account-menu"
        open={open}
        onClose={handleClose}
        onClick={handleClose}
        PaperProps={{
          elevation: 3,
          sx: {
            overflow: 'visible',
            filter: 'drop-shadow(0px 2px 8px rgba(0,0,0,0.32))',
            mt: 1.5,
            minWidth: 240,
            '& .MuiAvatar-root': {
              width: 32,
              height: 32,
              ml: -0.5,
              mr: 1,
            },
            '&:before': {
              content: '""',
              display: 'block',
              position: 'absolute',
              top: 0,
              right: 14,
              width: 10,
              height: 10,
              bgcolor: 'background.paper',
              transform: 'translateY(-50%) rotate(45deg)',
              zIndex: 0,
            },
          },
        }}
        transformOrigin={{ horizontal: 'right', vertical: 'top' }}
        anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
      >
        <Box sx={{ px: 2, py: 1 }}>
          <Typography variant="subtitle2" fontWeight={600}>
            {user.name}
          </Typography>
          <Typography variant="caption" color="text.secondary">
            {user.email}
          </Typography>
          <Box sx={{ mt: 0.5 }}>
            <Chip 
              label={user.role.toUpperCase()} 
              size="small" 
              color={getRoleColor(user.role) as any}
            />
          </Box>
        </Box>
        
        <Divider />
        
        <MenuItem onClick={handleProfile}>
          <ListItemIcon>
            <AccountCircle fontSize="small" />
          </ListItemIcon>
          <ListItemText>Profile</ListItemText>
        </MenuItem>
        
        <MenuItem onClick={handleSettings}>
          <ListItemIcon>
            <Settings fontSize="small" />
          </ListItemIcon>
          <ListItemText>Settings</ListItemText>
        </MenuItem>
        
        <Divider />
        
        <MenuItem onClick={handleLogout} sx={{ color: 'error.main' }}>
          <ListItemIcon>
            <Logout fontSize="small" color="error" />
          </ListItemIcon>
          <ListItemText>Logout</ListItemText>
        </MenuItem>
      </Menu>
    </>
  )
}