/**
 * 通用 API 请求工具
 * - 自动附加 CAS/Token 认证头
 * - 自动检测 401 响应并跳转登录页
 */
export async function apiFetch(url, options = {}) {
  const token = localStorage.getItem('pingmesh_token')
  const legacy = localStorage.getItem('pingmesh_user')

  const headers = { ...options.headers }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  } else if (legacy) {
    headers['Authorization'] = `Bearer legacy-${legacy}`
  }

  const response = await fetch(url, { ...options, headers })

  // 自动检测 401：克隆响应读取 body，不影响调用方后续读取
  const cloned = response.clone()
  try {
    const body = await cloned.json()
    if (body && body.code === 401) {
      localStorage.removeItem('pingmesh_token')
      localStorage.removeItem('pingmesh_username')
      localStorage.removeItem('pingmesh_display_name')
      localStorage.removeItem('pingmesh_logged_in')
      localStorage.removeItem('pingmesh_user')
      window.location.href = '/#/login'
    }
  } catch {
    // 非 JSON 响应，忽略
  }

  return response
}
