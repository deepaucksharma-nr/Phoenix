import { configureStore } from '@reduxjs/toolkit';
import authSlice from './slices/authSlice';
import experimentSlice from './slices/experimentSlice';
import pipelineSlice from './slices/pipelineSlice';
import notificationSlice from './slices/notificationSlice';
import uiSlice from './slices/uiSlice';

export const store = configureStore({
  reducer: {
    auth: authSlice,
    experiments: experimentSlice,
    pipeline: pipelineSlice,
    notifications: notificationSlice,
    ui: uiSlice,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // Ignore these action types
        ignoredActions: ['auth/setCredentials', 'notifications/addNotification'],
        // Ignore these field paths in all actions
        ignoredActionPaths: ['meta.arg', 'payload.timestamp'],
        // Ignore these paths in the state
        ignoredPaths: ['auth.user', 'notifications.items'],
      },
    }),
  devTools: import.meta.env.DEV,
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

export default store;