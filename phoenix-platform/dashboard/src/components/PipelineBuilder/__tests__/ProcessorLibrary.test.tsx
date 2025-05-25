import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@/test/utils'
import { ProcessorLibrary } from '../ProcessorLibrary'

// Mock React DnD
vi.mock('react-dnd', () => ({
  useDrag: () => [{ isDragging: false }, vi.fn()],
}))

describe('ProcessorLibrary', () => {
  it('renders all processor categories', () => {
    render(<ProcessorLibrary />)

    expect(screen.getByText('Filters')).toBeInTheDocument()
    expect(screen.getByText('Transforms')).toBeInTheDocument()
    expect(screen.getByText('Aggregators')).toBeInTheDocument()
    expect(screen.getByText('Utilities')).toBeInTheDocument()
  })

  it('displays processors in each category', () => {
    render(<ProcessorLibrary />)

    // Filter processors
    expect(screen.getByText('Priority Filter')).toBeInTheDocument()
    expect(screen.getByText('Top K Filter')).toBeInTheDocument()
    expect(screen.getByText('Resource Filter')).toBeInTheDocument()

    // Transform processors
    expect(screen.getByText('Process Classifier')).toBeInTheDocument()
    expect(screen.getByText('Metric Transform')).toBeInTheDocument()
    expect(screen.getByText('Add Attributes')).toBeInTheDocument()

    // Aggregator processors
    expect(screen.getByText('Group by Attributes')).toBeInTheDocument()
    expect(screen.getByText('Span Metrics')).toBeInTheDocument()

    // Utility processors
    expect(screen.getByText('Memory Limiter')).toBeInTheDocument()
    expect(screen.getByText('Batch Processor')).toBeInTheDocument()
  })

  it('shows processor descriptions', () => {
    render(<ProcessorLibrary />)

    expect(
      screen.getByText('Filter processes by priority (critical, high, medium, low)')
    ).toBeInTheDocument()
    
    expect(
      screen.getByText('Keep only top K processes by CPU or memory usage')
    ).toBeInTheDocument()
  })

  it('displays processor icons', () => {
    const { container } = render(<ProcessorLibrary />)

    // Check for MUI icons
    const filterIcons = container.querySelectorAll('[data-testid="FilterListIcon"]')
    expect(filterIcons.length).toBeGreaterThan(0)

    const transformIcons = container.querySelectorAll('[data-testid="TransformIcon"]')
    expect(transformIcons.length).toBeGreaterThan(0)
  })

  it('filters processors by search term', () => {
    render(<ProcessorLibrary />)

    const searchInput = screen.getByPlaceholderText('Search processors...')
    
    // Search for "memory"
    fireEvent.change(searchInput, { target: { value: 'memory' } })

    // Should show memory-related processors
    expect(screen.getByText('Memory Limiter')).toBeInTheDocument()
    expect(screen.getByText('Resource Filter')).toBeInTheDocument()

    // Should hide non-matching processors
    expect(screen.queryByText('Priority Filter')).not.toBeInTheDocument()
    expect(screen.queryByText('Batch Processor')).not.toBeInTheDocument()
  })

  it('collapses and expands categories', () => {
    render(<ProcessorLibrary />)

    const filtersAccordion = screen.getByText('Filters').closest('button')
    
    // Initially expanded
    expect(screen.getByText('Priority Filter')).toBeVisible()

    // Click to collapse
    fireEvent.click(filtersAccordion!)

    // Content should be hidden (but still in DOM)
    const priorityFilter = screen.getByText('Priority Filter')
    expect(priorityFilter.closest('[role="region"]')).toHaveAttribute('hidden')
  })

  it('shows empty state when no processors match search', () => {
    render(<ProcessorLibrary />)

    const searchInput = screen.getByPlaceholderText('Search processors...')
    
    // Search for non-existent processor
    fireEvent.change(searchInput, { target: { value: 'nonexistent' } })

    expect(screen.getByText('No processors match your search')).toBeInTheDocument()
  })

  it('clears search when clear button is clicked', () => {
    render(<ProcessorLibrary />)

    const searchInput = screen.getByPlaceholderText('Search processors...')
    
    // Enter search term
    fireEvent.change(searchInput, { target: { value: 'memory' } })
    expect(searchInput).toHaveValue('memory')

    // Click clear button
    const clearButton = screen.getByLabelText('clear')
    fireEvent.click(clearButton)

    expect(searchInput).toHaveValue('')
    // All processors should be visible again
    expect(screen.getByText('Priority Filter')).toBeInTheDocument()
  })

  it('highlights matched text in search results', () => {
    render(<ProcessorLibrary />)

    const searchInput = screen.getByPlaceholderText('Search processors...')
    fireEvent.change(searchInput, { target: { value: 'filter' } })

    // Check if filter text is highlighted in results
    const filterElements = screen.getAllByText(/filter/i)
    expect(filterElements.length).toBeGreaterThan(0)
  })

  it('maintains accordion state during search', () => {
    render(<ProcessorLibrary />)

    // Collapse Utilities section
    const utilitiesAccordion = screen.getByText('Utilities').closest('button')
    fireEvent.click(utilitiesAccordion!)

    // Search for something
    const searchInput = screen.getByPlaceholderText('Search processors...')
    fireEvent.change(searchInput, { target: { value: 'batch' } })

    // Clear search
    fireEvent.change(searchInput, { target: { value: '' } })

    // Utilities should still be collapsed
    const batchProcessor = screen.getByText('Batch Processor')
    expect(batchProcessor.closest('[role="region"]')).toHaveAttribute('hidden')
  })
})