import React, { useState } from 'react'
import {
  Box,
  Paper,
  Typography,
  Tabs,
  Tab,
  TextField,
  Button,
  Switch,
  FormControlLabel,
  FormControl,
  FormLabel,
  RadioGroup,
  Radio,
  Select,
  MenuItem,
  InputLabel,
  Divider,
  Alert,
  Grid,
  Card,
  CardContent,
  CardActions,
  Chip,
  IconButton,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  DialogContentText,
} from '@mui/material'
import {
  Save as SaveIcon,
  ContentCopy as CopyIcon,
  Add as AddIcon,
  Delete as DeleteIcon,
  Refresh as RefreshIcon,
  Warning as WarningIcon,
} from '@mui/icons-material'
import { useAuthStore } from '../store/useAuthStore'
import { useNotification } from '../hooks/useNotification'

interface TabPanelProps {
  children?: React.ReactNode
  index: number
  value: number
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`settings-tabpanel-${index}`}
      aria-labelledby={`settings-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  )
}

export const Settings: React.FC = () => {
  const { user } = useAuthStore()
  const { showNotification } = useNotification()
  const [tabValue, setTabValue] = useState(0)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  
  // Profile settings
  const [profile, setProfile] = useState({
    name: user?.name || '',
    email: user?.email || '',
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
  })

  // Notification preferences
  const [notifications, setNotifications] = useState({
    emailAlerts: true,
    experimentComplete: true,
    experimentFailed: true,
    costThreshold: true,
    weeklyReport: false,
    realTimeUpdates: true,
  })

  // Platform settings
  const [platform, setPlatform] = useState({
    theme: 'light',
    timezone: 'UTC',
    dateFormat: 'MM/DD/YYYY',
    refreshInterval: 5,
    autoRefresh: true,
    compactView: false,
  })

  // API keys
  const [apiKeys, setApiKeys] = useState([
    {
      id: '1',
      name: 'Production API Key',
      key: 'phoenix_prod_xxxxxxxxxxxxxxxxxxx',
      created: '2025-01-15',
      lastUsed: '2025-01-20',
    },
  ])

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue)
  }

  const handleProfileSave = () => {
    if (profile.newPassword && profile.newPassword !== profile.confirmPassword) {
      showNotification('Passwords do not match', 'error')
      return
    }
    showNotification('Profile updated successfully', 'success')
  }

  const handleNotificationSave = () => {
    showNotification('Notification preferences updated', 'success')
  }

  const handlePlatformSave = () => {
    showNotification('Platform settings updated', 'success')
  }

  const generateApiKey = () => {
    const newKey = {
      id: Date.now().toString(),
      name: 'New API Key',
      key: `phoenix_${Math.random().toString(36).substring(2)}`,
      created: new Date().toISOString().split('T')[0],
      lastUsed: 'Never',
    }
    setApiKeys([...apiKeys, newKey])
    showNotification('API key generated', 'success')
  }

  const deleteApiKey = (id: string) => {
    setApiKeys(apiKeys.filter(key => key.id !== id))
    showNotification('API key deleted', 'success')
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    showNotification('Copied to clipboard', 'success')
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Settings
      </Typography>
      
      <Paper sx={{ mt: 3 }}>
        <Tabs
          value={tabValue}
          onChange={handleTabChange}
          variant="scrollable"
          scrollButtons="auto"
        >
          <Tab label="Profile" />
          <Tab label="Notifications" />
          <Tab label="Platform" />
          <Tab label="API Keys" />
          <Tab label="Danger Zone" />
        </Tabs>

        <TabPanel value={tabValue} index={0}>
          <Grid container spacing={3}>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Name"
                value={profile.name}
                onChange={(e) => setProfile({ ...profile, name: e.target.value })}
                margin="normal"
              />
              <TextField
                fullWidth
                label="Email"
                type="email"
                value={profile.email}
                onChange={(e) => setProfile({ ...profile, email: e.target.value })}
                margin="normal"
                disabled
                helperText="Contact support to change your email"
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <Typography variant="h6" gutterBottom sx={{ mt: 2 }}>
                Change Password
              </Typography>
              <TextField
                fullWidth
                label="Current Password"
                type="password"
                value={profile.currentPassword}
                onChange={(e) => setProfile({ ...profile, currentPassword: e.target.value })}
                margin="normal"
              />
              <TextField
                fullWidth
                label="New Password"
                type="password"
                value={profile.newPassword}
                onChange={(e) => setProfile({ ...profile, newPassword: e.target.value })}
                margin="normal"
              />
              <TextField
                fullWidth
                label="Confirm New Password"
                type="password"
                value={profile.confirmPassword}
                onChange={(e) => setProfile({ ...profile, confirmPassword: e.target.value })}
                margin="normal"
              />
            </Grid>
            <Grid item xs={12}>
              <Button
                variant="contained"
                startIcon={<SaveIcon />}
                onClick={handleProfileSave}
                sx={{ mt: 2 }}
              >
                Save Profile
              </Button>
            </Grid>
          </Grid>
        </TabPanel>

        <TabPanel value={tabValue} index={1}>
          <Typography variant="h6" gutterBottom>
            Email Notifications
          </Typography>
          <FormControlLabel
            control={
              <Switch
                checked={notifications.emailAlerts}
                onChange={(e) => setNotifications({ ...notifications, emailAlerts: e.target.checked })}
              />
            }
            label="Enable email alerts"
          />
          <Box sx={{ ml: 4, mt: 2 }}>
            <FormControlLabel
              control={
                <Switch
                  checked={notifications.experimentComplete}
                  onChange={(e) => setNotifications({ ...notifications, experimentComplete: e.target.checked })}
                  disabled={!notifications.emailAlerts}
                />
              }
              label="Experiment completed"
            />
            <FormControlLabel
              control={
                <Switch
                  checked={notifications.experimentFailed}
                  onChange={(e) => setNotifications({ ...notifications, experimentFailed: e.target.checked })}
                  disabled={!notifications.emailAlerts}
                />
              }
              label="Experiment failed"
            />
            <FormControlLabel
              control={
                <Switch
                  checked={notifications.costThreshold}
                  onChange={(e) => setNotifications({ ...notifications, costThreshold: e.target.checked })}
                  disabled={!notifications.emailAlerts}
                />
              }
              label="Cost threshold exceeded"
            />
            <FormControlLabel
              control={
                <Switch
                  checked={notifications.weeklyReport}
                  onChange={(e) => setNotifications({ ...notifications, weeklyReport: e.target.checked })}
                  disabled={!notifications.emailAlerts}
                />
              }
              label="Weekly summary report"
            />
          </Box>
          
          <Typography variant="h6" gutterBottom sx={{ mt: 4 }}>
            Real-time Updates
          </Typography>
          <FormControlLabel
            control={
              <Switch
                checked={notifications.realTimeUpdates}
                onChange={(e) => setNotifications({ ...notifications, realTimeUpdates: e.target.checked })}
              />
            }
            label="Enable real-time WebSocket updates"
          />
          
          <Button
            variant="contained"
            startIcon={<SaveIcon />}
            onClick={handleNotificationSave}
            sx={{ mt: 3 }}
          >
            Save Notifications
          </Button>
        </TabPanel>

        <TabPanel value={tabValue} index={2}>
          <Grid container spacing={3}>
            <Grid item xs={12} md={6}>
              <FormControl fullWidth margin="normal">
                <FormLabel>Theme</FormLabel>
                <RadioGroup
                  value={platform.theme}
                  onChange={(e) => setPlatform({ ...platform, theme: e.target.value })}
                >
                  <FormControlLabel value="light" control={<Radio />} label="Light" />
                  <FormControlLabel value="dark" control={<Radio />} label="Dark (Coming Soon)" disabled />
                  <FormControlLabel value="auto" control={<Radio />} label="System" disabled />
                </RadioGroup>
              </FormControl>
              
              <FormControl fullWidth margin="normal">
                <InputLabel>Timezone</InputLabel>
                <Select
                  value={platform.timezone}
                  onChange={(e) => setPlatform({ ...platform, timezone: e.target.value })}
                  label="Timezone"
                >
                  <MenuItem value="UTC">UTC</MenuItem>
                  <MenuItem value="America/New_York">Eastern Time</MenuItem>
                  <MenuItem value="America/Chicago">Central Time</MenuItem>
                  <MenuItem value="America/Denver">Mountain Time</MenuItem>
                  <MenuItem value="America/Los_Angeles">Pacific Time</MenuItem>
                </Select>
              </FormControl>
              
              <FormControl fullWidth margin="normal">
                <InputLabel>Date Format</InputLabel>
                <Select
                  value={platform.dateFormat}
                  onChange={(e) => setPlatform({ ...platform, dateFormat: e.target.value })}
                  label="Date Format"
                >
                  <MenuItem value="MM/DD/YYYY">MM/DD/YYYY</MenuItem>
                  <MenuItem value="DD/MM/YYYY">DD/MM/YYYY</MenuItem>
                  <MenuItem value="YYYY-MM-DD">YYYY-MM-DD</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            
            <Grid item xs={12} md={6}>
              <Typography variant="h6" gutterBottom>
                Dashboard Settings
              </Typography>
              
              <FormControlLabel
                control={
                  <Switch
                    checked={platform.autoRefresh}
                    onChange={(e) => setPlatform({ ...platform, autoRefresh: e.target.checked })}
                  />
                }
                label="Auto-refresh data"
              />
              
              <TextField
                fullWidth
                type="number"
                label="Refresh interval (seconds)"
                value={platform.refreshInterval}
                onChange={(e) => setPlatform({ ...platform, refreshInterval: parseInt(e.target.value) })}
                margin="normal"
                disabled={!platform.autoRefresh}
                inputProps={{ min: 5, max: 300 }}
              />
              
              <FormControlLabel
                control={
                  <Switch
                    checked={platform.compactView}
                    onChange={(e) => setPlatform({ ...platform, compactView: e.target.checked })}
                  />
                }
                label="Compact view"
              />
            </Grid>
            
            <Grid item xs={12}>
              <Button
                variant="contained"
                startIcon={<SaveIcon />}
                onClick={handlePlatformSave}
                sx={{ mt: 2 }}
              >
                Save Platform Settings
              </Button>
            </Grid>
          </Grid>
        </TabPanel>

        <TabPanel value={tabValue} index={3}>
          <Alert severity="info" sx={{ mb: 3 }}>
            Use API keys to authenticate external applications and CI/CD pipelines with Phoenix.
          </Alert>
          
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={generateApiKey}
            sx={{ mb: 3 }}
          >
            Generate New API Key
          </Button>
          
          <List>
            {apiKeys.map((apiKey) => (
              <Card key={apiKey.id} sx={{ mb: 2 }}>
                <CardContent>
                  <Typography variant="h6">{apiKey.name}</Typography>
                  <Box sx={{ display: 'flex', alignItems: 'center', mt: 1 }}>
                    <TextField
                      value={apiKey.key}
                      size="small"
                      fullWidth
                      InputProps={{
                        readOnly: true,
                        endAdornment: (
                          <IconButton onClick={() => copyToClipboard(apiKey.key)}>
                            <CopyIcon />
                          </IconButton>
                        ),
                      }}
                    />
                  </Box>
                  <Box sx={{ mt: 2 }}>
                    <Chip label={`Created: ${apiKey.created}`} size="small" sx={{ mr: 1 }} />
                    <Chip label={`Last used: ${apiKey.lastUsed}`} size="small" />
                  </Box>
                </CardContent>
                <CardActions>
                  <Button
                    size="small"
                    color="error"
                    startIcon={<DeleteIcon />}
                    onClick={() => deleteApiKey(apiKey.id)}
                  >
                    Delete
                  </Button>
                </CardActions>
              </Card>
            ))}
          </List>
        </TabPanel>

        <TabPanel value={tabValue} index={4}>
          <Alert severity="error" sx={{ mb: 3 }}>
            <Typography variant="h6">Danger Zone</Typography>
            These actions are irreversible. Please be certain.
          </Alert>
          
          <Card sx={{ border: '1px solid', borderColor: 'error.main' }}>
            <CardContent>
              <Typography variant="h6" color="error">
                Delete Account
              </Typography>
              <Typography variant="body2" sx={{ mt: 1 }}>
                Once you delete your account, there is no going back. All your experiments,
                configurations, and data will be permanently deleted.
              </Typography>
            </CardContent>
            <CardActions>
              <Button
                variant="outlined"
                color="error"
                startIcon={<WarningIcon />}
                onClick={() => setDeleteDialogOpen(true)}
              >
                Delete Account
              </Button>
            </CardActions>
          </Card>
        </TabPanel>
      </Paper>

      <Dialog
        open={deleteDialogOpen}
        onClose={() => setDeleteDialogOpen(false)}
      >
        <DialogTitle>Delete Account</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Are you absolutely sure you want to delete your account? This action cannot be undone.
            All your experiments, configurations, and data will be permanently deleted.
          </DialogContentText>
          <TextField
            autoFocus
            margin="dense"
            label="Type DELETE to confirm"
            fullWidth
            variant="outlined"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>Cancel</Button>
          <Button color="error" variant="contained">
            Delete Account
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}