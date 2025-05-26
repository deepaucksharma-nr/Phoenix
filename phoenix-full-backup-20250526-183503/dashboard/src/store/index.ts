import { configureStore } from '@reduxjs/toolkit'

// Simple store without complex slices to avoid circular dependencies
export const store = configureStore({
  reducer: {
    app: (state = { initialized: true }) => state,
  },
  devTools: import.meta.env.DEV,
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch

export default store