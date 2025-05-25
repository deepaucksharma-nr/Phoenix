import React, { ReactElement } from 'react'
import { render, RenderOptions } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ThemeProvider, createTheme } from '@mui/material/styles'
import { NotificationProvider } from '@/components/Notifications/NotificationProvider'

const theme = createTheme()

interface AllTheProvidersProps {
  children: React.ReactNode
}

const AllTheProviders: React.FC<AllTheProvidersProps> = ({ children }) => {
  return (
    <BrowserRouter>
      <ThemeProvider theme={theme}>
        <NotificationProvider>{children}</NotificationProvider>
      </ThemeProvider>
    </BrowserRouter>
  )
}

const customRender = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) => render(ui, { wrapper: AllTheProviders, ...options })

// Re-export everything
export * from '@testing-library/react'
export { customRender as render }

// Mock data generators
export const mockExperiment = (overrides = {}) => ({
  id: 'exp-123',
  name: 'Test Experiment',
  description: 'Test description',
  spec: {
    baseline: {
      name: 'baseline',
      receivers: ['hostmetrics'],
      processors: [{ type: 'memory_limiter', config: {} }],
      exporters: ['prometheus'],
    },
    candidate: {
      name: 'candidate',
      receivers: ['hostmetrics'],
      processors: [
        { type: 'memory_limiter', config: {} },
        { type: 'filter/priority', config: { minPriority: 'high' } },
      ],
      exporters: ['prometheus'],
    },
    targetHosts: ['host-1', 'host-2'],
    duration: '1h',
    loadProfile: 'realistic',
  },
  status: 'running',
  owner: 'user@example.com',
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
})

export const mockUser = (overrides = {}) => ({
  id: 'user-123',
  email: 'user@example.com',
  name: 'Test User',
  role: 'admin',
  organization: 'Test Org',
  ...overrides,
})

export const mockPipeline = (overrides = {}) => ({
  id: 'pipeline-123',
  name: 'Test Pipeline',
  receivers: ['hostmetrics'],
  processors: [
    { type: 'memory_limiter', config: { limit_mib: 512 } },
    { type: 'batch', config: { timeout: '10s' } },
  ],
  exporters: ['prometheus', 'newrelic'],
  ...overrides,
})

export const mockMetrics = (overrides = {}) => ({
  timestamp: Array.from({ length: 24 }, (_, i) => Date.now() - i * 3600000),
  values: Array.from({ length: 24 }, () => Math.random() * 100),
  metric: 'cardinality',
  variant: 'baseline',
  ...overrides,
})