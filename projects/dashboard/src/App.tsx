import { RouterProvider } from 'react-router-dom'
import { NotificationProvider } from '@/components/Notifications/NotificationProvider'
import router from '@/router'

function App() {
  return (
    <NotificationProvider>
      <RouterProvider router={router} />
    </NotificationProvider>
  )
}

export default App