import { describe, it, expect } from 'vitest'
import { render, screen } from '@/test/utils'
import { ComparisonTable } from '../ComparisonTable'
import { MetricsSummary } from '@/types'

describe('ComparisonTable', () => {
  const mockBaselineMetrics: MetricsSummary = {
    cardinality: 150000,
    cpuUsage: 12.5,
    memoryUsage: 2048,
    networkTraffic: 125,
    dataPointsPerSecond: 5000,
    uniqueProcesses: 450,
  }

  const mockCandidateMetrics: MetricsSummary = {
    cardinality: 52500,
    cpuUsage: 15.0,
    memoryUsage: 2150,
    networkTraffic: 45,
    dataPointsPerSecond: 1750,
    uniqueProcesses: 125,
  }

  it('renders all metrics with correct values', () => {
    render(
      <ComparisonTable
        baseline={mockBaselineMetrics}
        candidate={mockCandidateMetrics}
      />
    )

    // Check headers
    expect(screen.getByText('Metric')).toBeInTheDocument()
    expect(screen.getByText('Baseline')).toBeInTheDocument()
    expect(screen.getByText('Candidate')).toBeInTheDocument()
    expect(screen.getByText('Change')).toBeInTheDocument()
    expect(screen.getByText('Impact')).toBeInTheDocument()

    // Check metric names
    expect(screen.getByText('Time Series Cardinality')).toBeInTheDocument()
    expect(screen.getByText('CPU Usage')).toBeInTheDocument()
    expect(screen.getByText('Memory Usage')).toBeInTheDocument()
    expect(screen.getByText('Network Traffic')).toBeInTheDocument()
    expect(screen.getByText('Data Points/Second')).toBeInTheDocument()
    expect(screen.getByText('Unique Processes')).toBeInTheDocument()

    // Check baseline values
    expect(screen.getByText('150,000')).toBeInTheDocument() // cardinality
    expect(screen.getByText('12.5%')).toBeInTheDocument() // cpu
    expect(screen.getByText('2048 MB')).toBeInTheDocument() // memory

    // Check candidate values
    expect(screen.getByText('52,500')).toBeInTheDocument() // cardinality
    expect(screen.getByText('15.0%')).toBeInTheDocument() // cpu
    expect(screen.getByText('2150 MB')).toBeInTheDocument() // memory
  })

  it('calculates and displays percentage changes correctly', () => {
    render(
      <ComparisonTable
        baseline={mockBaselineMetrics}
        candidate={mockCandidateMetrics}
      />
    )

    // Cardinality: (52500 - 150000) / 150000 * 100 = -65%
    expect(screen.getByText('-65.0%')).toBeInTheDocument()

    // CPU: (15 - 12.5) / 12.5 * 100 = +20%
    expect(screen.getByText('+20.0%')).toBeInTheDocument()

    // Memory: (2150 - 2048) / 2048 * 100 = +5%
    expect(screen.getByText('+5.0%')).toBeInTheDocument()

    // Network: (45 - 125) / 125 * 100 = -64%
    expect(screen.getByText('-64.0%')).toBeInTheDocument()
  })

  it('displays appropriate impact levels', () => {
    render(
      <ComparisonTable
        baseline={mockBaselineMetrics}
        candidate={mockCandidateMetrics}
      />
    )

    // High impact for large changes (>50%)
    const highImpactChips = screen.getAllByText('High')
    expect(highImpactChips.length).toBeGreaterThan(0)

    // Medium impact for moderate changes (10-50%)
    const mediumImpactChips = screen.getAllByText('Medium')
    expect(mediumImpactChips.length).toBeGreaterThan(0)

    // Low impact for small changes (<10%)
    const lowImpactChips = screen.getAllByText('Low')
    expect(lowImpactChips.length).toBeGreaterThan(0)
  })

  it('handles zero baseline values gracefully', () => {
    const baselineWithZero: MetricsSummary = {
      ...mockBaselineMetrics,
      networkTraffic: 0,
    }

    render(
      <ComparisonTable
        baseline={baselineWithZero}
        candidate={mockCandidateMetrics}
      />
    )

    // Should not crash and should display values
    expect(screen.getByText('0 KB/s')).toBeInTheDocument()
  })

  it('shows no change for identical values', () => {
    const identicalCandidate: MetricsSummary = {
      ...mockBaselineMetrics,
    }

    render(
      <ComparisonTable
        baseline={mockBaselineMetrics}
        candidate={identicalCandidate}
      />
    )

    // Should show 0% change
    const zeroChanges = screen.getAllByText('0.0%')
    expect(zeroChanges.length).toBe(6) // All 6 metrics

    // Should show "No Change" impact
    const noChangeChips = screen.getAllByText('No Change')
    expect(noChangeChips.length).toBe(6)
  })

  it('applies correct colors based on metric type and change direction', () => {
    const { container } = render(
      <ComparisonTable
        baseline={mockBaselineMetrics}
        candidate={mockCandidateMetrics}
      />
    )

    // For cardinality reduction (good), should use success color
    const cardinalityRow = container.querySelector('tr:nth-child(1)')
    const cardinalityChange = cardinalityRow?.querySelector('[class*="color-success"]')
    expect(cardinalityChange).toBeInTheDocument()

    // For CPU increase (bad), should use error color
    const cpuRow = container.querySelector('tr:nth-child(2)')
    const cpuChange = cpuRow?.querySelector('[class*="color-error"]')
    expect(cpuChange).toBeInTheDocument()
  })
})