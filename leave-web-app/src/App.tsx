import { Routes, Route, Navigate } from 'react-router-dom'
import { UserProvider, useUser } from './context/UserContext'
import LoginPage from './pages/LoginPage'
import EmployeePage from './pages/EmployeePage'
import ManagerPage from './pages/ManagerPage'
import NavBar from './components/NavBar'

function AppRoutes() {
  const { user } = useUser()

  if (!user) {
    return (
      <Routes>
        <Route path="*" element={<LoginPage />} />
      </Routes>
    )
  }

  return (
    <>
      <NavBar />
      <div className="main-content">
        <Routes>
          {user.role === 'employee' && (
            <>
              <Route path="/" element={<EmployeePage />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </>
          )}
          {user.role === 'manager' && (
            <>
              <Route path="/" element={<ManagerPage />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </>
          )}
        </Routes>
      </div>
    </>
  )
}

export default function App() {
  return (
    <UserProvider>
      <AppRoutes />
    </UserProvider>
  )
}
