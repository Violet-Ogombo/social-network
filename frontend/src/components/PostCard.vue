<template>
  <div class="card post-card">
    <!-- Post Header -->
    <div class="card-header d-flex align-items-center justify-content-between">
      <div class="d-flex align-items-center">
        <div class="user-avatar me-3">
          <i class="fas fa-user-circle fa-2x text-primary"></i>
        </div>
        <div>
          <h6 class="mb-0 fw-semibold">{{ post.nickname || 'Anonymous User' }}</h6>
          <small class="text-muted">
            <i class="far fa-clock me-1"></i>
            {{ formatDate(post.created_at) }}
          </small>
        </div>
      </div>
      <div class="dropdown">
        <button class="btn btn-sm btn-outline-secondary dropdown-toggle" type="button" data-bs-toggle="dropdown">
          <i class="fas fa-ellipsis-h"></i>
        </button>
        <ul class="dropdown-menu">
          <li><a class="dropdown-item" href="#"><i class="fas fa-share me-2"></i>Share</a></li>
          <li><a class="dropdown-item" href="#"><i class="fas fa-bookmark me-2"></i>Save</a></li>
          <li><hr class="dropdown-divider"></li>
          <li><a class="dropdown-item text-danger" href="#"><i class="fas fa-flag me-2"></i>Report</a></li>
        </ul>
      </div>
    </div>

    <!-- Post Content -->
    <div class="card-body">
      <p class="card-text post-content">{{ post.content }}</p>
      
      <!-- Post Image -->
      <div v-if="post.image_url" class="post-image mb-3">
        <img :src="post.image_url" class="img-fluid rounded" alt="Post image" />
      </div>

      <!-- Post Actions -->
      <div class="post-actions d-flex align-items-center justify-content-between mb-3">
        <div class="d-flex gap-3">
          <button class="btn btn-sm btn-outline-primary" @click="toggleLike">
            <i :class="['fas', isLiked ? 'fa-heart text-danger' : 'fa-heart']"></i>
            <span class="ms-1">{{ likeCount }}</span>
          </button>
          <button class="btn btn-sm btn-outline-secondary" @click="showComments = !showComments">
            <i class="fas fa-comment"></i>
            <span class="ms-1">{{ commentCount }} Comments</span>
          </button>
          <button class="btn btn-sm btn-outline-info">
            <i class="fas fa-share"></i>
            <span class="ms-1">Share</span>
          </button>
        </div>
        <div class="privacy-indicator">
          <span class="badge" :class="getPrivacyClass(post.privacy)">
            <i :class="getPrivacyIcon(post.privacy)"></i>
            {{ post.privacy }}
          </span>
        </div>
      </div>

      <!-- Comments Section -->
      <div v-if="showComments" class="comments-section">
        <hr class="my-3">
        <Comment :postId="post.id" @comment-added="onCommentAdded" />
      </div>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import * as api from '@/api/post'
import Comment from './Comment.vue'
export default {
  props: ['post'],
  components: { Comment },
  emits: ['comment-added'],
  setup(props, { emit }) {
    const showComments = ref(false)
    const isLiked = ref(false)
    const likeCount = ref(Math.floor(Math.random() * 50)) // Mock data
    const commentCount = ref(Math.floor(Math.random() * 10)) // Mock data

    const onCommentAdded = () => {
      commentCount.value++
      emit('comment-added')
    }

    const toggleLike = () => {
      isLiked.value = !isLiked.value
      likeCount.value += isLiked.value ? 1 : -1
    }

    const formatDate = (dateString) => {
      if (!dateString) return 'Just now'
      const date = new Date(dateString)
      const now = new Date()
      const diffMs = now - date
      const diffMins = Math.floor(diffMs / 60000)
      const diffHours = Math.floor(diffMs / 3600000)
      const diffDays = Math.floor(diffMs / 86400000)

      if (diffMins < 1) return 'Just now'
      if (diffMins < 60) return `${diffMins}m ago`
      if (diffHours < 24) return `${diffHours}h ago`
      if (diffDays < 7) return `${diffDays}d ago`
      return date.toLocaleDateString()
    }

    const getPrivacyClass = (privacy) => {
      switch(privacy) {
        case 'public': return 'bg-success'
        case 'followers': return 'bg-warning'
        case 'private': return 'bg-danger'
        default: return 'bg-secondary'
      }
    }

    const getPrivacyIcon = (privacy) => {
      switch(privacy) {
        case 'public': return 'fas fa-globe'
        case 'followers': return 'fas fa-users'
        case 'private': return 'fas fa-lock'
        default: return 'fas fa-question'
      }
    }

    return { 
      showComments, 
      isLiked, 
      likeCount, 
      commentCount,
      onCommentAdded, 
      toggleLike, 
      formatDate, 
      getPrivacyClass, 
      getPrivacyIcon 
    }
  }
}
</script>

<style scoped>
.post-card {
  border: none;
  border-radius: 1rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  overflow: hidden;
  background: white;
}

.post-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.12);
}

.card-header {
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  padding: 1rem 1.25rem;
}

.user-avatar {
  flex-shrink: 0;
}

.post-content {
  font-size: 1rem;
  line-height: 1.6;
  color: #2d3748;
  margin-bottom: 1rem;
  word-wrap: break-word;
}

.post-image {
  border-radius: 0.75rem;
  overflow: hidden;
}

.post-image img {
  border-radius: 0.75rem;
  cursor: pointer;
  transition: transform 0.3s ease;
}

.post-image img:hover {
  transform: scale(1.02);
}

.post-actions {
  border-top: 1px solid rgba(0, 0, 0, 0.05);
  padding-top: 0.75rem;
}

.post-actions .btn {
  border-radius: 0.5rem;
  font-weight: 500;
  padding: 0.375rem 0.75rem;
  transition: all 0.2s ease;
}

.post-actions .btn:hover {
  transform: translateY(-1px);
}

.post-actions .btn-outline-primary:hover {
  background-color: #667eea;
  border-color: #667eea;
}

.post-actions .btn-outline-secondary:hover {
  background-color: #6c757d;
  border-color: #6c757d;
}

.post-actions .btn-outline-info:hover {
  background-color: #17a2b8;
  border-color: #17a2b8;
}

.privacy-indicator .badge {
  font-size: 0.75rem;
  padding: 0.375rem 0.5rem;
  border-radius: 0.5rem;
}

.comments-section {
  background-color: #f8f9fa;
  margin: 0 -1.25rem -1.25rem;
  padding: 1rem 1.25rem 1.25rem;
  border-radius: 0 0 1rem 1rem;
}

.dropdown-menu {
  border: none;
  border-radius: 0.75rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.dropdown-item {
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
  margin: 0.125rem 0.5rem;
  transition: all 0.2s ease;
}

.dropdown-item:hover {
  background-color: #f8f9fa;
  transform: translateX(2px);
}

.text-danger {
  color: #dc3545 !important;
}

.fa-heart.text-danger {
  color: #e91e63 !important;
  animation: heartbeat 1s ease-in-out;
}

@keyframes heartbeat {
  0% { transform: scale(1); }
  50% { transform: scale(1.2); }
  100% { transform: scale(1); }
}
</style>
