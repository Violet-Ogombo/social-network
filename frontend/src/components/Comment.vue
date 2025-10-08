<template>
	<div class="comment-box">
		<input v-model="text" placeholder="Write a comment..." class="form-control" />
		<button class="btn btn-sm btn-primary mt-1" @click="submit">Comment</button>
	</div>
</template>

<script>
import { ref } from 'vue'
import * as api from '@/api/post'

export default {
	props: { postId: { type: [Number, String], required: true } },
	emits: ['comment-added'],
	setup(props, { emit }) {
		const text = ref('')
		const submit = async () => {
			if (!text.value) return
			try {
				await api.addComment(props.postId, text.value)
				text.value = ''
				emit('comment-added')
			} catch (e) {
				// swallow for now
				console.error('Failed to add comment', e)
			}
		}
		return { text, submit }
	}
}
</script>

<style scoped>
.comment-box input { display:block; margin-bottom:6px }
</style>
