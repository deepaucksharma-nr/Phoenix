import React from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  TableSortLabel,
  Paper,
  CircularProgress,
  Box,
  Typography,
  Alert,
  Checkbox,
} from '@mui/material'

export interface Column<T> {
  id: keyof T | string
  label: string
  minWidth?: number
  align?: 'left' | 'right' | 'center'
  format?: (value: any, row: T) => React.ReactNode
  sortable?: boolean
}

interface DataTableProps<T> {
  columns: Column<T>[]
  data: T[]
  loading?: boolean
  error?: Error | null
  emptyMessage?: string
  rowKey: keyof T
  onRowClick?: (row: T) => void
  selectable?: boolean
  selected?: T[]
  onSelectionChange?: (selected: T[]) => void
  pagination?: {
    count: number
    page: number
    rowsPerPage: number
    onPageChange: (event: unknown, page: number) => void
    onRowsPerPageChange: (event: React.ChangeEvent<HTMLInputElement>) => void
  }
  sortable?: boolean
  orderBy?: string
  order?: 'asc' | 'desc'
  onSort?: (property: string) => void
}

export function DataTable<T extends Record<string, any>>({
  columns,
  data,
  loading = false,
  error = null,
  emptyMessage = 'No data available',
  rowKey,
  onRowClick,
  selectable = false,
  selected = [],
  onSelectionChange,
  pagination,
  sortable = false,
  orderBy = '',
  order = 'asc',
  onSort,
}: DataTableProps<T>) {
  const isSelected = (row: T) => {
    return selected.some(item => item[rowKey] === row[rowKey])
  }

  const handleSelectAll = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      onSelectionChange?.(data)
    } else {
      onSelectionChange?.([])
    }
  }

  const handleSelectRow = (row: T) => {
    const selectedIndex = selected.findIndex(item => item[rowKey] === row[rowKey])
    let newSelected: T[] = []

    if (selectedIndex === -1) {
      newSelected = [...selected, row]
    } else {
      newSelected = selected.filter(item => item[rowKey] !== row[rowKey])
    }

    onSelectionChange?.(newSelected)
  }

  const createSortHandler = (property: string) => () => {
    onSort?.(property)
  }

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight={200}>
        <CircularProgress />
      </Box>
    )
  }

  if (error) {
    return (
      <Alert severity="error" sx={{ m: 2 }}>
        {error.message}
      </Alert>
    )
  }

  if (data.length === 0) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight={200}>
        <Typography color="textSecondary">{emptyMessage}</Typography>
      </Box>
    )
  }

  return (
    <Paper>
      <TableContainer>
        <Table stickyHeader>
          <TableHead>
            <TableRow>
              {selectable && (
                <TableCell padding="checkbox">
                  <Checkbox
                    indeterminate={selected.length > 0 && selected.length < data.length}
                    checked={data.length > 0 && selected.length === data.length}
                    onChange={handleSelectAll}
                  />
                </TableCell>
              )}
              {columns.map((column) => (
                <TableCell
                  key={column.id.toString()}
                  align={column.align}
                  style={{ minWidth: column.minWidth }}
                >
                  {sortable && column.sortable !== false ? (
                    <TableSortLabel
                      active={orderBy === column.id}
                      direction={orderBy === column.id ? order : 'asc'}
                      onClick={createSortHandler(column.id.toString())}
                    >
                      {column.label}
                    </TableSortLabel>
                  ) : (
                    column.label
                  )}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {data.map((row) => {
              const isItemSelected = isSelected(row)
              return (
                <TableRow
                  hover
                  onClick={() => onRowClick?.(row)}
                  role={onRowClick ? 'button' : undefined}
                  tabIndex={-1}
                  key={row[rowKey]}
                  selected={isItemSelected}
                  sx={{ cursor: onRowClick ? 'pointer' : 'default' }}
                >
                  {selectable && (
                    <TableCell padding="checkbox">
                      <Checkbox
                        checked={isItemSelected}
                        onClick={(event) => {
                          event.stopPropagation()
                          handleSelectRow(row)
                        }}
                      />
                    </TableCell>
                  )}
                  {columns.map((column) => {
                    const value = column.id.includes('.')
                      ? column.id.split('.').reduce((obj, key) => obj?.[key], row)
                      : row[column.id]
                    return (
                      <TableCell key={column.id.toString()} align={column.align}>
                        {column.format ? column.format(value, row) : value}
                      </TableCell>
                    )
                  })}
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </TableContainer>
      {pagination && (
        <TablePagination
          component="div"
          count={pagination.count}
          page={pagination.page}
          onPageChange={pagination.onPageChange}
          rowsPerPage={pagination.rowsPerPage}
          onRowsPerPageChange={pagination.onRowsPerPageChange}
        />
      )}
    </Paper>
  )
}