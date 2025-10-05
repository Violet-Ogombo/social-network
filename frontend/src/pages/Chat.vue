<template>
	<div class="page card">
		<h2>Chat</h2>
		<div>
			<button @click="connect">Connect</button>
			<button @click="disconnect">Disconnect</button>
		</div>

		<div class="messages" style="height:200px;overflow:auto;border:1px solid #eee;padding:8px;background:#fafafa;margin:8px 0">
			<div v-for="(m, i) in messages" :key="i">{{ m.sender_name || m.sender_id }}: {{ m.content }}</div>
		</div>

		<input v-model="receiver" placeholder="Receiver user id" />
		<textarea v-model="content" placeholder="Message..."></textarea>
		<button @click="send">Send</button>
	</div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useChatStore } from '@/store/chat'

export default {
	setup() {
		const store = useChatStore()
		const messages = store.messages
		const receiver = ref('')
		const content = ref('')

		const connect = () => store.connect()
		const disconnect = () => store.disconnect()
		const send = () => {
			store.sendMessage({ type: 'message', receiver_id: receiver.value, content: content.value })
			content.value = ''
		}

		onMounted(() => {
			// auto-connect if user is present
			if (store.connected === false) {
				// don't auto-connect by default; user can click
			}
		})

		return { messages, receiver, content, connect, disconnect, send }
	}
}
</script>

<style scoped>
.page { max-width:720px; margin:24px auto }
textarea { width:100%; height:80px }
</style>

