import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Login from './pages/Login'
import Register from './pages/Register'
import ApiKeys from './pages/ApiKeys'
import Workers from './pages/Workers'
import Nav from './components/Nav'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token')
  return token ? <>{children}</> : <Navigate to="/login" replace />
}

export default function App() {
  return (
    <BrowserRouter>
      <div className="min-h-screen">
        <Nav />
        <main className="max-w-7xl mx-auto px-4 py-8">
          <Routes>
            <Route path="/login"    element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/" element={<PrivateRoute><Dashboard /></PrivateRoute>} />
            <Route path="/keys"    element={<PrivateRoute><ApiKeys /></PrivateRoute>} />
            <Route path="/workers" element={<PrivateRoute><Workers /></PrivateRoute>} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}
