import { describe, it, expect } from 'vitest'
import { render, screen } from '@/test/utils'
import { MetricCard } from '../MetricCard'
import { TrendingUp } from '@mui/icons-material'

describe('MetricCard', () => {
  it('renders title and value', () => {
    render(
      <MetricCard
        title="Test Metric"
        value="100"
      />
    )

    expect(screen.getByText('Test Metric')).toBeInTheDocument()
    expect(screen.getByText('100')).toBeInTheDocument()
  })

  it('renders with icon', () => {
    render(
      <MetricCard
        title="Test Metric"
        value="100"
        icon={<TrendingUp data-testid="metric-icon" />}
      />
    )

    expect(screen.getByTestId('metric-icon')).toBeInTheDocument()
  })

  it('shows positive change indicator', () => {
    render(
      <MetricCard
        title="Test Metric"
        value="100"
        change={25}
      />
    )

    expect(screen.getByText('+25.0%')).toBeInTheDocument()
  })

  it('shows negative change indicator', () => {
    render(
      <MetricCard
        title="Test Metric"
        value="100"
        change={-15.5}
      />
    )

    expect(screen.getByText('-15.5%')).toBeInTheDocument()
  })

  it('renders subtitle when provided', () => {
    render(
      <MetricCard
        title="Test Metric"
        value="100"
        subtitle="Additional information"
      />
    )

    expect(screen.getByText('Additional information')).toBeInTheDocument()
  })

  it('applies correct color theme', () => {
    const { container } = render(
      <MetricCard
        title="Test Metric"
        value="100"
        color="success"
        icon={<TrendingUp />}
      />
    )

    const iconWrapper = container.querySelector('[class*="bgcolor-success.light"]')
    expect(iconWrapper).toBeInTheDocument()
  })
})