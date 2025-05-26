import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

interface Pipeline {
  id: string;
  name: string;
  processors: any[];
  connections: any[];
  metadata?: {
    createdAt: string;
    impact?: any;
  };
}

interface PipelineStore {
  pipelines: Pipeline[];
  selectedPipeline: Pipeline | null;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  fetchPipelines: () => Promise<void>;
  selectPipeline: (pipeline: Pipeline | null) => void;
  savePipeline: (pipeline: any) => Promise<void>;
  deletePipeline: (id: string) => Promise<void>;
  deployPipeline: (pipelineId: string, hosts: string[]) => Promise<void>;
}

export const usePipelineStore = create<PipelineStore>()(
  devtools(
    (set, get) => ({
      pipelines: [],
      selectedPipeline: null,
      isLoading: false,
      error: null,
      
      fetchPipelines: async () => {
        set({ isLoading: true, error: null });
        try {
          // Fetch saved pipelines
          const response = await fetch('/api/v1/pipelines');
          if (!response.ok) throw new Error('Failed to fetch pipelines');
          
          const data = await response.json();
          set({ pipelines: data, isLoading: false });
        } catch (error) {
          set({ error: (error as Error).message, isLoading: false });
        }
      },
      
      selectPipeline: (pipeline) => {
        set({ selectedPipeline: pipeline });
      },
      
      savePipeline: async (pipelineData) => {
        set({ isLoading: true, error: null });
        try {
          const pipeline: Pipeline = {
            id: `pipeline-${Date.now()}`,
            ...pipelineData,
          };
          
          // In a real app, this would save to the backend
          set(state => ({
            pipelines: [...state.pipelines, pipeline],
            selectedPipeline: pipeline,
            isLoading: false,
          }));
        } catch (error) {
          set({ error: (error as Error).message, isLoading: false });
        }
      },
      
      deletePipeline: async (id) => {
        set({ isLoading: true, error: null });
        try {
          // In a real app, this would delete from the backend
          set(state => ({
            pipelines: state.pipelines.filter(p => p.id !== id),
            selectedPipeline: state.selectedPipeline?.id === id ? null : state.selectedPipeline,
            isLoading: false,
          }));
        } catch (error) {
          set({ error: (error as Error).message, isLoading: false });
        }
      },
      
      deployPipeline: async (pipelineId, hosts) => {
        set({ isLoading: true, error: null });
        try {
          const pipeline = get().pipelines.find(p => p.id === pipelineId);
          if (!pipeline) throw new Error('Pipeline not found');
          
          const response = await fetch('/api/v1/pipelines/quick-deploy', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              pipeline_config: {
                processors: pipeline.processors,
                connections: pipeline.connections,
              },
              target_hosts: hosts,
            }),
          });
          
          if (!response.ok) throw new Error('Failed to deploy pipeline');
          
          set({ isLoading: false });
        } catch (error) {
          set({ error: (error as Error).message, isLoading: false });
        }
      },
    }),
    {
      name: 'pipeline-store',
    }
  )
);