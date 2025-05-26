import { describe, it, expect } from 'vitest'
import { render, screen } from '@/test/utils'
import { PipelineViewer } from '../PipelineViewer'
import { mockPipeline } from '@/test/utils'

describe('PipelineViewer', () => {
  it('renders pipeline configuration', () => {
    const pipeline = mockPipeline()
    
    render(<PipelineViewer pipeline={pipeline} />)

    // Check receivers section
    expect(screen.getByText('Receivers')).toBeInTheDocument()
    expect(screen.getByText('hostmetrics')).toBeInTheDocument()

    // Check exporters section
    expect(screen.getByText('Exporters')).toBeInTheDocument()
    expect(screen.getByText('prometheus')).toBeInTheDocument()
    expect(screen.getByText('newrelic')).toBeInTheDocument()

    // Check processors section
    expect(screen.getByText(/Processors \(2\)/)).toBeInTheDocument()
    expect(screen.getByText('memory_limiter')).toBeInTheDocument()
    expect(screen.getByText('batch')).toBeInTheDocument()
  })

  it('renders compact view when specified', () => {
    const pipeline = mockPipeline()
    
    render(<PipelineViewer pipeline={pipeline} compact />)

    // In compact mode, processors should be shown as chips
    expect(screen.getByText('memory_limiter')).toBeInTheDocument()
    expect(screen.getByText('batch')).toBeInTheDocument()
    
    // But not the full sections
    expect(screen.queryByText('Receivers')).not.toBeInTheDocument()
    expect(screen.queryByText('Exporters')).not.toBeInTheDocument()
  })

  it('handles empty pipeline gracefully', () => {
    render(<PipelineViewer pipeline={null as any} />)

    expect(screen.getByText('No pipeline configuration')).toBeInTheDocument()
  })

  it('displays processor configuration details', () => {
    const pipeline = mockPipeline({
      processors: [
        {
          type: 'filter/priority',
          config: {
            minPriority: 'high',
            includeProcesses: ['nginx', 'postgres'],
          },
        },
      ],
    })
    
    render(<PipelineViewer pipeline={pipeline} />)

    expect(screen.getByText('filter/priority')).toBeInTheDocument()
    expect(screen.getByText(/minPriority:/)).toBeInTheDocument()
    expect(screen.getByText(/high/)).toBeInTheDocument()
  })

  it('handles processors without configuration', () => {
    const pipeline = mockPipeline({
      processors: [
        {
          type: 'noop',
          config: null,
        },
      ],
    })
    
    render(<PipelineViewer pipeline={pipeline} />)

    expect(screen.getByText('noop')).toBeInTheDocument()
    expect(screen.getByText('No configuration')).toBeInTheDocument()
  })
})