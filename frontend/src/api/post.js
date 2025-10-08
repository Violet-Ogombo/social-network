import axios from 'axios';

const apiClient = axios.create({
  baseURL: 'http://localhost:8080',
  withCredentials: true,
});

export const createPost = async (formData) => {
  const res = await apiClient.post('/api/posts/create', formData, { headers: { 'Content-Type': 'multipart/form-data' } });
  return res.data;
}

export const listPosts = async (user_id) => {
  const url = user_id ? `/api/posts?user_id=${user_id}` : '/api/posts';
  const res = await apiClient.get(url);
  return res.data;
}

export const addComment = async (post_id, content) => {
  const res = await apiClient.post('/api/posts/comment', { post_id, content });
  return res.data;
}
