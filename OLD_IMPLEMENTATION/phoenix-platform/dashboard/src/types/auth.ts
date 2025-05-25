export interface User {
  id: string
  email: string
  name: string
  role: 'admin' | 'user' | 'viewer'
  organization?: string
}

export interface AuthResponse {
  user: User
  token: string
}

export interface RegisterData {
  name: string
  email: string
  password: string
  organization: string
}

export interface WebSocketMessage {
  type: 'experiment.update' | 'metrics.update' | 'alert' | 'notification'
  payload: any
  timestamp: string
}

export interface Notification {
  id: string
  type: 'info' | 'success' | 'warning' | 'error'
  title: string
  message?: string
  timestamp: string
  read: boolean
  experimentId?: string
}