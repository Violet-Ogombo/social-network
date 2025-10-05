<template>
	<header class="app-header">
		<div class="left">
			<h1 class="logo">Social Network</h1>
		</div>
		<div class="right">
			<template v-if="user">
				<img v-if="user.avatar" :src="user.avatar" alt="avatar" class="avatar" />
				<span class="nickname">{{ user.nickname || user.user_id || 'You' }}</span>
				<button @click="onLogout">Logout</button>
			</template>
			<template v-else>
				<router-link to="/login">Login</router-link>
				<router-link to="/register">Register</router-link>
			</template>
		</div>
	</header>
</template>

<script>
import { useAuthStore } from '@/store/auth'
import { defineComponent } from 'vue'

export default defineComponent({
	setup() {
		const auth = useAuthStore()
		const user = auth.user
		const onLogout = async () => {
			await auth.logout()
			// navigate to home
			location.href = '/'
		}
		return { user, onLogout }
	}
})
</script>

<style scoped>
.app-header { display:flex; justify-content:space-between; align-items:center; padding:12px 16px; background:#fff }
.avatar { width:32px; height:32px; border-radius:50%; margin-right:8px }
.nickname { margin-right:12px }
</style>

