import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Pipeline, PipelineNode, PipelineConnection } from '@types/pipeline';

interface PipelineState {
  pipelines: Pipeline[];
  currentPipeline: Pipeline | null;
  nodes: PipelineNode[];
  connections: PipelineConnection[];
  selectedNodeId: string | null;
  isLoading: boolean;
  error: string | null;
  isDirty: boolean;
}

const initialState: PipelineState = {
  pipelines: [],
  currentPipeline: null,
  nodes: [],
  connections: [],
  selectedNodeId: null,
  isLoading: false,
  error: null,
  isDirty: false,
};

const pipelineSlice = createSlice({
  name: 'pipelines',
  initialState,
  reducers: {
    setPipelines: (state, action: PayloadAction<Pipeline[]>) => {
      state.pipelines = action.payload;
      state.isLoading = false;
      state.error = null;
    },
    setCurrentPipeline: (state, action: PayloadAction<Pipeline | null>) => {
      state.currentPipeline = action.payload;
      state.nodes = action.payload?.nodes || [];
      state.connections = action.payload?.connections || [];
      state.isDirty = false;
    },
    addNode: (state, action: PayloadAction<PipelineNode>) => {
      state.nodes.push(action.payload);
      state.isDirty = true;
    },
    updateNode: (state, action: PayloadAction<PipelineNode>) => {
      const index = state.nodes.findIndex(
        (node) => node.id === action.payload.id
      );
      if (index !== -1) {
        state.nodes[index] = action.payload;
        state.isDirty = true;
      }
    },
    removeNode: (state, action: PayloadAction<string>) => {
      state.nodes = state.nodes.filter((node) => node.id !== action.payload);
      state.connections = state.connections.filter(
        (conn) =>
          conn.source !== action.payload && conn.target !== action.payload
      );
      state.isDirty = true;
    },
    addConnection: (state, action: PayloadAction<PipelineConnection>) => {
      state.connections.push(action.payload);
      state.isDirty = true;
    },
    removeConnection: (state, action: PayloadAction<string>) => {
      state.connections = state.connections.filter(
        (conn) => conn.id !== action.payload
      );
      state.isDirty = true;
    },
    setSelectedNode: (state, action: PayloadAction<string | null>) => {
      state.selectedNodeId = action.payload;
    },
    updateNodePosition: (
      state,
      action: PayloadAction<{ id: string; x: number; y: number }>
    ) => {
      const node = state.nodes.find((n) => n.id === action.payload.id);
      if (node) {
        node.position = { x: action.payload.x, y: action.payload.y };
        state.isDirty = true;
      }
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
      state.isLoading = false;
    },
    clearPipeline: (state) => {
      state.currentPipeline = null;
      state.nodes = [];
      state.connections = [];
      state.selectedNodeId = null;
      state.isDirty = false;
    },
    setDirty: (state, action: PayloadAction<boolean>) => {
      state.isDirty = action.payload;
    },
  },
});

export const {
  setPipelines,
  setCurrentPipeline,
  addNode,
  updateNode,
  removeNode,
  addConnection,
  removeConnection,
  setSelectedNode,
  updateNodePosition,
  setLoading,
  setError,
  clearPipeline,
  setDirty,
} = pipelineSlice.actions;

export default pipelineSlice.reducer;