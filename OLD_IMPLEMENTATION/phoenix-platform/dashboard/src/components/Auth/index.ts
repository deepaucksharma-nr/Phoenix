// Authentication Components
export { PrivateRoute } from './PrivateRoute'
export { UserProfile } from './UserProfile'
export { RoleGuard, usePermissions } from './RoleGuard'
export { AuthProvider, useAuthContext, withAuth } from './AuthProvider'
export { SessionTimeout, useSessionTimeout } from './SessionTimeout'

// Re-export auth store hooks for convenience
export { useAuthStore, useAuth, useRequireAuth } from '../../store/useAuthStore'