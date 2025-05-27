import { useState, useCallback, useMemo } from 'react'

interface UsePaginationOptions {
  initialPage?: number
  initialPageSize?: number
  pageSizeOptions?: number[]
}

interface UsePaginationReturn {
  page: number
  pageSize: number
  pageSizeOptions: number[]
  setPage: (page: number) => void
  setPageSize: (size: number) => void
  nextPage: () => void
  prevPage: () => void
  firstPage: () => void
  lastPage: (totalPages: number) => void
  paginate: <T>(items: T[]) => T[]
  paginationProps: {
    page: number
    rowsPerPage: number
    onPageChange: (event: unknown, page: number) => void
    onRowsPerPageChange: (event: React.ChangeEvent<HTMLInputElement>) => void
  }
}

export function usePagination(options: UsePaginationOptions = {}): UsePaginationReturn {
  const {
    initialPage = 0,
    initialPageSize = 10,
    pageSizeOptions = [5, 10, 25, 50, 100]
  } = options

  const [page, setPage] = useState(initialPage)
  const [pageSize, setPageSize] = useState(initialPageSize)

  const nextPage = useCallback(() => {
    setPage(prev => prev + 1)
  }, [])

  const prevPage = useCallback(() => {
    setPage(prev => Math.max(0, prev - 1))
  }, [])

  const firstPage = useCallback(() => {
    setPage(0)
  }, [])

  const lastPage = useCallback((totalPages: number) => {
    setPage(Math.max(0, totalPages - 1))
  }, [])

  const handlePageChange = useCallback((event: unknown, newPage: number) => {
    setPage(newPage)
  }, [])

  const handlePageSizeChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    const newSize = parseInt(event.target.value, 10)
    setPageSize(newSize)
    setPage(0) // Reset to first page when changing page size
  }, [])

  const paginate = useCallback(<T,>(items: T[]): T[] => {
    const start = page * pageSize
    const end = start + pageSize
    return items.slice(start, end)
  }, [page, pageSize])

  const paginationProps = useMemo(() => ({
    page,
    rowsPerPage: pageSize,
    onPageChange: handlePageChange,
    onRowsPerPageChange: handlePageSizeChange,
  }), [page, pageSize, handlePageChange, handlePageSizeChange])

  return {
    page,
    pageSize,
    pageSizeOptions,
    setPage,
    setPageSize,
    nextPage,
    prevPage,
    firstPage,
    lastPage,
    paginate,
    paginationProps,
  }
}

// Hook for server-side pagination
export function useServerPagination(options: UsePaginationOptions = {}) {
  const pagination = usePagination(options)
  
  const getPaginationParams = useCallback(() => ({
    page: pagination.page + 1, // Convert to 1-based for server
    page_size: pagination.pageSize,
  }), [pagination.page, pagination.pageSize])

  return {
    ...pagination,
    getPaginationParams,
  }
}