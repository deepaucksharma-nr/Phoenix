import { configureStore } from '@reduxjs/toolkit'

// Simple initial store
const experimentsSlice = {
  name: 'experiments',
  initialState: {
    experiments: [
      {
        id: '1',
        name: 'Process Metrics Optimization',
        description: 'Testing cardinality reduction strategies',
        status: 'running',
        createdAt: '2024-03-20T10:00:00Z',
        spec: {
          duration: '2h',
          targetHosts: ['host-1', 'host-2'],
        },
      },
      {
        id: '2', 
        name: 'Memory Usage Optimization',
        description: 'Reducing memory footprint of collectors',
        status: 'completed',
        createdAt: '2024-03-19T14:30:00Z',
        spec: {
          duration: '4h',
          targetHosts: ['host-3', 'host-4'],
        },
      },
    ],
    loading: false,
    error: null,
  },
  reducers: {
    setLoading: (state: any, action: any) => {
      state.loading = action.payload
    },
  },
}

export const store = configureStore({
  reducer: {
    experiments: (state = experimentsSlice.initialState, action) => {
      switch (action.type) {
        case 'experiments/setLoading':
          return { ...state, loading: action.payload }
        default:
          return state
      }
    },
    pipelines: (state = { deployments: [], loading: false }, action) => state,
  },
  devTools: import.meta.env.DEV,
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch

export default store