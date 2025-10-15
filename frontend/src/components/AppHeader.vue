<template>
	<nav class="navbar navbar-expand-lg navbar-dark bg-primary shadow-sm">
		<div class="container-fluid">
			<router-link class="navbar-brand fw-bold text-white" to="/">
				<i class="fas fa-users me-2"></i>Social Network
			</router-link>

			<button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav">
				<span class="navbar-toggler-icon"></span>
			</button>

			<div class="collapse navbar-collapse" id="navbarNav">
				<ul class="navbar-nav me-auto" v-if="user">
					<li class="nav-item">
						<router-link class="nav-link text-white" to="/">
							<i class="fas fa-home me-1"></i>Home
						</router-link>
					</li>
					<li class="nav-item">
						<router-link class="nav-link text-white" to="/profile">
							<i class="fas fa-user me-1"></i>Profile
						</router-link>
					</li>
					<li class="nav-item">
						<router-link class="nav-link text-white" to="/people">
							<i class="fas fa-user-friends me-1"></i>People
						</router-link>
					</li>
					<li class="nav-item">
						<router-link class="nav-link text-white" to="/chat">
							<i class="fas fa-comments me-1"></i>Chat
						</router-link>
					</li>
					<li class="nav-item">
						<router-link class="nav-link text-white" to="/groups">
							<i class="fas fa-users-cog me-1"></i>Groups
						</router-link>
					</li>
				</ul>

				<div class="d-flex align-items-center" v-if="user">
					<!-- Notifications -->
					<div class="dropdown me-3">
						<button class="btn btn-outline-light position-relative" @click="toggleOpen" data-bs-toggle="dropdown">
							<i class="fas fa-bell"></i>
							<span v-if="unreadCount" class="position-absolute top-0 start-100 translate-middle badge rounded-pill bg-danger">
								{{ unreadCount }}
							</span>
						</button>
						<div class="dropdown-menu dropdown-menu-end p-0" style="width: 320px;" v-if="open">
							<div class="dropdown-header bg-light">
								<strong>Notifications</strong>
							</div>
							<div v-if="!notifications || notifications.length === 0" class="dropdown-item-text text-muted text-center py-3">
								No notifications
							</div>
							<div v-else>
								<div v-for="n in notifications" :key="n.id" 
									 class="dropdown-item border-bottom" 
									 :class="{ 'bg-light': !n.is_read }">
									<div class="d-flex justify-content-between align-items-start">
										<span class="fw-semibold text-primary">{{ n.type }}</span>
										<small class="text-muted">{{ n.created_at }}</small>
									</div>
									<div v-if="n.data" class="small text-muted mt-1">{{ n.data }}</div>
									<button v-if="!n.is_read" 
											@click.prevent="markRead(n.id)" 
											class="btn btn-sm btn-outline-primary mt-2">
										Mark read
									</button>
								</div>
								<div class="dropdown-footer p-2 border-top">
									<button class="btn btn-sm btn-primary w-100" @click="markAll">
										Mark all read
									</button>
								</div>
							</div>
						</div>
					</div>

					<!-- User Profile -->
					<div class="dropdown">
						<button class="btn btn-outline-light dropdown-toggle d-flex align-items-center" 
								data-bs-toggle="dropdown" aria-expanded="false">
							<img v-if="user.avatar" :src="user.avatar" alt="avatar" 
								 class="rounded-circle me-2" style="width: 24px; height: 24px;" />
							<i v-else class="fas fa-user-circle me-2"></i>
							{{ user.nickname || user.user_id || 'You' }}
						</button>
						<ul class="dropdown-menu dropdown-menu-end">
							<li><router-link class="dropdown-item" to="/profile">
								<i class="fas fa-user me-2"></i>My Profile
							</router-link></li>
							<li><router-link class="dropdown-item" to="/profile/edit">
								<i class="fas fa-edit me-2"></i>Edit Profile
							</router-link></li>
							<li><hr class="dropdown-divider"></li>
							<li><button class="dropdown-item text-danger" @click="onLogout">
								<i class="fas fa-sign-out-alt me-2"></i>Logout
							</button></li>
						</ul>
					</div>
				</div>

				<div class="d-flex" v-else>
					<router-link class="btn btn-outline-light me-2" to="/login">
						<i class="fas fa-sign-in-alt me-1"></i>Login
					</router-link>
					<router-link class="btn btn-warning" to="/register">
						<i class="fas fa-user-plus me-1"></i>Register
					</router-link>
				</div>
			</div>
		</div>
	</nav>
</template>

<script>
import { defineComponent, ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/store/auth'
import { useNotificationStore } from '@/store/notification'

export default defineComponent({
	setup() {
		const auth = useAuthStore()
		const notif = useNotificationStore()

		const { user } = storeToRefs(auth)
		const { list: notifications } = storeToRefs(notif)
		
		const unreadCount = computed(() => (notifications.value || []).filter(n => !n.is_read).length)

		const open = ref(false)
		const toggleOpen = async () => {
			open.value = !open.value
			if (open.value) await notif.fetch()
		}

		const markAll = async () => {
			await notif.markAllRead()
		}

		const markRead = async (id) => {
			await notif.markRead(id)
		}

		const onLogout = async () => {
			await auth.logout()
			location.href = '/'
		}

		return { user, notifications, unreadCount, open, toggleOpen, markAll, markRead, onLogout }
	}
})
</script>

<style scoped>
.navbar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
}

.navbar-brand {
  font-size: 1.5rem;
  font-weight: 700;
}

.nav-link {
  font-weight: 500;
  transition: color 0.3s ease;
}

.nav-link:hover {
  color: #ffc107 !important;
}

.dropdown-menu {
  border: none;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  border-radius: 0.5rem;
}

.dropdown-item:hover {
  background-color: #f8f9fa;
}

.btn-outline-light:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.badge {
  font-size: 0.6rem;
}
</style>

