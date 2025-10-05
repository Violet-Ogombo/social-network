<template>
	<div class="page card">
		<h2>Login</h2>
		<input v-model="identifier" placeholder="Email or nickname" />
		<input v-model="password" type="password" placeholder="Password" />
		<button @click="onLogin">Login</button>
		<div class="msg" v-if="msg">{{ msg }}</div>
	</div>
</template>

<script>
import { ref } from 'vue'
import { useAuthStore } from '@/store/auth'
import { useRouter } from 'vue-router'

export default {
	setup() {
		const identifier = ref('')
		const password = ref('')
		const msg = ref('')
		const auth = useAuthStore()
		const router = useRouter()

		const onLogin = async () => {
			try {
				await auth.login(identifier.value, password.value)
				router.push('/')
			} catch (e) {
				msg.value = e?.response?.data || 'Login failed'
			}
		}

		return { identifier, password, onLogin, msg }
	}
}
</script>

<style scoped>
.page { max-width:480px; margin:24px auto }
input { display:block; margin:8px 0; padding:8px }
</style>

