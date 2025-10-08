import api from './index'

export function listGroups() {
  return api.get('/api/groups')
}

export function createGroup(data) {
  return api.post('/api/group/create', data)
}

export function getGroup(id) {
  return api.get('/api/group?id=' + id)
}

export function inviteToGroup(payload) {
  return api.post('/api/group/invite', payload)
}

export function respondInvite(payload) {
  return api.post('/api/group/invite/respond', payload)
}

export function listGroupPosts(group_id) {
  return api.get('/api/group/posts?group_id=' + group_id)
}

export function createGroupPost(formData) {
  return api.post('/api/group/post/create', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
}

export function addGroupComment(payload) {
  return api.post('/api/group/comment', payload)
}

export function createEvent(payload) {
  return api.post('/api/group/event/create', payload)
}

export function voteEvent(payload) {
  return api.post('/api/group/event/vote', payload)
}
