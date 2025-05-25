import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit';
import { Experiment, ExperimentMetrics } from '@/types/experiment';
import axiosInstance from '@/config/axios';

interface ExperimentState {
  experiments: Experiment[];
  currentExperiment: Experiment | null;
  metrics: Record<string, ExperimentMetrics>;
  loading: boolean;
  error: string | null;
  filters: {
    status: string[];
    dateRange: [Date | null, Date | null];
    search: string;
  };
}

// Async thunks
export const fetchExperiments = createAsyncThunk(
  'experiments/fetchAll',
  async () => {
    const response = await axiosInstance.get('/api/experiments');
    return response.data;
  }
);

export const fetchExperimentById = createAsyncThunk(
  'experiments/fetchById',
  async (id: string) => {
    const response = await axiosInstance.get(`/api/experiments/${id}`);
    return response.data;
  }
);

export const createExperiment = createAsyncThunk(
  'experiments/create',
  async (data: Partial<Experiment>) => {
    const response = await axiosInstance.post('/api/experiments', data);
    return response.data;
  }
);

export const updateExperimentStatus = createAsyncThunk(
  'experiments/updateStatus',
  async ({ id, status }: { id: string; status: string }) => {
    const response = await axiosInstance.patch(`/api/experiments/${id}/status`, { status });
    return response.data;
  }
);

export const deleteExperiment = createAsyncThunk(
  'experiments/delete',
  async (id: string) => {
    await axiosInstance.delete(`/api/experiments/${id}`);
    return id;
  }
);

const initialState: ExperimentState = {
  experiments: [],
  currentExperiment: null,
  metrics: {},
  loading: false,
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
      state.loading = false;
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
      state.loading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
      state.loading = false;
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
  extraReducers: (builder) => {
    builder
      // Fetch experiments
      .addCase(fetchExperiments.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchExperiments.fulfilled, (state, action) => {
        state.loading = false;
        state.experiments = action.payload;
      })
      .addCase(fetchExperiments.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch experiments';
      })
      // Fetch experiment by ID
      .addCase(fetchExperimentById.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchExperimentById.fulfilled, (state, action) => {
        state.loading = false;
        state.currentExperiment = action.payload;
      })
      .addCase(fetchExperimentById.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch experiment';
      })
      // Create experiment
      .addCase(createExperiment.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(createExperiment.fulfilled, (state, action) => {
        state.loading = false;
        state.experiments.push(action.payload);
      })
      .addCase(createExperiment.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to create experiment';
      })
      // Update experiment status
      .addCase(updateExperimentStatus.fulfilled, (state, action) => {
        const index = state.experiments.findIndex(exp => exp.id === action.payload.id);
        if (index !== -1) {
          state.experiments[index] = action.payload;
        }
        if (state.currentExperiment?.id === action.payload.id) {
          state.currentExperiment = action.payload;
        }
      })
      // Delete experiment
      .addCase(deleteExperiment.fulfilled, (state, action) => {
        state.experiments = state.experiments.filter(exp => exp.id !== action.payload);
        if (state.currentExperiment?.id === action.payload) {
          state.currentExperiment = null;
        }
      });
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