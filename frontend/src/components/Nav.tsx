import { Link, useNavigate } from 'react-router-dom'

export default function Nav() {
  const navigate = useNavigate()
  const token = localStorage.getItem('token')

  function logout() {
    localStorage.removeItem('token')
    navigate('/login')
  }

  return (
    <nav className="border-b border-gray-800 bg-gray-950">
      <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <Link to="/" className="text-brand-500 font-bold tracking-tight text-lg">
            InferRoute
          </Link>
          {token && (
            <>
              <Link to="/"        className="text-gray-400 hover:text-white text-sm">Dashboard</Link>
              <Link to="/workers" className="text-gray-400 hover:text-white text-sm">Workers</Link>
              <Link to="/keys"    className="text-gray-400 hover:text-white text-sm">API Keys</Link>
            </>
          )}
        </div>
        <div>
          {token ? (
            <button onClick={logout} className="text-sm text-gray-400 hover:text-white">
              Logout
            </button>
          ) : (
            <Link to="/login" className="text-sm text-brand-500 hover:text-brand-400">Login</Link>
          )}
        </div>
      </div>
    </nav>
  )
}
