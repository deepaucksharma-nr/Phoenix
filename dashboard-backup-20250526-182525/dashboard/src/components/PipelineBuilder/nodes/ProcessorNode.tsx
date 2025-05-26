import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { Paper, Typography, Chip, IconButton, Box } from '@mui/material';
import { Settings, Delete, Warning } from '@mui/icons-material';
import { useAppDispatch, useAppSelector } from '@/store/hooks';
import { setSelectedNode, removeNode } from '@/store/slices/pipelineSlice';

export interface ProcessorNodeData {
  label: string;
  processorType: string;
  config: Record<string, any>;
  status?: 'active' | 'inactive' | 'error';
  metrics?: {
    throughput?: number;
    latency?: number;
    errors?: number;
  };
}

const ProcessorNode = memo(({ id, data, selected }: NodeProps<ProcessorNodeData>) => {
  const dispatch = useAppDispatch();
  const isDirty = useAppSelector((state) => state.pipelines.isDirty);

  const handleClick = () => {
    dispatch(setSelectedNode(id));
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    dispatch(removeNode(id));
  };

  const getStatusColor = () => {
    switch (data.status) {
      case 'active':
        return '#4caf50';
      case 'error':
        return '#f44336';
      default:
        return '#9e9e9e';
    }
  };

  return (
    <Paper
      elevation={selected ? 8 : 2}
      onClick={handleClick}
      sx={{
        padding: 2,
        borderRadius: 2,
        border: selected ? '2px solid #1976d2' : '1px solid #e0e0e0',
        cursor: 'pointer',
        minWidth: 200,
        position: 'relative',
        backgroundColor: selected ? 'rgba(25, 118, 210, 0.04)' : 'white',
        transition: 'all 0.2s ease',
        '&:hover': {
          boxShadow: 4,
          borderColor: '#1976d2',
        },
      }}
    >
      <Handle
        type="target"
        position={Position.Top}
        style={{
          background: '#1976d2',
          width: 12,
          height: 12,
          border: '2px solid white',
          boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
        }}
      />
      
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
        <Box
          sx={{
            width: 8,
            height: 8,
            borderRadius: '50%',
            backgroundColor: getStatusColor(),
            mr: 1,
            flexShrink: 0,
          }}
        />
        <Typography variant="subtitle2" sx={{ fontWeight: 600, flex: 1 }}>
          {data.label}
        </Typography>
        {data.status === 'error' && (
          <Warning fontSize="small" color="error" sx={{ ml: 1 }} />
        )}
      </Box>

      <Chip
        label={data.processorType}
        size="small"
        color="primary"
        variant="outlined"
        sx={{ mb: 1 }}
      />

      {data.metrics && (
        <Box sx={{ mt: 1, fontSize: '0.75rem', color: 'text.secondary' }}>
          {data.metrics.throughput && (
            <Typography variant="caption" display="block">
              Throughput: {data.metrics.throughput}/s
            </Typography>
          )}
          {data.metrics.latency && (
            <Typography variant="caption" display="block">
              Latency: {data.metrics.latency}ms
            </Typography>
          )}
          {data.metrics.errors !== undefined && data.metrics.errors > 0 && (
            <Typography variant="caption" display="block" color="error">
              Errors: {data.metrics.errors}
            </Typography>
          )}
        </Box>
      )}

      {selected && (
        <Box sx={{ position: 'absolute', top: 4, right: 4 }}>
          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation();
              // TODO: Open configuration dialog
            }}
            sx={{ mr: 0.5 }}
          >
            <Settings fontSize="small" />
          </IconButton>
          <IconButton
            size="small"
            onClick={handleDelete}
            color="error"
          >
            <Delete fontSize="small" />
          </IconButton>
        </Box>
      )}

      <Handle
        type="source"
        position={Position.Bottom}
        style={{
          background: '#1976d2',
          width: 12,
          height: 12,
          border: '2px solid white',
          boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
        }}
      />
    </Paper>
  );
});

ProcessorNode.displayName = 'ProcessorNode';

export default ProcessorNode;