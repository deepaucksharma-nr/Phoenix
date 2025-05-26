import { configureStore } from '@reduxjs/toolkit'
import authReducer from './slices/authSlice'
import experimentReducer from './slices/experimentSlice'
import pipelineReducer from './slices/pipelineSlice'
import notificationReducer from './slices/notificationSlice'
import uiReducer from './slices/uiSlice'

export const store = configureStore({
  reducer: {
    auth: authReducer,
    experiments: experimentReducer,
    pipeline: pipelineReducer,
    notifications: notificationReducer,
    ui: uiReducer,
  },
  devTools: import.meta.env.DEV,
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch

export default store