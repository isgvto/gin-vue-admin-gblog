import axios from '@/plugins/axios'

const owner = 'LZMclear'
const repo = 'gin-vue-admin-gblog'
const branch = 'main'

export function getRepositoryCommits(page = 1, perPage = 20) {
	return axios({
		url: `https://api.github.com/repos/${owner}/${repo}/commits`,
		method: 'GET',
		params: {
			sha: branch,
			page,
			per_page: perPage
		}
	})
}

