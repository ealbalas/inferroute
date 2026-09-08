import { useEffect, useRef, useState } from 'react'

interface UseWebSocketOptions {
  onMessage: (data: unknown) => void
  reconnectDelay?: number
}

// useWebSocket connects to a WebSocket endpoint and calls onMessage on each event.
// It automatically reconnects with an exponential backoff on disconnect.
export function useWebSocket(url: string, { onMessage, reconnectDelay = 2000 }: UseWebSocketOptions) {
  const [connected, setConnected] = useState(false)
  const ws = useRef<WebSocket | null>(null)
  const retries = useRef(0)

  useEffect(() => {
    let timeout: ReturnType<typeof setTimeout>

    function connect() {
      ws.current = new WebSocket(url)

      ws.current.onopen = () => {
        setConnected(true)
        retries.current = 0
      }

      ws.current.onmessage = (event) => {
        try {
          onMessage(JSON.parse(event.data))
        } catch {
          // ignore malformed frames
        }
      }

      ws.current.onclose = () => {
        setConnected(false)
        // Exponential backoff, cap at 30 s
        const delay = Math.min(reconnectDelay * 2 ** retries.current, 30000)
        retries.current++
        timeout = setTimeout(connect, delay)
      }
    }

    connect()
    return () => {
      ws.current?.close()
      clearTimeout(timeout)
    }
  }, [url]) // eslint-disable-line react-hooks/exhaustive-deps

  return { connected }
}
