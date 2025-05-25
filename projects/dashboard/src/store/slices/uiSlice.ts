import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface UIState {
  sidebarOpen: boolean;
  theme: 'light' | 'dark';
  experimentWizardOpen: boolean;
  pipelineBuilderFullscreen: boolean;
  loadingStates: Record<string, boolean>;
  modals: {
    confirmDialog: {
      open: boolean;
      title: string;
      message: string;
      onConfirm: (() => void) | null;
      onCancel: (() => void) | null;
    };
  };
}

const initialState: UIState = {
  sidebarOpen: true,
  theme: (localStorage.getItem('theme') as 'light' | 'dark') || 'light',
  experimentWizardOpen: false,
  pipelineBuilderFullscreen: false,
  loadingStates: {},
  modals: {
    confirmDialog: {
      open: false,
      title: '',
      message: '',
      onConfirm: null,
      onCancel: null,
    },
  },
};

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    toggleSidebar: (state) => {
      state.sidebarOpen = !state.sidebarOpen;
    },
    setSidebarOpen: (state, action: PayloadAction<boolean>) => {
      state.sidebarOpen = action.payload;
    },
    toggleTheme: (state) => {
      state.theme = state.theme === 'light' ? 'dark' : 'light';
      localStorage.setItem('theme', state.theme);
    },
    setTheme: (state, action: PayloadAction<'light' | 'dark'>) => {
      state.theme = action.payload;
      localStorage.setItem('theme', action.payload);
    },
    setExperimentWizardOpen: (state, action: PayloadAction<boolean>) => {
      state.experimentWizardOpen = action.payload;
    },
    setPipelineBuilderFullscreen: (state, action: PayloadAction<boolean>) => {
      state.pipelineBuilderFullscreen = action.payload;
    },
    setLoadingState: (
      state,
      action: PayloadAction<{ key: string; loading: boolean }>
    ) => {
      const { key, loading } = action.payload;
      if (loading) {
        state.loadingStates[key] = true;
      } else {
        delete state.loadingStates[key];
      }
    },
    showConfirmDialog: (
      state,
      action: PayloadAction<{
        title: string;
        message: string;
        onConfirm: () => void;
        onCancel?: () => void;
      }>
    ) => {
      state.modals.confirmDialog = {
        open: true,
        title: action.payload.title,
        message: action.payload.message,
        onConfirm: action.payload.onConfirm,
        onCancel: action.payload.onCancel || null,
      };
    },
    hideConfirmDialog: (state) => {
      state.modals.confirmDialog = {
        open: false,
        title: '',
        message: '',
        onConfirm: null,
        onCancel: null,
      };
    },
  },
});

export const {
  toggleSidebar,
  setSidebarOpen,
  toggleTheme,
  setTheme,
  setExperimentWizardOpen,
  setPipelineBuilderFullscreen,
  setLoadingState,
  showConfirmDialog,
  hideConfirmDialog,
} = uiSlice.actions;

export default uiSlice.reducer;