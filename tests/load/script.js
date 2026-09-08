/**
 * k6 load test for InferRoute
 *
 * Usage:
 *   k6 run tests/load/script.js
 *
 * Set API_URL and API_KEY via env:
 *   k6 run -e API_URL=http://localhost:8080 -e API_KEY=ir_live_... tests/load/script.js
 */

import http from 'k6/http'
import { check, sleep } from 'k6'
import { Rate, Trend } from 'k6/metrics'

const errorRate     = new Rate('errors')
const cacheHitRate  = new Rate('cache_hits')
const latencyTrend  = new Trend('request_latency_ms', true)

export const options = {
  stages: [
    { duration: '1m',  target: 10  },  // warm up
    { duration: '3m',  target: 50  },  // ramp to 50 VUs
    { duration: '3m',  target: 100 },  // peak load
    { duration: '1m',  target: 0   },  // cool down
  ],
  thresholds: {
    http_req_duration:   ['p(95)<500'],  // 95th percentile < 500 ms
    http_req_failed:     ['rate<0.05'],  // error rate < 5%
    errors:              ['rate<0.05'],
  },
}

const BASE_URL = __ENV.API_URL || 'http://localhost:8080'
const API_KEY  = __ENV.API_KEY  || 'ir_live_placeholder'

const prompts = [
  'Explain TCP congestion control',
  'What is the CAP theorem?',
  'How does Redis expire keys?',
  'What is consistent hashing?',
  'Explain exponential backoff',
]

export default function () {
  const prompt = prompts[Math.floor(Math.random() * prompts.length)]

  const res = http.post(
    `${BASE_URL}/v1/generate`,
    JSON.stringify({ prompt, priority: 'latency' }),
    {
      headers: {
        'Content-Type':  'application/json',
        'Authorization': `Bearer ${API_KEY}`,
      },
      timeout: '10s',
    },
  )

  const ok = check(res, {
    'status is 200': (r) => r.status === 200,
    'has output':    (r) => r.json('output') !== '',
  })

  errorRate.add(!ok)
  cacheHitRate.add(res.json('cached') === true)
  latencyTrend.add(res.timings.duration)

  sleep(0.1)
}
