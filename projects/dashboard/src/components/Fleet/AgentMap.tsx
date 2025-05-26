import React, { useEffect, useState } from 'react';
import { Box, Card, CardHeader, CardContent, Typography, Chip, Stack, Avatar, AvatarGroup, Tooltip } from '@mui/material';
import { ComputerOutlined, CheckCircleOutline, ErrorOutline, UpdateOutlined } from '@mui/icons-material';
import { motion } from 'framer-motion';
import { useWebSocket } from '../../hooks/useWebSocket';

interface AgentLocation {
  hostId: string;
  hostname: string;
  status: 'healthy' | 'updating' | 'offline';
  group: string;
  location?: {
    region: string;
    zone: string;
  };
  costSavings: number;
  metricsPerSec: number;
  activeTasks: number;
}

interface AgentGroup {
  name: string;
  agents: AgentLocation[];
  totalSavings: number;
  status: 'healthy' | 'partial' | 'offline';
}

export const AgentMap: React.FC = () => {
  const [agentGroups, setAgentGroups] = useState<AgentGroup[]>([]);
  const [selectedGroup, setSelectedGroup] = useState<string | null>(null);
  const { subscribe } = useWebSocket();
  
  useEffect(() => {
    // Subscribe to agent status updates
    const unsubscribe = subscribe('agent_status', (data) => {
      updateAgentStatus(data);
    });
    
    // Fetch initial fleet status
    fetchFleetStatus();
    
    return unsubscribe;
  }, [subscribe]);
  
  const fetchFleetStatus = async () => {
    try {
      const response = await fetch('/api/v1/fleet/status');
      const data = await response.json();
      organizeAgentsByGroup(data.agents);
    } catch (error) {
      console.error('Failed to fetch fleet status:', error);
    }
  };
  
  const updateAgentStatus = (agentUpdate: any) => {
    setAgentGroups(prev => {
      const updated = [...prev];
      // Update specific agent in the groups
      for (const group of updated) {
        const agentIndex = group.agents.findIndex(a => a.hostId === agentUpdate.host_id);
        if (agentIndex !== -1) {
          group.agents[agentIndex] = {
            ...group.agents[agentIndex],
            status: agentUpdate.status,
            metricsPerSec: agentUpdate.metrics.metrics_per_sec,
            costSavings: agentUpdate.cost_savings,
            activeTasks: agentUpdate.active_tasks.length,
          };
          // Recalculate group totals
          group.totalSavings = group.agents.reduce((sum, a) => sum + a.costSavings, 0);
          group.status = calculateGroupStatus(group.agents);
        }
      }
      return updated;
    });
  };
  
  const organizeAgentsByGroup = (agents: any[]) => {
    const groupMap = new Map<string, AgentLocation[]>();
    
    agents.forEach(agent => {
      const group = agent.group || 'default';
      if (!groupMap.has(group)) {
        groupMap.set(group, []);
      }
      
      groupMap.get(group)!.push({
        hostId: agent.host_id,
        hostname: agent.hostname || agent.host_id,
        status: agent.status,
        group: group,
        location: agent.location,
        costSavings: agent.cost_savings || 0,
        metricsPerSec: agent.metrics?.metrics_per_sec || 0,
        activeTasks: agent.active_tasks?.length || 0,
      });
    });
    
    const groups: AgentGroup[] = Array.from(groupMap.entries()).map(([name, agents]) => ({
      name,
      agents,
      totalSavings: agents.reduce((sum, a) => sum + a.costSavings, 0),
      status: calculateGroupStatus(agents),
    }));
    
    setAgentGroups(groups);
  };
  
  const calculateGroupStatus = (agents: AgentLocation[]): 'healthy' | 'partial' | 'offline' => {
    const healthyCount = agents.filter(a => a.status === 'healthy').length;
    if (healthyCount === agents.length) return 'healthy';
    if (healthyCount === 0) return 'offline';
    return 'partial';
  };
  
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
        return <CheckCircleOutline sx={{ color: 'success.main' }} />;
      case 'updating':
        return <UpdateOutlined sx={{ color: 'warning.main' }} />;
      case 'offline':
        return <ErrorOutline sx={{ color: 'error.main' }} />;
      default:
        return <ComputerOutlined />;
    }
  };
  
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'success';
      case 'updating':
        return 'warning';
      case 'offline':
        return 'error';
      default:
        return 'default';
    }
  };
  
  return (
    <Card sx={{ height: '100%' }}>
      <CardHeader
        title="Fleet Status Map"
        subheader={`${agentGroups.reduce((sum, g) => sum + g.agents.length, 0)} agents across ${agentGroups.length} groups`}
      />
      <CardContent>
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
            gap: 2,
          }}
        >
          {agentGroups.map((group) => (
            <motion.div
              key={group.name}
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.3 }}
            >
              <Card
                variant="outlined"
                sx={{
                  cursor: 'pointer',
                  transition: 'all 0.2s',
                  '&:hover': {
                    boxShadow: 3,
                    transform: 'translateY(-2px)',
                  },
                  borderColor: selectedGroup === group.name ? 'primary.main' : 'divider',
                  borderWidth: selectedGroup === group.name ? 2 : 1,
                }}
                onClick={() => setSelectedGroup(selectedGroup === group.name ? null : group.name)}
              >
                <CardContent>
                  <Stack direction="row" justifyContent="space-between" alignItems="center" mb={1}>
                    <Typography variant="h6">{group.name}</Typography>
                    <Chip
                      label={group.status}
                      color={getStatusColor(group.status) as any}
                      size="small"
                      icon={getStatusIcon(group.status)}
                    />
                  </Stack>
                  
                  <Stack direction="row" justifyContent="space-between" alignItems="center" mb={2}>
                    <Typography variant="body2" color="textSecondary">
                      {group.agents.length} agents
                    </Typography>
                    <Typography variant="body2" color="success.main" fontWeight="bold">
                      ₹{group.totalSavings.toFixed(0)} saved/hr
                    </Typography>
                  </Stack>
                  
                  <AvatarGroup max={6} sx={{ justifyContent: 'flex-start' }}>
                    {group.agents.map((agent) => (
                      <Tooltip
                        key={agent.hostId}
                        title={
                          <Box>
                            <Typography variant="caption" display="block">
                              {agent.hostname}
                            </Typography>
                            <Typography variant="caption" display="block">
                              Status: {agent.status}
                            </Typography>
                            <Typography variant="caption" display="block">
                              {agent.metricsPerSec} metrics/sec
                            </Typography>
                            {agent.activeTasks > 0 && (
                              <Typography variant="caption" display="block">
                                {agent.activeTasks} active tasks
                              </Typography>
                            )}
                          </Box>
                        }
                      >
                        <Avatar
                          sx={{
                            width: 32,
                            height: 32,
                            bgcolor: `${getStatusColor(agent.status)}.main`,
                            fontSize: '0.75rem',
                          }}
                        >
                          {agent.hostname.substring(0, 2).toUpperCase()}
                        </Avatar>
                      </Tooltip>
                    ))}
                  </AvatarGroup>
                  
                  {selectedGroup === group.name && (
                    <motion.div
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ opacity: 1, height: 'auto' }}
                      exit={{ opacity: 0, height: 0 }}
                    >
                      <Box mt={2} pt={2} borderTop={1} borderColor="divider">
                        {group.agents.map((agent) => (
                          <Stack
                            key={agent.hostId}
                            direction="row"
                            justifyContent="space-between"
                            alignItems="center"
                            py={0.5}
                          >
                            <Stack direction="row" alignItems="center" spacing={1}>
                              {getStatusIcon(agent.status)}
                              <Typography variant="body2">{agent.hostname}</Typography>
                            </Stack>
                            <Stack direction="row" spacing={2}>
                              <Typography variant="caption" color="textSecondary">
                                {agent.metricsPerSec} m/s
                              </Typography>
                              <Typography variant="caption" color="success.main">
                                ₹{agent.costSavings.toFixed(0)}/hr
                              </Typography>
                            </Stack>
                          </Stack>
                        ))}
                      </Box>
                    </motion.div>
                  )}
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </Box>
        
        <Box mt={3}>
          <Stack direction="row" spacing={2} alignItems="center">
            <Typography variant="caption" color="textSecondary">
              Legend:
            </Typography>
            <Chip icon={<CheckCircleOutline />} label="Healthy" color="success" size="small" />
            <Chip icon={<UpdateOutlined />} label="Updating" color="warning" size="small" />
            <Chip icon={<ErrorOutline />} label="Offline" color="error" size="small" />
          </Stack>
        </Box>
      </CardContent>
    </Card>
  );
};