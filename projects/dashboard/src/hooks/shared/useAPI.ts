import { useState, useEffect, useCallback } from 'react'
import axios, { AxiosError, AxiosRequestConfig } from 'axios'
import { useNotification } from '../useNotification'

interface UseAPIOptions<T> extends AxiosRequestConfig {
  initialData?: T
  onSuccess?: (data: T) => void
  onError?: (error: Error) => void
  autoFetch?: boolean
  deps?: any[]
}

interface UseAPIReturn<T> {
  data: T | undefined
  loading: boolean
  error: Error | null
  refetch: () => Promise<void>
  mutate: (newData: T) => void
}

export function useAPI<T = any>(
  url: string,
  options: UseAPIOptions<T> = {}
): UseAPIReturn<T> {
  const {
    initialData,
    onSuccess,
    onError,
    autoFetch = true,
    deps = [],
    ...axiosConfig
  } = options

  const [data, setData] = useState<T | undefined>(initialData)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const { showError } = useNotification()

  const fetchData = useCallback(async () => {
    if (!url) return

    setLoading(true)
    setError(null)

    try {
      const response = await axios({
        url,
        ...axiosConfig,
      })

      setData(response.data)
      onSuccess?.(response.data)
    } catch (err) {
      const error = err as AxiosError
      const errorMessage = error.response?.data?.error || error.message
      
      setError(new Error(errorMessage))
      showError(`Failed to fetch data: ${errorMessage}`)
      onError?.(new Error(errorMessage))
    } finally {
      setLoading(false)
    }
  }, [url, ...deps])

  useEffect(() => {
    if (autoFetch) {
      fetchData()
    }
  }, [fetchData, autoFetch])

  const mutate = useCallback((newData: T) => {
    setData(newData)
  }, [])

  return {
    data,
    loading,
    error,
    refetch: fetchData,
    mutate,
  }
}

// Hook for POST/PUT/DELETE operations
export function useMutation<TData = any, TVariables = any>(
  url: string | ((variables: TVariables) => string),
  options: {
    method?: 'POST' | 'PUT' | 'PATCH' | 'DELETE'
    onSuccess?: (data: TData, variables: TVariables) => void
    onError?: (error: Error, variables: TVariables) => void
  } = {}
) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const { showError, showSuccess } = useNotification()

  const mutate = useCallback(
    async (variables: TVariables) => {
      setLoading(true)
      setError(null)

      try {
        const finalUrl = typeof url === 'function' ? url(variables) : url
        const response = await axios({
          url: finalUrl,
          method: options.method || 'POST',
          data: variables,
        })

        options.onSuccess?.(response.data, variables)
        return response.data
      } catch (err) {
        const error = err as AxiosError
        const errorMessage = error.response?.data?.error || error.message
        
        setError(new Error(errorMessage))
        showError(`Operation failed: ${errorMessage}`)
        options.onError?.(new Error(errorMessage), variables)
        throw error
      } finally {
        setLoading(false)
      }
    },
    [url, options.method, options.onSuccess, options.onError, showError]
  )

  return {
    mutate,
    loading,
    error,
  }
}