import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export type NotificationType = 'success' | 'error' | 'warning' | 'info';

export interface Notification {
  id: string;
  type: NotificationType;
  message: string;
  description?: string;
  duration?: number;
  timestamp: number;
  action?: {
    label: string;
    handler: () => void;
  };
}

interface NotificationState {
  notifications: Notification[];
  maxNotifications: number;
}

const initialState: NotificationState = {
  notifications: [],
  maxNotifications: 5,
};

const notificationSlice = createSlice({
  name: 'notifications',
  initialState,
  reducers: {
    addNotification: (state, action: PayloadAction<Omit<Notification, 'id' | 'timestamp'>>) => {
      const notification: Notification = {
        ...action.payload,
        id: `${Date.now()}-${Math.random()}`,
        timestamp: Date.now(),
        duration: action.payload.duration || 5000,
      };
      
      state.notifications.unshift(notification);
      
      // Keep only the latest notifications
      if (state.notifications.length > state.maxNotifications) {
        state.notifications = state.notifications.slice(0, state.maxNotifications);
      }
    },
    removeNotification: (state, action: PayloadAction<string>) => {
      state.notifications = state.notifications.filter(
        (notif) => notif.id !== action.payload
      );
    },
    clearNotifications: (state) => {
      state.notifications = [];
    },
    clearOldNotifications: (state) => {
      const now = Date.now();
      state.notifications = state.notifications.filter(
        (notif) => {
          if (!notif.duration) return true;
          return now - notif.timestamp < notif.duration;
        }
      );
    },
  },
});

export const {
  addNotification,
  removeNotification,
  clearNotifications,
  clearOldNotifications,
} = notificationSlice.actions;

export default notificationSlice.reducer;

// Helper functions for creating notifications
export const createSuccessNotification = (
  message: string,
  description?: string,
  duration?: number
): Omit<Notification, 'id' | 'timestamp'> => ({
  type: 'success',
  message,
  description,
  duration,
});

export const createErrorNotification = (
  message: string,
  description?: string,
  duration?: number
): Omit<Notification, 'id' | 'timestamp'> => ({
  type: 'error',
  message,
  description,
  duration: duration || 10000, // Errors stay longer
});

export const createWarningNotification = (
  message: string,
  description?: string,
  duration?: number
): Omit<Notification, 'id' | 'timestamp'> => ({
  type: 'warning',
  message,
  description,
  duration,
});

export const createInfoNotification = (
  message: string,
  description?: string,
  duration?: number
): Omit<Notification, 'id' | 'timestamp'> => ({
  type: 'info',
  message,
  description,
  duration,
});