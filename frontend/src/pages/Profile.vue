<template>
  <div class="container py-4">
    <div v-if="profile" class="profile-container">
      <!-- Profile Header -->
      <div class="row">
        <div class="col-12">
          <div class="card profile-header-card">
            <div class="profile-cover"></div>
            <div class="card-body position-relative">
              <!-- Profile Avatar -->
              <div class="profile-avatar-container">
                <div class="profile-avatar">
                  <img v-if="profile.avatar" :src="profile.avatar" alt="avatar" class="avatar-img" />
                  <i v-else class="fas fa-user-circle fa-5x text-primary"></i>
                </div>
                <div v-if="isMine" class="edit-avatar-btn">
                  <button class="btn btn-sm btn-outline-primary">
                    <i class="fas fa-camera"></i>
                  </button>
                </div>
              </div>

              <!-- Profile Info -->
              <div class="profile-info mt-4">
                <div class="row align-items-start">
                  <div class="col-md-8">
                    <h2 class="profile-name mb-1">{{ profile.nickname || 'Anonymous User' }}</h2>
                    <h5 class="text-muted mb-2">{{ profile.first_name }} {{ profile.last_name }}</h5>
                    <p class="profile-about mb-3" v-if="profile.about">{{ profile.about }}</p>
                    
                    <!-- Profile Stats -->
                    <div class="profile-stats d-flex gap-4 mb-3">
                      <div class="stat-item">
                        <strong class="d-block text-primary">{{ followers.length }}</strong>
                        <small class="text-muted">Followers</small>
                      </div>
                      <div class="stat-item">
                        <strong class="d-block text-primary">{{ followingList.length }}</strong>
                        <small class="text-muted">Following</small>
                      </div>
                      <div class="stat-item">
                        <strong class="d-block text-primary">{{ postCount }}</strong>
                        <small class="text-muted">Posts</small>
                      </div>
                    </div>

                    <!-- Contact Info -->
                    <div v-if="profile.email && (isMine || profile.profile_type === 'public')" class="contact-info mb-3">
                      <small class="text-muted">
                        <i class="fas fa-envelope me-2"></i>
                        {{ profile.email }}
                      </small>
                    </div>
                  </div>

                  <div class="col-md-4 text-md-end">
                    <!-- Privacy Settings (Own Profile) -->
                    <div v-if="isMine" class="privacy-controls mb-3">
                      <label class="form-label fw-semibold">
                        <i class="fas fa-shield-alt text-primary me-2"></i>
                        Profile Privacy
                      </label>
                      <select v-model="profile.profile_type" @change="changePrivacy" class="form-select">
                        <option value="public">
                          <i class="fas fa-globe"></i> Public
                        </option>
                        <option value="private">
                          <i class="fas fa-lock"></i> Private
                        </option>
                      </select>
                    </div>

                    <!-- Follow Controls (Other Profiles) -->
                    <div v-else class="follow-controls">
                      <button 
                        @click="toggleFollow" 
                        :class="following ? 'btn btn-outline-danger' : 'btn btn-primary'"
                        class="btn-lg px-4"
                      >
                        <i :class="following ? 'fas fa-user-minus' : 'fas fa-user-plus'"></i>
                        {{ following ? 'Unfollow' : 'Follow' }}
                      </button>
                    </div>

                    <!-- Privacy Badge -->
                    <div class="privacy-badge">
                      <span class="badge" :class="profile.profile_type === 'public' ? 'bg-success' : 'bg-warning'">
                        <i :class="profile.profile_type === 'public' ? 'fas fa-globe' : 'fas fa-lock'"></i>
                        {{ profile.profile_type }} Profile
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Profile Content Tabs -->
      <div class="row mt-4">
        <div class="col-md-6">
          <div class="card followers-card">
            <div class="card-header bg-gradient">
              <h5 class="mb-0 text-white">
                <i class="fas fa-users me-2"></i>
                Followers ({{ followers.length }})
              </h5>
            </div>
            <div class="card-body">
              <div v-if="followers.length === 0" class="text-center py-3">
                <i class="fas fa-user-friends fa-2x text-muted mb-2"></i>
                <p class="text-muted mb-0">No followers yet</p>
              </div>
              <div v-else class="followers-list">
                <div v-for="f in followers" :key="f.id" class="follower-item d-flex align-items-center mb-2">
                  <i class="fas fa-user-circle fa-2x text-primary me-3"></i>
                  <div>
                    <strong class="d-block">{{ f.nickname }}</strong>
                    <small class="text-muted">{{ f.first_name }} {{ f.last_name }}</small>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="col-md-6">
          <div class="card following-card">
            <div class="card-header bg-gradient">
              <h5 class="mb-0 text-white">
                <i class="fas fa-heart me-2"></i>
                Following ({{ followingList.length }})
              </h5>
            </div>
            <div class="card-body">
              <div v-if="followingList.length === 0" class="text-center py-3">
                <i class="fas fa-heart fa-2x text-muted mb-2"></i>
                <p class="text-muted mb-0">Not following anyone yet</p>
              </div>
              <div v-else class="following-list">
                <div v-for="f in followingList" :key="f.id" class="following-item d-flex align-items-center mb-2">
                  <i class="fas fa-user-circle fa-2x text-success me-3"></i>
                  <div>
                    <strong class="d-block">{{ f.nickname }}</strong>
                    <small class="text-muted">{{ f.first_name }} {{ f.last_name }}</small>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-else class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
      <p class="text-muted mt-2">Loading profile...</p>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue'
import * as api from '@/api/users'
import { follow, unfollow } from '@/api/users'
import { useAuthStore } from '@/store/auth'

export default {
  setup() {
    const auth = useAuthStore()
    const profile = ref(null)
    const followers = ref([])
    const followingList = ref([])
    const following = ref(false)
    const postCount = ref(Math.floor(Math.random() * 100)) // Mock data

    const isMine = computed(() => auth.user && profile.value && String(auth.user.user_id) === String(profile.value.id))

    const load = async () => {
      // load own profile by default
      profile.value = await api.getProfile()
      followers.value = await api.getFollowers()
      followingList.value = await api.getFollowing()
    }

    const changePrivacy = async () => {
      if (!isMine.value) return
      await api.setPrivacy(profile.value.profile_type)
    }

    const toggleFollow = async () => {
      if (!profile.value) return
      if (following.value) {
        await unfollow(profile.value.id)
        following.value = false
      } else {
        await follow(profile.value.id)
        following.value = true
      }
      followers.value = await api.getFollowers(profile.value.id)
    }

    onMounted(load)

    return { profile, followers, followingList, following, postCount, isMine, changePrivacy, toggleFollow }
  }
}
</script>

<style scoped>
.profile-container {
  max-width: 1000px;
  margin: 0 auto;
}

.profile-header-card {
  border: none;
  border-radius: 1rem;
  overflow: hidden;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.profile-cover {
  height: 200px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
}

.profile-avatar-container {
  position: absolute;
  top: -60px;
  left: 30px;
}

.profile-avatar {
  position: relative;
  width: 120px;
  height: 120px;
  border-radius: 50%;
  border: 4px solid white;
  overflow: hidden;
  background: white;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}

.edit-avatar-btn {
  position: absolute;
  bottom: 5px;
  right: 5px;
}

.profile-info {
  margin-left: 150px;
}

.profile-name {
  color: #2d3748;
  font-weight: 700;
  font-size: 1.75rem;
}

.profile-about {
  color: #4a5568;
  font-size: 1rem;
  line-height: 1.5;
}

.profile-stats {
  padding: 1rem 0;
  border-top: 1px solid #e2e8f0;
  border-bottom: 1px solid #e2e8f0;
}

.stat-item {
  text-align: center;
}

.privacy-controls .form-select {
  max-width: 200px;
  border-radius: 0.75rem;
}

.follow-controls .btn {
  border-radius: 0.75rem;
  font-weight: 600;
  transition: all 0.3s ease;
}

.follow-controls .btn:hover {
  transform: translateY(-2px);
}

.privacy-badge {
  margin-top: 1rem;
}

.privacy-badge .badge {
  font-size: 0.8rem;
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
}

.card {
  border: none;
  border-radius: 1rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
}

.card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.12);
}

.card-header.bg-gradient {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  border: none;
}

.follower-item, .following-item {
  padding: 0.75rem;
  border-radius: 0.75rem;
  transition: all 0.2s ease;
  cursor: pointer;
}

.follower-item:hover, .following-item:hover {
  background-color: #f8f9fa;
  transform: translateX(4px);
}

.contact-info {
  padding: 0.75rem;
  background-color: #f8f9fa;
  border-radius: 0.5rem;
  display: inline-block;
}

@media (max-width: 768px) {
  .profile-info {
    margin-left: 0;
    margin-top: 70px;
  }
  
  .profile-avatar-container {
    left: 50%;
    transform: translateX(-50%);
  }
}
</style>
