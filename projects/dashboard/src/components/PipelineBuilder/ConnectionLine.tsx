import React from 'react';
import { motion } from 'framer-motion';

interface ConnectionLineProps {
  start: { x: number; y: number };
  end: { x: number; y: number };
  color?: string;
  animated?: boolean;
}

export const ConnectionLine: React.FC<ConnectionLineProps> = ({
  start,
  end,
  color = '#2196F3',
  animated = true,
}) => {
  // Calculate bezier curve control points
  const midX = (start.x + end.x) / 2;
  const controlPoint1 = { x: midX, y: start.y };
  const controlPoint2 = { x: midX, y: end.y };
  
  const pathData = `
    M ${start.x} ${start.y}
    C ${controlPoint1.x} ${controlPoint1.y},
      ${controlPoint2.x} ${controlPoint2.y},
      ${end.x} ${end.y}
  `;
  
  return (
    <g>
      {/* Connection line */}
      <motion.path
        d={pathData}
        fill="none"
        stroke={color}
        strokeWidth={2}
        initial={{ pathLength: 0, opacity: 0 }}
        animate={{ pathLength: 1, opacity: 1 }}
        transition={{ duration: 0.5, ease: 'easeOut' }}
      />
      
      {/* Animated flow particles */}
      {animated && (
        <>
          <circle r="3" fill={color}>
            <animateMotion
              dur="2s"
              repeatCount="indefinite"
              path={pathData}
            />
          </circle>
          <circle r="3" fill={color}>
            <animateMotion
              dur="2s"
              repeatCount="indefinite"
              path={pathData}
              begin="0.5s"
            />
          </circle>
          <circle r="3" fill={color}>
            <animateMotion
              dur="2s"
              repeatCount="indefinite"
              path={pathData}
              begin="1s"
            />
          </circle>
        </>
      )}
      
      {/* Arrowhead */}
      <defs>
        <marker
          id={`arrowhead-${start.x}-${start.y}-${end.x}-${end.y}`}
          markerWidth="10"
          markerHeight="10"
          refX="9"
          refY="3"
          orient="auto"
        >
          <polygon
            points="0 0, 10 3, 0 6"
            fill={color}
          />
        </marker>
      </defs>
      
      <path
        d={pathData}
        fill="none"
        stroke={color}
        strokeWidth={2}
        markerEnd={`url(#arrowhead-${start.x}-${start.y}-${end.x}-${end.y})`}
        opacity={0.3}
      />
    </g>
  );
};