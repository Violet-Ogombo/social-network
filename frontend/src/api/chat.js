// api/chat.js — thin wrapper that delegates to the Pinia chat store
import { useChatStore } from '@/store/chat'

export const connectWebSocket = () => {
  const store = useChatStore()
  return store.connect()
}

export const disconnectWebSocket = () => {
  const store = useChatStore()
  return store.disconnect()
}

export const sendMessage = (payload) => {
  const store = useChatStore()
  return store.sendMessage(payload)
}

export const getMessages = () => {
  const store = useChatStore()
  return store.messages
}