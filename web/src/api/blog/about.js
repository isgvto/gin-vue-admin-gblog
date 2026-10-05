import service from '@/utils/request'

const normalizeAboutData = (data) => {
  if (Array.isArray(data)) {
    return data.reduce((values, item = {}) => {
      const key = item.nameEn ?? item.name_en
      if (key) {
        values[key] = item.value ?? ''
      }
      return values
    }, {})
  }

  if (data && typeof data === 'object') {
    return data.values && typeof data.values === 'object' ? data.values : data
  }

  return {}
}

export function getAbout() {
  return service({
    url: '/admin/about',
    method: 'GET'
  }).then(res => {
    const values = normalizeAboutData(res.data)
    return {
      ...res,
      data: {
        title: values.title || '',
        content: values.content || '',
        commentEnabled: String(values.commentEnabled ?? 'true') === 'true'
      }
    }
  })
}

export function updateAbout(form) {
  return service({
    url: '/admin/about',
    method: 'PUT',
    data: {
      values: {
        title: String(form.title ?? ''),
        content: String(form.content ?? ''),
        commentEnabled: String(Boolean(form.commentEnabled))
      }
    }
  })
}
