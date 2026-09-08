import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import axios from 'axios'

export default function Register() {
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const res = await axios.post('/v1/auth/register', { email, password })
      localStorage.setItem('token', res.data.token)
      navigate('/')
    } catch (err) {
      setError(axios.isAxiosError(err) ? (err.response?.data?.error ?? 'Registration failed.') : 'Registration failed.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-sm mx-auto mt-20">
      <h1 className="text-2xl font-bold text-white mb-6">Create an account</h1>
      <form onSubmit={submit} className="space-y-4">
        <div>
          <label className="block text-sm text-gray-400 mb-1">Email</label>
          <input
            type="email" value={email} onChange={(e) => setEmail(e.target.value)}
            required
            className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-white focus:outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Password (min 8 chars)</label>
          <input
            type="password" value={password} onChange={(e) => setPassword(e.target.value)}
            required minLength={8}
            className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-white focus:outline-none focus:border-brand-500"
          />
        </div>
        {error && <p className="text-red-400 text-sm">{error}</p>}
        <button
          type="submit" disabled={loading}
          className="w-full bg-brand-500 hover:bg-brand-600 text-white py-2 rounded font-medium transition-colors disabled:opacity-50"
        >
          {loading ? 'Creating account…' : 'Create account'}
        </button>
      </form>
      <p className="text-gray-400 text-sm mt-4">
        Already have an account?{' '}
        <Link to="/login" className="text-brand-500 hover:underline">Sign in</Link>
      </p>
    </div>
  )
}
