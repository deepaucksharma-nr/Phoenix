import React, { useState, useCallback, useRef } from 'react';
import { Box, Paper, Typography, IconButton, Button, Stack, Chip } from '@mui/material';
import { Delete, PlayArrow, Save } from '@mui/icons-material';
import { motion, AnimatePresence } from 'framer-motion';
import { ProcessorBlock } from './ProcessorBlock';
import { ConnectionLine } from './ConnectionLine';
import { usePipelineStore } from '../../store/pipelineStore';

interface Processor {
  id: string;
  type: string;
  name: string;
  config: any;
  position: { x: number; y: number };
  inputs: string[];
  outputs: string[];
}

interface Connection {
  id: string;
  source: string;
  target: string;
  sourcePort: string;
  targetPort: string;
}

export const DragDropCanvas: React.FC = () => {
  const canvasRef = useRef<HTMLDivElement>(null);
  const [processors, setProcessors] = useState<Processor[]>([]);
  const [connections, setConnections] = useState<Connection[]>([]);
  const [selectedProcessor, setSelectedProcessor] = useState<string | null>(null);
  const [draggedProcessor, setDraggedProcessor] = useState<string | null>(null);
  const [connectionStart, setConnectionStart] = useState<{ processorId: string; port: string } | null>(null);
  const [impact, setImpact] = useState<any>(null);
  
  const { savePipeline } = usePipelineStore();
  
  // Handle processor drag
  const handleProcessorDrag = useCallback((processorId: string, newPosition: { x: number; y: number }) => {
    setProcessors(prev => prev.map(p => 
      p.id === processorId ? { ...p, position: newPosition } : p
    ));
  }, []);
  
  // Handle processor drop from palette
  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    const processorType = e.dataTransfer.getData('processorType');
    if (!processorType || !canvasRef.current) return;
    
    const rect = canvasRef.current.getBoundingClientRect();
    const position = {
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
    };
    
    const newProcessor: Processor = {
      id: `processor-${Date.now()}`,
      type: processorType,
      name: `${processorType} ${processors.length + 1}`,
      config: getDefaultConfig(processorType),
      position,
      inputs: ['input'],
      outputs: ['output'],
    };
    
    setProcessors(prev => [...prev, newProcessor]);
    calculateImpact([...processors, newProcessor], connections);
  }, [processors, connections]);
  
  // Handle connection creation
  const handlePortClick = useCallback((processorId: string, port: string, isOutput: boolean) => {
    if (!connectionStart) {
      // Start connection
      if (isOutput) {
        setConnectionStart({ processorId, port });
      }
    } else {
      // Complete connection
      if (!isOutput && connectionStart.processorId !== processorId) {
        const newConnection: Connection = {
          id: `conn-${Date.now()}`,
          source: connectionStart.processorId,
          target: processorId,
          sourcePort: connectionStart.port,
          targetPort: port,
        };
        setConnections(prev => [...prev, newConnection]);
        calculateImpact(processors, [...connections, newConnection]);
      }
      setConnectionStart(null);
    }
  }, [connectionStart, processors, connections]);
  
  // Delete processor
  const handleDeleteProcessor = useCallback((processorId: string) => {
    setProcessors(prev => prev.filter(p => p.id !== processorId));
    setConnections(prev => prev.filter(c => c.source !== processorId && c.target !== processorId));
    setSelectedProcessor(null);
  }, []);
  
  // Calculate pipeline impact
  const calculateImpact = async (procs: Processor[], conns: Connection[]) => {
    try {
      const response = await fetch('/api/v1/pipelines/preview', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          pipeline_config: {
            processors: procs.map(p => ({ type: p.type, config: p.config })),
            connections: conns,
          },
          target_hosts: ['preview'],
        }),
      });
      
      if (response.ok) {
        const data = await response.json();
        setImpact(data);
      }
    } catch (error) {
      console.error('Failed to calculate impact:', error);
    }
  };
  
  // Get default config for processor type
  const getDefaultConfig = (type: string): any => {
    switch (type) {
      case 'filter':
        return { patterns: [], mode: 'include' };
      case 'sample':
        return { rate: 0.1 };
      case 'aggregate':
        return { groupBy: [], operations: ['sum'] };
      case 'transform':
        return { rules: [] };
      case 'top_k':
        return { k: 20, metric: 'value' };
      default:
        return {};
    }
  };
  
  // Save pipeline
  const handleSave = () => {
    const pipelineConfig = {
      name: 'Custom Pipeline',
      processors,
      connections,
      metadata: {
        createdAt: new Date().toISOString(),
        impact,
      },
    };
    
    savePipeline(pipelineConfig);
  };
  
  return (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Toolbar */}
      <Paper sx={{ p: 2, mb: 2 }}>
        <Stack direction="row" justifyContent="space-between" alignItems="center">
          <Typography variant="h6">Pipeline Builder</Typography>
          <Stack direction="row" spacing={2} alignItems="center">
            {impact && (
              <>
                <Chip
                  label={`Cost: -${impact.estimated_cost_reduction}%`}
                  color="success"
                  size="small"
                />
                <Chip
                  label={`CPU: +${impact.estimated_cpu_impact}%`}
                  color="warning"
                  size="small"
                />
              </>
            )}
            <Button
              startIcon={<PlayArrow />}
              variant="outlined"
              size="small"
              onClick={() => calculateImpact(processors, connections)}
            >
              Preview
            </Button>
            <Button
              startIcon={<Save />}
              variant="contained"
              size="small"
              onClick={handleSave}
              disabled={processors.length === 0}
            >
              Save Pipeline
            </Button>
          </Stack>
        </Stack>
      </Paper>
      
      {/* Canvas */}
      <Paper
        ref={canvasRef}
        sx={{
          flex: 1,
          position: 'relative',
          overflow: 'hidden',
          backgroundColor: '#f8f9fa',
          backgroundImage: 'radial-gradient(circle, #e0e0e0 1px, transparent 1px)',
          backgroundSize: '20px 20px',
        }}
        onDrop={handleDrop}
        onDragOver={(e) => e.preventDefault()}
      >
        {/* Connection lines */}
        <svg
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            pointerEvents: 'none',
          }}
        >
          {connections.map(conn => {
            const source = processors.find(p => p.id === conn.source);
            const target = processors.find(p => p.id === conn.target);
            if (!source || !target) return null;
            
            return (
              <ConnectionLine
                key={conn.id}
                start={{ x: source.position.x + 150, y: source.position.y + 40 }}
                end={{ x: target.position.x, y: target.position.y + 40 }}
              />
            );
          })}
        </svg>
        
        {/* Processors */}
        <AnimatePresence>
          {processors.map(processor => (
            <ProcessorBlock
              key={processor.id}
              processor={processor}
              isSelected={selectedProcessor === processor.id}
              onSelect={() => setSelectedProcessor(processor.id)}
              onDelete={() => handleDeleteProcessor(processor.id)}
              onDrag={handleProcessorDrag}
              onPortClick={handlePortClick}
            />
          ))}
        </AnimatePresence>
        
        {/* Drop hint */}
        {processors.length === 0 && (
          <Box
            sx={{
              position: 'absolute',
              top: '50%',
              left: '50%',
              transform: 'translate(-50%, -50%)',
              textAlign: 'center',
            }}
          >
            <Typography variant="h6" color="textSecondary">
              Drag processors here to build your pipeline
            </Typography>
            <Typography variant="body2" color="textSecondary">
              Connect processors by clicking output ports then input ports
            </Typography>
          </Box>
        )}
      </Paper>
      
      {/* Impact preview */}
      {impact && (
        <Paper sx={{ p: 2, mt: 2 }}>
          <Typography variant="subtitle2" gutterBottom>
            Pipeline Impact Preview
          </Typography>
          <Stack direction="row" spacing={4}>
            <Box>
              <Typography variant="caption" color="textSecondary">Cost Reduction</Typography>
              <Typography variant="h6" color="success.main">
                {impact.estimated_cost_reduction}%
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="textSecondary">Cardinality Reduction</Typography>
              <Typography variant="h6" color="primary">
                {impact.estimated_cardinality_reduction}%
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="textSecondary">CPU Impact</Typography>
              <Typography variant="h6" color="warning.main">
                +{impact.estimated_cpu_impact}%
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="textSecondary">Memory Impact</Typography>
              <Typography variant="h6">
                +{impact.estimated_memory_impact}MB
              </Typography>
            </Box>
          </Stack>
        </Paper>
      )}
    </Box>
  );
};