import { createBrowserRouter, Navigate } from 'react-router-dom';
import { lazy, Suspense } from 'react';
import { MainLayout } from '@components/Layout/MainLayout';
import { CircularProgress, Box } from '@mui/material';

// Lazy load pages for better performance
const Dashboard = lazy(() => import('@pages/Dashboard'));
const Experiments = lazy(() => import('@pages/Experiments'));
const ExperimentDetails = lazy(() => import('@pages/ExperimentDetails'));
const Pipelines = lazy(() => import('@pages/Pipelines'));
const DeployedPipelines = lazy(() => import('@pages/DeployedPipelines'));
const PipelineCatalog = lazy(() => import('@pages/PipelineCatalog'));
const Analysis = lazy(() => import('@pages/Analysis'));
const Settings = lazy(() => import('@pages/Settings'));

// Loading component
const PageLoader = () => (
  <Box
    sx={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      height: '100vh',
    }}
  >
    <CircularProgress />
  </Box>
);

// Suspense wrapper for lazy loaded components
const SuspenseWrapper = ({ children }: { children: React.ReactNode }) => (
  <Suspense fallback={<PageLoader />}>{children}</Suspense>
);

export const router = createBrowserRouter([
  {
    path: '/',
    element: <MainLayout />,
    children: [
      {
        index: true,
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: 'dashboard',
        element: (
          <SuspenseWrapper>
            <Dashboard />
          </SuspenseWrapper>
        ),
      },
      {
        path: 'experiments',
        children: [
          {
            index: true,
            element: (
              <SuspenseWrapper>
                <Experiments />
              </SuspenseWrapper>
            ),
          },
          {
            path: ':experimentId',
            element: (
              <SuspenseWrapper>
                <ExperimentDetails />
              </SuspenseWrapper>
            ),
          },
        ],
      },
      {
        path: 'pipeline-viewer',
        element: (
          <SuspenseWrapper>
            <Pipelines />
          </SuspenseWrapper>
        ),
      },
      {
        path: 'pipelines',
        children: [
          {
            index: true,
            element: (
              <SuspenseWrapper>
                <DeployedPipelines />
              </SuspenseWrapper>
            ),
          },
          {
            path: 'catalog',
            element: (
              <SuspenseWrapper>
                <PipelineCatalog />
              </SuspenseWrapper>
            ),
          },
        ],
      },
      {
        path: 'analysis',
        element: (
          <SuspenseWrapper>
            <Analysis />
          </SuspenseWrapper>
        ),
      },
      {
        path: 'settings',
        element: (
          <SuspenseWrapper>
            <Settings />
          </SuspenseWrapper>
        ),
      },
    ],
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);

export default router;