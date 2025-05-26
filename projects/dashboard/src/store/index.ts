import { configureStore } from '@reduxjs/toolkit'
import experimentReducer from './slices/experimentSlice'
import pipelineReducer from './slices/pipelineSlice'

export const store = configureStore({
  reducer: {
    experiments: experimentReducer,
    pipelines: pipelineReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // Ignore these action types for async thunks
        ignoredActions: [
          'experiments/fetchExperiments/pending',
          'experiments/fetchExperiments/fulfilled',
          'experiments/fetchExperiments/rejected',
          'pipelines/fetchDeployments/pending',
          'pipelines/fetchDeployments/fulfilled',
          'pipelines/fetchDeployments/rejected',
        ],
      },
    }),
  devTools: import.meta.env.DEV,
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch

export default store