import React, { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import {
  Container,
  Paper,
  TextField,
  Button,
  Typography,
  Box,
  Alert,
  InputAdornment,
  Stepper,
  Step,
  StepLabel,
} from '@mui/material'
import { Email, Lock, VpnKey } from '@mui/icons-material'
import { useAuthStore } from '../store/useAuthStore'

interface ForgotPasswordFormData {
  email: string
  code: string
  newPassword: string
  confirmPassword: string
}

interface FormErrors {
  email?: string
  code?: string
  newPassword?: string
  confirmPassword?: string
}

export const ForgotPassword: React.FC = () => {
  const navigate = useNavigate()
  const { requestPasswordReset, resetPassword, loading, error } = useAuthStore()
  const [activeStep, setActiveStep] = useState(0)
  const [formData, setFormData] = useState<ForgotPasswordFormData>({
    email: '',
    code: '',
    newPassword: '',
    confirmPassword: '',
  })
  const [errors, setErrors] = useState<FormErrors>({})
  const [successMessage, setSuccessMessage] = useState('')

  const steps = ['Enter Email', 'Verify Code', 'Reset Password']

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }))
    // Clear error for this field
    setErrors((prev) => ({ ...prev, [name]: undefined }))
  }

  const validateEmail = (): boolean => {
    const newErrors: FormErrors = {}

    if (!formData.email.trim()) {
      newErrors.email = 'Email is required'
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Email is invalid'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const validateCode = (): boolean => {
    const newErrors: FormErrors = {}

    if (!formData.code.trim()) {
      newErrors.code = 'Verification code is required'
    } else if (formData.code.length !== 6) {
      newErrors.code = 'Code must be 6 digits'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const validatePassword = (): boolean => {
    const newErrors: FormErrors = {}

    if (!formData.newPassword) {
      newErrors.newPassword = 'Password is required'
    } else if (formData.newPassword.length < 8) {
      newErrors.newPassword = 'Password must be at least 8 characters'
    } else if (!/(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/.test(formData.newPassword)) {
      newErrors.newPassword = 'Password must contain uppercase, lowercase, and number'
    }

    if (!formData.confirmPassword) {
      newErrors.confirmPassword = 'Please confirm your password'
    } else if (formData.newPassword !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleEmailSubmit = async () => {
    if (!validateEmail()) return

    try {
      await requestPasswordReset(formData.email)
      setSuccessMessage('Verification code sent to your email')
      setActiveStep(1)
    } catch (err) {
      // Error handled by store
    }
  }

  const handleCodeSubmit = async () => {
    if (!validateCode()) return
    setActiveStep(2)
  }

  const handlePasswordSubmit = async () => {
    if (!validatePassword()) return

    try {
      await resetPassword(formData.email, formData.code, formData.newPassword)
      setSuccessMessage('Password reset successfully')
      setTimeout(() => {
        navigate('/login')
      }, 2000)
    } catch (err) {
      // Error handled by store
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    switch (activeStep) {
      case 0:
        handleEmailSubmit()
        break
      case 1:
        handleCodeSubmit()
        break
      case 2:
        handlePasswordSubmit()
        break
    }
  }

  const renderStepContent = () => {
    switch (activeStep) {
      case 0:
        return (
          <>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Enter your email address and we'll send you a code to reset your password.
            </Typography>
            <TextField
              margin="normal"
              required
              fullWidth
              id="email"
              label="Email Address"
              name="email"
              autoComplete="email"
              autoFocus
              value={formData.email}
              onChange={handleChange}
              error={!!errors.email}
              helperText={errors.email}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <Email color="action" />
                  </InputAdornment>
                ),
              }}
            />
          </>
        )
      case 1:
        return (
          <>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Enter the 6-digit verification code sent to {formData.email}
            </Typography>
            <TextField
              margin="normal"
              required
              fullWidth
              id="code"
              label="Verification Code"
              name="code"
              autoComplete="off"
              autoFocus
              value={formData.code}
              onChange={handleChange}
              error={!!errors.code}
              helperText={errors.code}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <VpnKey color="action" />
                  </InputAdornment>
                ),
              }}
              inputProps={{
                maxLength: 6,
                pattern: '[0-9]*',
              }}
            />
            <Button
              variant="text"
              size="small"
              onClick={handleEmailSubmit}
              sx={{ mt: 1 }}
            >
              Resend Code
            </Button>
          </>
        )
      case 2:
        return (
          <>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Enter your new password
            </Typography>
            <TextField
              margin="normal"
              required
              fullWidth
              name="newPassword"
              label="New Password"
              type="password"
              id="newPassword"
              autoComplete="new-password"
              autoFocus
              value={formData.newPassword}
              onChange={handleChange}
              error={!!errors.newPassword}
              helperText={errors.newPassword}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <Lock color="action" />
                  </InputAdornment>
                ),
              }}
            />
            <TextField
              margin="normal"
              required
              fullWidth
              name="confirmPassword"
              label="Confirm New Password"
              type="password"
              id="confirmPassword"
              autoComplete="new-password"
              value={formData.confirmPassword}
              onChange={handleChange}
              error={!!errors.confirmPassword}
              helperText={errors.confirmPassword}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <Lock color="action" />
                  </InputAdornment>
                ),
              }}
            />
          </>
        )
      default:
        return null
    }
  }

  return (
    <Container component="main" maxWidth="sm">
      <Box
        sx={{
          marginTop: 8,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
        }}
      >
        <Paper elevation={3} sx={{ padding: 4, width: '100%' }}>
          <Typography component="h1" variant="h4" align="center" gutterBottom>
            Reset Password
          </Typography>

          <Stepper activeStep={activeStep} sx={{ mb: 4 }}>
            {steps.map((label) => (
              <Step key={label}>
                <StepLabel>{label}</StepLabel>
              </Step>
            ))}
          </Stepper>

          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          {successMessage && (
            <Alert severity="success" sx={{ mb: 2 }}>
              {successMessage}
            </Alert>
          )}

          <Box component="form" onSubmit={handleSubmit} noValidate>
            {renderStepContent()}

            <Button
              type="submit"
              fullWidth
              variant="contained"
              sx={{ mt: 3, mb: 2 }}
              disabled={loading}
            >
              {loading ? 'Processing...' : activeStep === 2 ? 'Reset Password' : 'Continue'}
            </Button>

            <Box sx={{ textAlign: 'center' }}>
              <Typography variant="body2">
                Remember your password?{' '}
                <Link to="/login" style={{ textDecoration: 'none' }}>
                  Sign in
                </Link>
              </Typography>
            </Box>
          </Box>
        </Paper>
      </Box>
    </Container>
  )
}