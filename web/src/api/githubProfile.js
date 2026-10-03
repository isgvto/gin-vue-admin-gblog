import service from '@/utils/request'

export const getGitHubProfile = () => service({
  url: '/user/github', method: 'GET', timeout: 16000,
  donNotShowLoading: true, silentError: true
})
export const setGitHubProfile = username => service({
  url: '/user/github', method: 'PUT', data: { username },
  donNotShowLoading: true, silentError: true
})
