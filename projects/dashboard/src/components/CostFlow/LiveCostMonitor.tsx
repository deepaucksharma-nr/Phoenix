import React, { useEffect, useState, useRef } from 'react';
import { Box, Card, CardHeader, CardContent, Typography, Chip, Button, Stack } from '@mui/material';
import { motion, AnimatePresence } from 'framer-motion';
import * as d3 from 'd3';
import { useWebSocket } from '../../hooks/useWebSocket';
import { formatCurrency } from '../../utils/formatting';

interface MetricFlowData {
  metricName: string;
  costPerMinute: number;
  percentage: number;
  cardinality: number;
}

export const LiveCostMonitor: React.FC = () => {
  const svgRef = useRef<SVGSVGElement>(null);
  const [totalCost, setTotalCost] = useState(0);
  const [metricFlows, setMetricFlows] = useState<MetricFlowData[]>([]);
  const [selectedMetric, setSelectedMetric] = useState<string | null>(null);
  
  const { subscribe } = useWebSocket();
  
  useEffect(() => {
    // Subscribe to metric flow updates
    const unsubscribe = subscribe('metric_flow', (data) => {
      setTotalCost(data.total_cost_rate);
      setMetricFlows(data.top_metrics.slice(0, 10));
    });
    
    return unsubscribe;
  }, [subscribe]);
  
  useEffect(() => {
    if (!svgRef.current || metricFlows.length === 0) return;
    
    const svg = d3.select(svgRef.current);
    const width = svgRef.current.clientWidth;
    const height = 300;
    
    // Clear previous render
    svg.selectAll('*').remove();
    
    // Create flow visualization
    const g = svg.append('g');
    
    // Color scale
    const colorScale = d3.scaleSequential(d3.interpolateReds)
      .domain([0, d3.max(metricFlows, d => d.costPerMinute) || 1]);
    
    // Create bars with animation
    const barHeight = 25;
    const barPadding = 5;
    
    const bars = g.selectAll('.metric-bar')
      .data(metricFlows)
      .enter()
      .append('g')
      .attr('class', 'metric-bar')
      .attr('transform', (d, i) => `translate(0, ${i * (barHeight + barPadding)})`)
      .style('cursor', 'pointer')
      .on('click', (event, d) => setSelectedMetric(d.metricName));
    
    // Background bars
    bars.append('rect')
      .attr('width', width - 200)
      .attr('height', barHeight)
      .attr('fill', '#f5f5f5')
      .attr('rx', 4);
    
    // Cost bars with animation
    bars.append('rect')
      .attr('width', 0)
      .attr('height', barHeight)
      .attr('fill', d => colorScale(d.costPerMinute))
      .attr('rx', 4)
      .transition()
      .duration(1000)
      .ease(d3.easeCubicOut)
      .attr('width', d => (d.percentage / 100) * (width - 200));
    
    // Metric names
    bars.append('text')
      .attr('x', 10)
      .attr('y', barHeight / 2)
      .attr('dy', '.35em')
      .attr('font-size', '12px')
      .attr('font-weight', '500')
      .text(d => d.metricName.length > 30 ? d.metricName.substring(0, 30) + '...' : d.metricName);
    
    // Cost labels
    bars.append('text')
      .attr('x', width - 190)
      .attr('y', barHeight / 2)
      .attr('dy', '.35em')
      .attr('text-anchor', 'start')
      .attr('font-size', '12px')
      .attr('font-weight', '600')
      .text(d => `₹${d.costPerMinute.toFixed(2)}/min`);
    
    // Percentage labels
    bars.append('text')
      .attr('x', width - 80)
      .attr('y', barHeight / 2)
      .attr('dy', '.35em')
      .attr('text-anchor', 'start')
      .attr('font-size', '11px')
      .attr('fill', '#666')
      .text(d => `${d.percentage.toFixed(1)}%`);
    
    // Animated flow particles
    const particles = g.selectAll('.particle')
      .data(d3.range(20))
      .enter()
      .append('circle')
      .attr('class', 'particle')
      .attr('r', 2)
      .attr('fill', '#ff6b6b')
      .attr('opacity', 0.6);
    
    function animateParticles() {
      particles
        .attr('cx', -10)
        .attr('cy', () => Math.random() * height)
        .transition()
        .duration(3000)
        .ease(d3.easeLinear)
        .attr('cx', width + 10)
        .on('end', animateParticles);
    }
    
    animateParticles();
    
  }, [metricFlows]);
  
  const handleQuickDeploy = async (metric: string) => {
    // Quick deploy filter for selected metric
    try {
      const response = await fetch('/api/v1/pipelines/quick-deploy', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          pipeline_template: 'filter-metric',
          target_hosts: ['all'],
          filter_metric: metric,
        }),
      });
      
      if (response.ok) {
        // Show success notification
        console.log('Filter deployed successfully');
      }
    } catch (error) {
      console.error('Failed to deploy filter:', error);
    }
  };
  
  return (
    <Card sx={{ height: '100%', position: 'relative', overflow: 'hidden' }}>
      <CardHeader
        title={
          <Box display="flex" alignItems="center" justifyContent="space-between">
            <Typography variant="h5">Live Cost Flow Monitor</Typography>
            <Stack direction="row" spacing={1} alignItems="center">
              <Typography variant="h6" color="primary">
                {formatCurrency(totalCost)}/min
              </Typography>
              <Chip 
                label={`${formatCurrency(totalCost * 60 * 24 * 30)}/month`}
                color="warning"
                size="small"
              />
            </Stack>
          </Box>
        }
      />
      <CardContent>
        <Box position="relative">
          <svg
            ref={svgRef}
            width="100%"
            height={300}
            style={{ overflow: 'visible' }}
          />
          
          <AnimatePresence>
            {selectedMetric && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: 20 }}
                style={{
                  position: 'absolute',
                  bottom: -60,
                  left: '50%',
                  transform: 'translateX(-50%)',
                  background: 'white',
                  padding: '16px',
                  borderRadius: '8px',
                  boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
                  zIndex: 10,
                }}
              >
                <Typography variant="subtitle2" gutterBottom>
                  Quick Actions for {selectedMetric}
                </Typography>
                <Stack direction="row" spacing={1}>
                  <Button
                    size="small"
                    variant="contained"
                    color="primary"
                    onClick={() => handleQuickDeploy(selectedMetric)}
                  >
                    Deploy Filter
                  </Button>
                  <Button
                    size="small"
                    variant="outlined"
                    onClick={() => setSelectedMetric(null)}
                  >
                    Cancel
                  </Button>
                </Stack>
              </motion.div>
            )}
          </AnimatePresence>
        </Box>
        
        <Box mt={3}>
          <Typography variant="caption" color="textSecondary">
            Click any metric to see quick optimization options. 
            Real-time data updates every second.
          </Typography>
        </Box>
      </CardContent>
    </Card>
  );
};