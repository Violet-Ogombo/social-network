import { defineStore } from 'pinia'

export const useChatStore = defineStore('chat', {
	state: () => ({
		socket: null,
		messages: [],
		connected: false,
	}),
	actions: {
		connect() {
			if (this.socket && this.connected) return
			this.socket = new WebSocket('ws://' + location.host + '/ws')
			this.socket.onopen = () => {
				this.connected = true
				console.log('chat socket open')
			}
			this.socket.onmessage = (e) => {
				try {
					const msg = JSON.parse(e.data)
					this.addMessage(msg)
				} catch (err) {
					console.error('invalid message', err)
				}
			}
			this.socket.onclose = () => {
				this.connected = false
				this.socket = null
				console.log('chat socket closed')
			}
			this.socket.onerror = (err) => console.error('chat socket error', err)
		},
		disconnect() {
			if (this.socket) {
				this.socket.close()
				this.socket = null
				this.connected = false
			}
		},
		sendMessage(payload) {
			if (!this.socket || !this.connected) {
				console.warn('socket not connected')
				return false
			}
			this.socket.send(JSON.stringify(payload))
			return true
		},
		addMessage(msg) {
			this.messages.push(msg)
		}
	}
})
