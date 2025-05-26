import { useAppSelector, useAppDispatch } from './redux'
import { login, logout, register, checkAuth } from '@store/slices/authSlice'

export const useAuth = () => {
  const dispatch = useAppDispatch()
  const { user, token, isAuthenticated, loading: isLoading, error } = useAppSelector(state => state.auth)

  return {
    user,
    token,
    isAuthenticated,
    isLoading,
    error,
    login: (credentials: any) => dispatch(login(credentials)),
    logout: () => dispatch(logout()),
    register: (data: any) => dispatch(register(data)),
    checkAuth: () => dispatch(checkAuth()),
  }
}