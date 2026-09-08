import { useEffect, useState } from 'react'
import MetricCard from '../components/MetricCard'
import WorkerTable, { type Worker } from '../components/WorkerTable'
import api from '../api/client'

interface Metrics {
  requestsPerSec: number
  p50LatencyMS: number
  p95LatencyMS: number
  p99LatencyMS: number
  errorRate: number
  cacheHitRate: number
  activeWorkers: number
  totalWorkers: number
}

const DEFAULT_METRICS: Metrics = {
  requestsPerSec: 0,
  p50LatencyMS: 0,
  p95LatencyMS: 0,
  p99LatencyMS: 0,
  errorRate: 0,
  cacheHitRate: 0,
  activeWorkers: 0,
  totalWorkers: 0,
}

export default function Dashboard() {
  const [metrics, setMetrics] = useState<Metrics>(DEFAULT_METRICS)
  const [workers, setWorkers] = useState<Worker[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function load() {
      try {
        const res = await api.get('/workers')
        const ws: Worker[] = res.data.workers ?? []
        setWorkers(ws)
        setMetrics((m) => ({
          ...m,
          totalWorkers: ws.length,
          activeWorkers: ws.filter((w) => w.Status === 'healthy').length,
        }))
      } catch (e) {
        console.error(e)
      } finally {
        setLoading(false)
      }
    }
    load()
    // Poll every 5 s — replace with WebSocket in Phase 2
    const interval = setInterval(load, 5000)
    return () => clearInterval(interval)
  }, [])

  if (loading) {
    return <p className="text-gray-400">Loading dashboard...</p>
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-white">Dashboard</h1>
        <p className="text-gray-400 text-sm mt-1">Live system overview · auto-refreshes every 5 s</p>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <MetricCard title="Requests / sec" value={metrics.requestsPerSec.toFixed(1)} highlight />
        <MetricCard title="P50 Latency"    value={metrics.p50LatencyMS.toFixed(0)} unit="ms" />
        <MetricCard title="P95 Latency"    value={metrics.p95LatencyMS.toFixed(0)} unit="ms" />
        <MetricCard title="P99 Latency"    value={metrics.p99LatencyMS.toFixed(0)} unit="ms" />
        <MetricCard title="Error Rate"      value={(metrics.errorRate * 100).toFixed(2)} unit="%" trend="neutral" />
        <MetricCard title="Cache Hit Rate"  value={(metrics.cacheHitRate * 100).toFixed(1)} unit="%" />
        <MetricCard title="Active Workers"  value={`${metrics.activeWorkers} / ${metrics.totalWorkers}`} />
      </div>

      <div>
        <h2 className="text-lg font-semibold text-white mb-3">Worker Health</h2>
        <WorkerTable workers={workers} />
      </div>
    </div>
  )
}
