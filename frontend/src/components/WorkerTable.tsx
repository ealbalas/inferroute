export interface Worker {
  Name: string
  Status: string
  AvgLatencyMS: number
  Load: number
  ErrorRate: number
  LastSeen: string
}

interface WorkerTableProps {
  workers: Worker[]
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    healthy:  'bg-green-900 text-green-300',
    degraded: 'bg-yellow-900 text-yellow-300',
    offline:  'bg-red-900 text-red-300',
    unknown:  'bg-gray-800 text-gray-400',
  }
  return (
    <span className={`px-2 py-0.5 rounded text-xs font-medium ${colors[status] ?? colors.unknown}`}>
      {status}
    </span>
  )
}

export default function WorkerTable({ workers }: WorkerTableProps) {
  return (
    <div className="overflow-x-auto rounded-lg border border-gray-800">
      <table className="w-full text-sm text-left">
        <thead className="bg-gray-900 text-gray-400 text-xs uppercase tracking-wider">
          <tr>
            <th className="px-4 py-3">Worker</th>
            <th className="px-4 py-3">Status</th>
            <th className="px-4 py-3">Latency</th>
            <th className="px-4 py-3">Load</th>
            <th className="px-4 py-3">Error Rate</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-800">
          {workers.map((w) => (
            <tr key={w.Name} className="bg-gray-950 hover:bg-gray-900 transition-colors">
              <td className="px-4 py-3 font-medium text-white">{w.Name}</td>
              <td className="px-4 py-3"><StatusBadge status={w.Status} /></td>
              <td className="px-4 py-3 text-gray-300">{w.AvgLatencyMS.toFixed(0)} ms</td>
              <td className="px-4 py-3 text-gray-300">{(w.Load * 100).toFixed(1)}%</td>
              <td className="px-4 py-3 text-gray-300">{(w.ErrorRate * 100).toFixed(2)}%</td>
            </tr>
          ))}
          {workers.length === 0 && (
            <tr>
              <td colSpan={5} className="px-4 py-8 text-center text-gray-500">
                No workers found.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  )
}
