import React, { useState } from 'react';
import { Box, Paper, Typography, IconButton, Tooltip } from '@mui/material';
import { Delete, Settings, Input, Output } from '@mui/icons-material';
import { motion } from 'framer-motion';

interface ProcessorBlockProps {
  processor: {
    id: string;
    type: string;
    name: string;
    config: any;
    position: { x: number; y: number };
    inputs: string[];
    outputs: string[];
  };
  isSelected: boolean;
  onSelect: () => void;
  onDelete: () => void;
  onDrag: (id: string, position: { x: number; y: number }) => void;
  onPortClick: (processorId: string, port: string, isOutput: boolean) => void;
}

const processorColors: Record<string, string> = {
  filter: '#4CAF50',
  sample: '#2196F3',
  aggregate: '#FF9800',
  transform: '#9C27B0',
  top_k: '#F44336',
};

export const ProcessorBlock: React.FC<ProcessorBlockProps> = ({
  processor,
  isSelected,
  onSelect,
  onDelete,
  onDrag,
  onPortClick,
}) => {
  const [isDragging, setIsDragging] = useState(false);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  
  const handleMouseDown = (e: React.MouseEvent) => {
    const rect = (e.target as HTMLElement).getBoundingClientRect();
    setDragOffset({
      x: e.clientX - processor.position.x,
      y: e.clientY - processor.position.y,
    });
    setIsDragging(true);
    onSelect();
  };
  
  const handleMouseMove = (e: MouseEvent) => {
    if (isDragging) {
      onDrag(processor.id, {
        x: e.clientX - dragOffset.x,
        y: e.clientY - dragOffset.y,
      });
    }
  };
  
  const handleMouseUp = () => {
    setIsDragging(false);
  };
  
  React.useEffect(() => {
    if (isDragging) {
      window.addEventListener('mousemove', handleMouseMove);
      window.addEventListener('mouseup', handleMouseUp);
      return () => {
        window.removeEventListener('mousemove', handleMouseMove);
        window.removeEventListener('mouseup', handleMouseUp);
      };
    }
  }, [isDragging, dragOffset]);
  
  const color = processorColors[processor.type] || '#757575';
  
  return (
    <motion.div
      initial={{ scale: 0, opacity: 0 }}
      animate={{ scale: 1, opacity: 1 }}
      exit={{ scale: 0, opacity: 0 }}
      style={{
        position: 'absolute',
        left: processor.position.x,
        top: processor.position.y,
        cursor: isDragging ? 'grabbing' : 'grab',
      }}
    >
      <Paper
        elevation={isSelected ? 8 : 3}
        sx={{
          width: 150,
          height: 80,
          bgcolor: 'white',
          border: 2,
          borderColor: isSelected ? color : 'transparent',
          position: 'relative',
          transition: 'all 0.2s',
          '&:hover': {
            boxShadow: 6,
          },
        }}
        onMouseDown={handleMouseDown}
      >
        {/* Header */}
        <Box
          sx={{
            bgcolor: color,
            color: 'white',
            px: 1,
            py: 0.5,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <Typography variant="caption" fontWeight="bold">
            {processor.type.toUpperCase()}
          </Typography>
          <Box>
            <IconButton
              size="small"
              sx={{ color: 'white', p: 0.25 }}
              onClick={(e) => {
                e.stopPropagation();
                // Open config dialog
              }}
            >
              <Settings fontSize="small" />
            </IconButton>
            <IconButton
              size="small"
              sx={{ color: 'white', p: 0.25 }}
              onClick={(e) => {
                e.stopPropagation();
                onDelete();
              }}
            >
              <Delete fontSize="small" />
            </IconButton>
          </Box>
        </Box>
        
        {/* Body */}
        <Box sx={{ p: 1 }}>
          <Typography variant="body2" noWrap>
            {processor.name}
          </Typography>
        </Box>
        
        {/* Input ports */}
        {processor.inputs.map((input, index) => (
          <Tooltip key={input} title={`Input: ${input}`} placement="left">
            <Box
              sx={{
                position: 'absolute',
                left: -8,
                top: 30 + index * 20,
                width: 16,
                height: 16,
                borderRadius: '50%',
                bgcolor: 'white',
                border: 2,
                borderColor: color,
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                '&:hover': {
                  bgcolor: color,
                  color: 'white',
                },
              }}
              onClick={(e) => {
                e.stopPropagation();
                onPortClick(processor.id, input, false);
              }}
            >
              <Input sx={{ fontSize: 10 }} />
            </Box>
          </Tooltip>
        ))}
        
        {/* Output ports */}
        {processor.outputs.map((output, index) => (
          <Tooltip key={output} title={`Output: ${output}`} placement="right">
            <Box
              sx={{
                position: 'absolute',
                right: -8,
                top: 30 + index * 20,
                width: 16,
                height: 16,
                borderRadius: '50%',
                bgcolor: 'white',
                border: 2,
                borderColor: color,
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                '&:hover': {
                  bgcolor: color,
                  color: 'white',
                },
              }}
              onClick={(e) => {
                e.stopPropagation();
                onPortClick(processor.id, output, true);
              }}
            >
              <Output sx={{ fontSize: 10 }} />
            </Box>
          </Tooltip>
        ))}
      </Paper>
    </motion.div>
  );
};