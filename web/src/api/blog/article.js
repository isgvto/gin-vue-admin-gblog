import service from '@/utils/request'
import {
  mapArticleOut,
  mapToggleValue,
  normalizePageQuery
} from './_helpers'

export function getArticlePage(queryInfo) {
  return service({
      url: '/admin/blogs',
      method: 'GET',
      params: normalizePageQuery(queryInfo)
  })
}

export function getDataByQuery(queryInfo) {
  return Promise.all([
    getArticlePage(queryInfo),
    getCategoryAndTag()
  ]).then(([res, metaRes]) => ({
    ...res,
    data: {
      blogs: {
        ...(res.data || {}),
        list: res.data?.list || []
      },
      categories: metaRes.data.categories
    }
  }))
}

export function deleteBlogById(id) {
  return service({
    url: '/admin/blog',
    method: 'DELETE',
    data: { id }
  })
}

export function getCategoryAndTag() {
  return service({
    url: '/admin/categoryAndTag',
    method: 'GET'
  }).then(res => ({
    ...res,
    data: {
      categories: res.data?.categories || [],
      tags: res.data?.tags || []
    }
  }))
}

export function saveBlog(blog) {
  return service({
    url: '/admin/blog',
    method: 'POST',
    data: mapArticleOut(blog)
  })
}

export function updateTop(id, top) {
  return service({
    url: '/admin/blog/top',
    method: 'PUT',
    params: mapToggleValue(id, top)
  })
}

export function updateRecommend(id, recommend) {
  return service({
    url: '/admin/blog/recommend',
    method: 'PUT',
    params: mapToggleValue(id, recommend)
  })
}

export function updateVisibility(id, form) {
  return service({
    url: `/admin/blog/${id}/visibility`,
    method: 'PUT',
    data: {
      appreciation: form.appreciation,
      recommend: form.recommend,
      commentEnabled: form.commentEnabled,
      top: form.top,
      published: form.published,
      password: form.password
    }
  })
}

export function getBlogById(id) {
  return service({
    url: '/admin/blog',
    method: 'GET',
    params: { id }
  })
}

export function updateBlog(blog) {
  return service({
    url: '/admin/blog',
    method: 'PUT',
    data: mapArticleOut(blog)
  })
}
