import { useEffect, useState } from 'react'
import WorkerTable, { type Worker } from '../components/WorkerTable'
import api from '../api/client'

export default function Workers() {
  const [workers, setWorkers] = useState<Worker[]>([])
  const [loading, setLoading] = useState(true)

  async function load() {
    try {
      const res = await api.get('/workers')
      setWorkers(res.data.workers ?? [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
    const interval = setInterval(load, 5000)
    return () => clearInterval(interval)
  }, [])

  const healthy  = workers.filter((w) => w.Status === 'healthy').length
  const degraded = workers.filter((w) => w.Status === 'degraded').length
  const offline  = workers.filter((w) => w.Status === 'offline').length

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-white">Workers</h1>
        <p className="text-gray-400 text-sm mt-1">Polling every 5 s · {workers.length} workers registered</p>
      </div>

      <div className="flex gap-4 text-sm">
        <span className="text-green-400">{healthy} healthy</span>
        <span className="text-yellow-400">{degraded} degraded</span>
        <span className="text-red-400">{offline} offline</span>
      </div>

      {loading ? (
        <p className="text-gray-400">Loading workers…</p>
      ) : (
        <WorkerTable workers={workers} />
      )}
    </div>
  )
}
