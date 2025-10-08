<template>
  <div class="container mt-4">
    <h2>{{ group.group.name || 'Group' }}</h2>
    <p v-if="group.group.description">{{ group.group.description }}</p>
    <div class="mb-3">
      <h5>Members: {{ group.members }}</h5>
    </div>

    <div class="card mb-3">
      <div class="card-body">
        <h5>Create Post</h5>
        <textarea v-model="content" class="form-control" placeholder="What's on your mind?"></textarea>
        <input type="file" ref="file" class="form-control mt-2" />
        <button class="btn btn-primary mt-2" @click="createPost">Post</button>
      </div>
    </div>

    <div v-for="p in posts" :key="p.id" class="card mb-2">
      <div class="card-body">
        <p>{{ p.content }}</p>
        <img v-if="p.image_url" :src="p.image_url" class="img-fluid" />
        <div class="mt-2">
          <Comment :postId="p.id" @comment-added="load" />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { getGroup, listGroupPosts, createGroupPost, addGroupComment } from '../api/groups'
import Comment from '@/components/Comment.vue'

export default {
  data() { return { group: {}, posts: [], content: '', commentText: {} } },
  created() { this.load() },
  methods: {
    load() {
      const id = this.$route.params.id
      getGroup(id).then(r => { this.group = r.data })
      listGroupPosts(id).then(r => { this.posts = r.data })
    },
    createPost() {
      const fd = new FormData()
      fd.append('group_id', this.$route.params.id)
      fd.append('content', this.content)
      const f = this.$refs.file && this.$refs.file.files && this.$refs.file.files[0]
      if (f) fd.append('image', f)
      createGroupPost(fd).then(()=>{ this.content=''; this.load() })
    },
    addComment(postId) {
      addGroupComment({ post_id: postId, content: this.commentText[postId] }).then(()=> { this.commentText[postId] = ''; this.load() })
    }
  }
}
</script>

<style scoped>
</style>
