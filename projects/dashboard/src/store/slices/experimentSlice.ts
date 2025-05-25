import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Experiment, ExperimentMetrics } from '@types/experiment';

interface ExperimentState {
  experiments: Experiment[];
  currentExperiment: Experiment | null;
  metrics: Record<string, ExperimentMetrics>;
  isLoading: boolean;
  error: string | null;
  filters: {
    status: string[];
    dateRange: [Date | null, Date | null];
    search: string;
  };
}

const initialState: ExperimentState = {
  experiments: [],
  currentExperiment: null,
  metrics: {},
  isLoading: false,
  error: null,
  filters: {
    status: [],
    dateRange: [null, null],
    search: '',
  },
};

const experimentSlice = createSlice({
  name: 'experiments',
  initialState,
  reducers: {
    setExperiments: (state, action: PayloadAction<Experiment[]>) => {
      state.experiments = action.payload;
      state.isLoading = false;
      state.error = null;
    },
    addExperiment: (state, action: PayloadAction<Experiment>) => {
      state.experiments.push(action.payload);
    },
    updateExperiment: (state, action: PayloadAction<Experiment>) => {
      const index = state.experiments.findIndex(
        (exp) => exp.id === action.payload.id
      );
      if (index !== -1) {
        state.experiments[index] = action.payload;
      }
      if (state.currentExperiment?.id === action.payload.id) {
        state.currentExperiment = action.payload;
      }
    },
    removeExperiment: (state, action: PayloadAction<string>) => {
      state.experiments = state.experiments.filter(
        (exp) => exp.id !== action.payload
      );
      if (state.currentExperiment?.id === action.payload) {
        state.currentExperiment = null;
      }
    },
    setCurrentExperiment: (state, action: PayloadAction<Experiment | null>) => {
      state.currentExperiment = action.payload;
    },
    setExperimentMetrics: (
      state,
      action: PayloadAction<{ experimentId: string; metrics: ExperimentMetrics }>
    ) => {
      const { experimentId, metrics } = action.payload;
      state.metrics[experimentId] = metrics;
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
      state.isLoading = false;
    },
    setFilters: (
      state,
      action: PayloadAction<Partial<ExperimentState['filters']>>
    ) => {
      state.filters = { ...state.filters, ...action.payload };
    },
    clearFilters: (state) => {
      state.filters = initialState.filters;
    },
  },
});

export const {
  setExperiments,
  addExperiment,
  updateExperiment,
  removeExperiment,
  setCurrentExperiment,
  setExperimentMetrics,
  setLoading,
  setError,
  setFilters,
  clearFilters,
} = experimentSlice.actions;

export default experimentSlice.reducer;