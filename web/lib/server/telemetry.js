function telemetryEnabled() {
  const raw = String(process.env.NEXT_TELEMETRY_ENABLED || 'true').trim().toLowerCase()
  return raw !== 'false' && raw !== '0' && raw !== 'off'
}

function sanitize(value) {
  if (value === null || value === undefined) return null
  if (typeof value === 'string') return value.slice(0, 240)
  if (typeof value === 'number' || typeof value === 'boolean') return value
  if (Array.isArray(value)) return value.slice(0, 20).map((item) => sanitize(item))
  if (typeof value === 'object') {
    const output = {}
    for (const [key, item] of Object.entries(value).slice(0, 30)) {
      output[key] = sanitize(item)
    }
    return output
  }
  return String(value)
}

export async function trackEvent(event, payload = {}) {
  if (!telemetryEnabled()) return
  const safeEvent = String(event || '').trim().toLowerCase()
  if (!safeEvent) return

  const body = {
    event: safeEvent,
    at: new Date().toISOString(),
    payload: sanitize(payload),
  }

  console.log(`web_telemetry ${JSON.stringify(body)}`)
}
