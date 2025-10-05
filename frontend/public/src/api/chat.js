// Public shim for dev UI: provides the same API surface as the SPA's chat API
// so index.html/app.js can import '/src/api/chat.js' during development.

let socket = null

export function connectWebSocket(onMessage) {
  if (socket && socket.readyState === WebSocket.OPEN) return true
  socket = new WebSocket('ws://' + location.host + '/ws')
  socket.onopen = () => {
    window.dispatchEvent(new CustomEvent('chat:status', { detail: { status: 'connected' } }))
  }
  socket.onmessage = (e) => {
    let payload = e.data
    try {
      payload = JSON.parse(e.data)
    } catch (_) {}
    // call callback
    if (typeof onMessage === 'function') onMessage(payload)
    // emit a DOM event so the dev UI can listen without a framework
    window.dispatchEvent(new CustomEvent('chat:message', { detail: payload }))
  }
  socket.onclose = () => window.dispatchEvent(new CustomEvent('chat:status', { detail: { status: 'disconnected' } }))
  socket.onerror = (err) => window.dispatchEvent(new CustomEvent('chat:status', { detail: { status: 'error', error: err } }))
  return true
}

export function disconnectWebSocket() {
  if (!socket) return
  socket.close()
  socket = null
}

export function sendMessage(payload) {
  if (!socket || socket.readyState !== WebSocket.OPEN) return false
  try {
    socket.send(JSON.stringify(payload))
    return true
  } catch (err) {
    return false
  }
}

export function getMessages() {
  // Dev shim doesn't persist messages; SPA store will.
  return []
}
