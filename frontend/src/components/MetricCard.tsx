interface MetricCardProps {
  title: string
  value: string | number
  unit?: string
  trend?: 'up' | 'down' | 'neutral'
  highlight?: boolean
}

export default function MetricCard({ title, value, unit, trend, highlight }: MetricCardProps) {
  const trendColor = trend === 'up' ? 'text-green-400' : trend === 'down' ? 'text-red-400' : 'text-gray-400'
  const border = highlight ? 'border-brand-500' : 'border-gray-800'

  return (
    <div className={`bg-gray-900 border ${border} rounded-lg p-5`}>
      <p className="text-xs text-gray-500 uppercase tracking-widest mb-1">{title}</p>
      <div className="flex items-end gap-1">
        <span className="text-3xl font-bold text-white">{value}</span>
        {unit && <span className="text-sm text-gray-400 mb-1">{unit}</span>}
      </div>
      {trend && (
        <span className={`text-xs ${trendColor} mt-1 block`}>
          {trend === 'up' ? '▲' : trend === 'down' ? '▼' : '—'} live
        </span>
      )}
    </div>
  )
}
