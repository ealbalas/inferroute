import { useEffect, useState } from 'react'
import api from '../api/client'

interface APIKey {
  id: string
  name: string
  last_used_at: string | null
  created_at: string
}

export default function ApiKeys() {
  const [keys, setKeys] = useState<APIKey[]>([])
  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyValue, setNewKeyValue] = useState('')
  const [loading, setLoading] = useState(true)

  async function loadKeys() {
    const res = await api.get('/keys')
    setKeys(res.data.keys ?? [])
    setLoading(false)
  }

  async function createKey(e: React.FormEvent) {
    e.preventDefault()
    if (!newKeyName.trim()) return
    const res = await api.post('/keys', { name: newKeyName })
    setNewKeyValue(res.data.key)
    setNewKeyName('')
    loadKeys()
  }

  async function deleteKey(id: string) {
    await api.delete(`/keys/${id}`)
    setKeys((prev) => prev.filter((k) => k.id !== id))
  }

  useEffect(() => { loadKeys() }, [])

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold text-white">API Keys</h1>

      <form onSubmit={createKey} className="flex gap-3">
        <input
          value={newKeyName}
          onChange={(e) => setNewKeyName(e.target.value)}
          placeholder="Key name (e.g. prod-app)"
          className="flex-1 bg-gray-900 border border-gray-700 rounded px-3 py-2 text-white focus:outline-none focus:border-brand-500"
        />
        <button
          type="submit"
          className="bg-brand-500 hover:bg-brand-600 text-white px-4 py-2 rounded font-medium transition-colors"
        >
          Create key
        </button>
      </form>

      {newKeyValue && (
        <div className="bg-green-950 border border-green-700 rounded p-4">
          <p className="text-green-300 text-sm font-medium mb-1">New API key — copy it now, it won't be shown again.</p>
          <code className="text-green-200 break-all">{newKeyValue}</code>
        </div>
      )}

      {loading ? (
        <p className="text-gray-400">Loading keys…</p>
      ) : (
        <div className="divide-y divide-gray-800 border border-gray-800 rounded-lg overflow-hidden">
          {keys.map((k) => (
            <div key={k.id} className="bg-gray-950 px-4 py-3 flex items-center justify-between">
              <div>
                <p className="text-white font-medium">{k.name}</p>
                <p className="text-xs text-gray-500">
                  Created {new Date(k.created_at).toLocaleDateString()}
                  {k.last_used_at && ` · Last used ${new Date(k.last_used_at).toLocaleDateString()}`}
                </p>
              </div>
              <button
                onClick={() => deleteKey(k.id)}
                className="text-red-400 hover:text-red-300 text-sm"
              >
                Revoke
              </button>
            </div>
          ))}
          {keys.length === 0 && (
            <div className="px-4 py-8 text-center text-gray-500">No API keys yet.</div>
          )}
        </div>
      )}
    </div>
  )
}
