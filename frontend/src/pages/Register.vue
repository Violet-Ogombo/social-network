<template>
	<div class="page card">
		<h2>Register</h2>
		<input v-model="email" placeholder="Email" />
		<input v-model="password" type="password" placeholder="Password" />
		<input v-model="first_name" placeholder="First name" />
		<input v-model="last_name" placeholder="Last name" />
		<button @click="onRegister">Register</button>
		<div class="msg" v-if="msg">{{ msg }}</div>
	</div>
</template>

<script>
import { ref } from 'vue'
import { useAuthStore } from '@/store/auth'
import { useRouter } from 'vue-router'

export default {
	setup() {
		const email = ref('')
		const password = ref('')
		const first_name = ref('')
		const last_name = ref('')
		const msg = ref('')
		const auth = useAuthStore()
		const router = useRouter()

		const onRegister = async () => {
			try {
				await auth.register({ email: email.value, password: password.value, first_name: first_name.value, last_name: last_name.value })
				// fetch user and go home
				await auth.fetchUser()
				router.push('/')
			} catch (e) {
				msg.value = e?.response?.data?.error || 'Registration failed'
			}
		}

		return { email, password, first_name, last_name, onRegister, msg }
	}
}
</script>

<style scoped>
.page { max-width:480px; margin:24px auto }
input { display:block; margin:8px 0; padding:8px }
</style>

