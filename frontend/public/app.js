// Dev UI application logic (moved out of index.html)
const statusEl = document.getElementById('status')
const userEl = document.getElementById('user')
const userIdEl = document.getElementById('userId')
const formsEl = document.getElementById('forms')

async function checkSession() {
  try {
    const res = await fetch('/api/check-session', { credentials: 'include' })
    if (!res.ok) throw new Error('no session')
    const body = await res.json()
    showUser(body.user_id)
  } catch (e) {
    showForms()
  }
}

function showUser(id) {
  statusEl.textContent = 'You are signed in.'
  userIdEl.textContent = id
  userEl.classList.remove('hidden')
  formsEl.classList.add('hidden')
}

function showForms() {
  statusEl.textContent = 'Not signed in.'
  userEl.classList.add('hidden')
  formsEl.classList.remove('hidden')
}

// Login
document.getElementById('loginBtn').addEventListener('click', async () => {
  const identifier = document.getElementById('loginIdentifier').value
  const password = document.getElementById('loginPassword').value
  const msg = document.getElementById('loginMsg')
  msg.textContent = ''
  try {
    const res = await fetch('/login', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ identifier, password }),
    })
    if (!res.ok) {
      const err = await res.text()
      msg.textContent = err
      return
    }
    const body = await res.json()
    showUser(body.user_id)
  } catch (e) {
    msg.textContent = e.message
  }
})

// Register
document.getElementById('regBtn').addEventListener('click', async () => {
  const data = {
    email: document.getElementById('regEmail').value,
    password: document.getElementById('regPassword').value,
    first_name: document.getElementById('regFirst').value,
    last_name: document.getElementById('regLast').value,
    nickname: document.getElementById('regNickname').value,
    date_of_birth: document.getElementById('regDob').value,
    avatar: document.getElementById('regAvatar').value,
    about: document.getElementById('regAbout').value,
    profile_type: 'public'
  }
  const msg = document.getElementById('regMsg')
  msg.textContent = ''
  try {
    const res = await fetch('/register', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    })
    if (!res.ok) {
      const err = await res.text()
      msg.textContent = err
      return
    }
    const body = await res.json()
    msg.textContent = body.message || 'Registration successful'
  } catch (e) {
    msg.textContent = e.message
  }
})

// Logout
document.getElementById('logoutBtn').addEventListener('click', async () => {
  await fetch('/logout', { method: 'POST', credentials: 'include' })
  showForms()
})

// initial check
checkSession()

// WebSocket: delegate to chat API (Pinia store)
import { connectWebSocket, sendMessage } from '/src/api/chat.js'

function showChat() {
  document.getElementById('chat').classList.remove('hidden')
}

function appendMsg(text) {
  const box = document.getElementById('messages')
  const el = document.createElement('div')
  el.textContent = text
  box.appendChild(el)
  box.scrollTop = box.scrollHeight
}

// open ws after successful login or valid session
function openAfterAuth() {
  connectWebSocket()
  showChat()
}

// call openAfterAuth when we get a valid user
const originalShowUser = showUser
showUser = (id) => { originalShowUser(id); openAfterAuth(); }

// Send chat messages
document.getElementById('chatSend').addEventListener('click', async () => {
  const receiver = document.getElementById('chatReceiver').value
  const content = document.getElementById('chatContent').value
  const payload = { type: 'message', receiver_id: receiver, content }
  const ok = sendMessage(payload)
  if (!ok) document.getElementById('chatStatus').textContent = 'Socket not connected'
  else document.getElementById('chatContent').value = ''
})

// Listen for incoming messages from the shim
window.addEventListener('chat:message', (e) => {
  const payload = e.detail
  appendMsg(JSON.stringify(payload))
})

window.addEventListener('chat:status', (e) => {
  const s = e.detail && e.detail.status
  const el = document.getElementById('chatStatus')
  if (s === 'connected') el.textContent = 'Connected'
  else if (s === 'disconnected') el.textContent = 'Disconnected'
  else if (s === 'error') el.textContent = 'Error'
})
