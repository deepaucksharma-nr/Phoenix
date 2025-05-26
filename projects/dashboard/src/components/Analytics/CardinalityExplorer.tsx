import React, { useEffect, useState, useRef } from 'react';
import { Box, Card, CardHeader, CardContent, TextField, Typography, Button, Stack, Chip, IconButton } from '@mui/material';
import { Search, FilterList, Download } from '@mui/icons-material';
import * as d3 from 'd3';
import { motion } from 'framer-motion';

interface MetricNode {
  name: string;
  value: number;
  children?: MetricNode[];
}

export const CardinalityExplorer: React.FC = () => {
  const svgRef = useRef<SVGSVGElement>(null);
  const [data, setData] = useState<MetricNode | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedPath, setSelectedPath] = useState<string[]>([]);
  const [hoveredNode, setHoveredNode] = useState<string | null>(null);
  const [totalCardinality, setTotalCardinality] = useState(0);
  
  useEffect(() => {
    fetchCardinalityData();
  }, []);
  
  const fetchCardinalityData = async (namespace?: string, service?: string) => {
    try {
      const params = new URLSearchParams();
      if (namespace) params.append('namespace', namespace);
      if (service) params.append('service', service);
      
      const response = await fetch(`/api/v1/metrics/cardinality?${params}`);
      const cardinalityData = await response.json();
      
      // Transform to hierarchical structure
      const root = transformToHierarchy(cardinalityData);
      setData(root);
      setTotalCardinality(cardinalityData.total_cardinality);
    } catch (error) {
      console.error('Failed to fetch cardinality data:', error);
    }
  };
  
  const transformToHierarchy = (data: any): MetricNode => {
    // Transform flat metric list to hierarchical structure
    const root: MetricNode = {
      name: 'metrics',
      value: 0,
      children: [],
    };
    
    // Group by metric prefix
    const groups = new Map<string, MetricNode>();
    
    Object.entries(data.by_metric || {}).forEach(([metric, cardinality]) => {
      const parts = metric.split('.');
      let current = root;
      
      parts.forEach((part, index) => {
        const path = parts.slice(0, index + 1).join('.');
        
        if (!current.children) {
          current.children = [];
        }
        
        let child = current.children.find(c => c.name === part);
        if (!child) {
          child = {
            name: part,
            value: index === parts.length - 1 ? cardinality as number : 0,
            children: index < parts.length - 1 ? [] : undefined,
          };
          current.children.push(child);
        }
        
        current = child;
      });
    });
    
    return root;
  };
  
  useEffect(() => {
    if (!svgRef.current || !data) return;
    
    const width = svgRef.current.clientWidth;
    const height = 600;
    const radius = Math.min(width, height) / 2;
    
    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();
    
    const g = svg
      .append('g')
      .attr('transform', `translate(${width / 2},${height / 2})`);
    
    // Create partition layout (sunburst)
    const partition = d3.partition<MetricNode>()
      .size([2 * Math.PI, radius]);
    
    const root = d3.hierarchy(data)
      .sum(d => d.value)
      .sort((a, b) => (b.value || 0) - (a.value || 0));
    
    partition(root);
    
    // Color scale
    const color = d3.scaleOrdinal(d3.schemeCategory10);
    
    // Arc generator
    const arc = d3.arc<d3.HierarchyRectangularNode<MetricNode>>()
      .startAngle(d => d.x0)
      .endAngle(d => d.x1)
      .innerRadius(d => d.y0)
      .outerRadius(d => d.y1);
    
    // Create arcs
    const arcs = g.selectAll('path')
      .data(root.descendants())
      .enter()
      .append('path')
      .attr('d', arc)
      .style('fill', d => color((d.children ? d : d.parent)?.data.name || ''))
      .style('stroke', '#fff')
      .style('stroke-width', 2)
      .style('cursor', 'pointer')
      .on('click', (event, d) => {
        const path = d.ancestors().reverse().map(n => n.data.name);
        setSelectedPath(path);
      })
      .on('mouseover', (event, d) => {
        setHoveredNode(d.data.name);
        // Highlight path
        arcs.style('opacity', node => {
          return d.ancestors().includes(node) || node.ancestors().includes(d) ? 1 : 0.3;
        });
      })
      .on('mouseout', () => {
        setHoveredNode(null);
        arcs.style('opacity', 1);
      });
    
    // Add labels for larger segments
    g.selectAll('text')
      .data(root.descendants().filter(d => (d.x1 - d.x0) > 0.1))
      .enter()
      .append('text')
      .attr('transform', d => {
        const angle = (d.x0 + d.x1) / 2;
        const radius = (d.y0 + d.y1) / 2;
        return `rotate(${angle * 180 / Math.PI - 90}) translate(${radius},0) rotate(${angle > Math.PI ? 180 : 0})`;
      })
      .attr('text-anchor', 'middle')
      .attr('dy', '0.35em')
      .style('font-size', '11px')
      .style('pointer-events', 'none')
      .text(d => d.data.name);
    
    // Add center text
    const centerText = g.append('text')
      .attr('text-anchor', 'middle')
      .attr('dy', '0.35em')
      .style('font-size', '16px')
      .style('font-weight', 'bold');
    
    centerText.append('tspan')
      .attr('x', 0)
      .attr('dy', '-0.5em')
      .text('Total Cardinality');
    
    centerText.append('tspan')
      .attr('x', 0)
      .attr('dy', '1.5em')
      .text(d3.format(',')(totalCardinality));
    
  }, [data, totalCardinality]);
  
  const handleSearch = () => {
    // Filter data based on search term
    if (searchTerm) {
      // TODO: Implement search filtering
    } else {
      fetchCardinalityData();
    }
  };
  
  const handleExport = () => {
    // Export cardinality data as CSV
    const csv = generateCSV();
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `cardinality-report-${new Date().toISOString()}.csv`;
    a.click();
  };
  
  const generateCSV = (): string => {
    // TODO: Generate CSV from cardinality data
    return 'metric,cardinality,percentage\n';
  };
  
  return (
    <Card sx={{ height: '100%' }}>
      <CardHeader
        title="Cardinality Explorer"
        subheader={`Total metrics: ${d3.format(',')(totalCardinality)}`}
        action={
          <Stack direction="row" spacing={1}>
            <TextField
              size="small"
              placeholder="Search metrics..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
              InputProps={{
                startAdornment: <Search sx={{ mr: 1, color: 'text.secondary' }} />,
              }}
            />
            <Button
              size="small"
              startIcon={<FilterList />}
              onClick={handleSearch}
            >
              Filter
            </Button>
            <IconButton onClick={handleExport}>
              <Download />
            </IconButton>
          </Stack>
        }
      />
      <CardContent>
        {/* Breadcrumb path */}
        {selectedPath.length > 0 && (
          <Stack direction="row" spacing={1} mb={2}>
            {selectedPath.map((segment, index) => (
              <Chip
                key={index}
                label={segment}
                size="small"
                onClick={() => setSelectedPath(selectedPath.slice(0, index + 1))}
                sx={{ cursor: 'pointer' }}
              />
            ))}
            <Button
              size="small"
              onClick={() => setSelectedPath([])}
            >
              Clear
            </Button>
          </Stack>
        )}
        
        {/* Sunburst chart */}
        <Box position="relative">
          <svg
            ref={svgRef}
            width="100%"
            height={600}
            style={{ overflow: 'visible' }}
          />
          
          {/* Hover tooltip */}
          {hoveredNode && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              style={{
                position: 'absolute',
                bottom: 10,
                left: '50%',
                transform: 'translateX(-50%)',
                background: 'rgba(0, 0, 0, 0.8)',
                color: 'white',
                padding: '8px 16px',
                borderRadius: '4px',
                pointerEvents: 'none',
              }}
            >
              <Typography variant="body2">{hoveredNode}</Typography>
            </motion.div>
          )}
        </Box>
        
        {/* Instructions */}
        <Box mt={2}>
          <Typography variant="caption" color="textSecondary">
            Click segments to drill down • Hover to highlight paths • Larger segments = higher cardinality
          </Typography>
        </Box>
      </CardContent>
    </Card>
  );
};