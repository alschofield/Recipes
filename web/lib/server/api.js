const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || ''
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost'
const RECIPES_PORT = process.env.NEXT_PUBLIC_API_RECIPES_PORT || '8081'
const USERS_PORT = process.env.NEXT_PUBLIC_API_USERS_PORT || '8082'
const FAVORITES_PORT = process.env.NEXT_PUBLIC_API_FAVORITES_PORT || '8080'
const RECIPES_URL = process.env.NEXT_PUBLIC_API_RECIPES_URL || ''
const USERS_URL = process.env.NEXT_PUBLIC_API_USERS_URL || ''
const FAVORITES_URL = process.env.NEXT_PUBLIC_API_FAVORITES_URL || ''

function trimSlash(value) {
  return value.replace(/\/$/, '')
}

const sharedBase = API_BASE_URL ? trimSlash(API_BASE_URL) : ''

function endpoint(explicitURL, fallbackPort) {
  if (sharedBase) return sharedBase
  if (explicitURL) return trimSlash(explicitURL)
  return `${API_URL}:${fallbackPort}`
}

export const endpoints = {
  recipes: endpoint(RECIPES_URL, RECIPES_PORT),
  users: endpoint(USERS_URL, USERS_PORT),
  favorites: endpoint(FAVORITES_URL, FAVORITES_PORT),
}

export async function serverGet(baseURL, path, token) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers.Authorization = `Bearer ${token}`
  const res = await fetch(`${baseURL}${path}`, { headers, cache: 'no-store' })
  if (res.status === 204) return null
  const data = await res.json()
  if (!res.ok) {
    const error = new Error(data.error || 'Request failed')
    error.status = res.status
    throw error
  }
  return data
}

export async function serverPost(baseURL, path, body, token) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers.Authorization = `Bearer ${token}`
  const res = await fetch(`${baseURL}${path}`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
    cache: 'no-store',
  })
  if (res.status === 204) return null
  const data = await res.json()
  if (!res.ok) {
    const error = new Error(data.error || 'Request failed')
    error.status = res.status
    throw error
  }
  return data
}

export async function serverPut(baseURL, path, body, token) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers.Authorization = `Bearer ${token}`
  const res = await fetch(`${baseURL}${path}`, {
    method: 'PUT',
    headers,
    body: JSON.stringify(body),
    cache: 'no-store',
  })
  if (res.status === 204) return null
  const data = await res.json()
  if (!res.ok) {
    const error = new Error(data.error || 'Request failed')
    error.status = res.status
    throw error
  }
  return data
}

export async function serverDelete(baseURL, path, token) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers.Authorization = `Bearer ${token}`
  const res = await fetch(`${baseURL}${path}`, { method: 'DELETE', headers, cache: 'no-store' })
  if (res.status === 204) return null
  const data = await res.json()
  if (!res.ok) {
    const error = new Error(data.error || 'Request failed')
    error.status = res.status
    throw error
  }
  return data
}
