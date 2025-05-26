// Environment configuration with type safety

interface EnvConfig {
  API_BASE_URL: string;
  WS_URL: string;
  AUTH_TIMEOUT: number;
  REFRESH_TOKEN_INTERVAL: number;
  ENABLE_WEBSOCKET: boolean;
  ENABLE_ANALYTICS: boolean;
  ENABLE_EXPERIMENT_BUILDER: boolean;
  PROMETHEUS_URL: string;
  GRAFANA_URL: string;
  ENV: 'development' | 'staging' | 'production';
}

const getEnvVar = (key: string, defaultValue?: string): string => {
  const value = import.meta.env[`VITE_${key}`];
  if (value === undefined && defaultValue === undefined) {
    throw new Error(`Environment variable VITE_${key} is not defined`);
  }
  return value || defaultValue || '';
};

const getBooleanEnvVar = (key: string, defaultValue = false): boolean => {
  const value = import.meta.env[`VITE_${key}`];
  if (value === undefined) return defaultValue;
  return value === 'true' || value === true;
};

const getNumberEnvVar = (key: string, defaultValue?: number): number => {
  const value = import.meta.env[`VITE_${key}`];
  if (value === undefined && defaultValue === undefined) {
    throw new Error(`Environment variable VITE_${key} is not defined`);
  }
  const parsed = parseInt(value || '', 10);
  return isNaN(parsed) ? (defaultValue || 0) : parsed;
};

export const env: EnvConfig = {
  API_BASE_URL: getEnvVar('API_BASE_URL', 'http://localhost:8080/api/v1'),
  WS_URL: getEnvVar('WS_URL', 'ws://localhost:8080/ws'),
  AUTH_TIMEOUT: getNumberEnvVar('AUTH_TIMEOUT', 3600000),
  REFRESH_TOKEN_INTERVAL: getNumberEnvVar('REFRESH_TOKEN_INTERVAL', 300000),
  ENABLE_WEBSOCKET: getBooleanEnvVar('ENABLE_WEBSOCKET', true),
  ENABLE_ANALYTICS: getBooleanEnvVar('ENABLE_ANALYTICS', true),
  ENABLE_EXPERIMENT_BUILDER: getBooleanEnvVar('ENABLE_EXPERIMENT_BUILDER', true),
  PROMETHEUS_URL: getEnvVar('PROMETHEUS_URL', 'http://localhost:9090'),
  GRAFANA_URL: getEnvVar('GRAFANA_URL', 'http://localhost:3000'),
  ENV: (getEnvVar('ENV', 'development') as EnvConfig['ENV']),
};

// Validate environment on startup
export const validateEnvironment = (): void => {
  const requiredVars = ['API_BASE_URL', 'WS_URL'];
  const missing = requiredVars.filter(
    (key) => !import.meta.env[`VITE_${key}`]
  );
  
  if (missing.length > 0) {
    console.warn(
      `Missing environment variables: ${missing.map((k) => `VITE_${k}`).join(', ')}`
    );
  }
};

export default env;