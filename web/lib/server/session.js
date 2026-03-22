import { cookies } from 'next/headers'

export const SESSION_COOKIE = 'recipes_token'
export const USER_ID_COOKIE = 'recipes_user_id'
export const USERNAME_COOKIE = 'recipes_username'
export const EMAIL_COOKIE = 'recipes_email'
export const ROLE_COOKIE = 'recipes_role'

export async function getSession() {
  const store = await cookies()
  const token = store.get(SESSION_COOKIE)?.value
  if (!token) return null

  return {
    token,
    id: store.get(USER_ID_COOKIE)?.value || '',
    username: store.get(USERNAME_COOKIE)?.value || '',
    email: store.get(EMAIL_COOKIE)?.value || '',
    role: store.get(ROLE_COOKIE)?.value || 'user',
  }
}

export async function setSession(session) {
  const store = await cookies()
  const common = {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
    maxAge: 60 * 60 * 24 * 7,
  }

  store.set(SESSION_COOKIE, session.token, common)
  store.set(USER_ID_COOKIE, String(session.id || ''), common)
  store.set(USERNAME_COOKIE, String(session.username || ''), common)
  store.set(EMAIL_COOKIE, String(session.email || ''), common)
  store.set(ROLE_COOKIE, String(session.role || 'user'), common)
}

export async function clearSession() {
  const store = await cookies()
  for (const key of [SESSION_COOKIE, USER_ID_COOKIE, USERNAME_COOKIE, EMAIL_COOKIE, ROLE_COOKIE]) {
    store.set(key, '', {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      path: '/',
      maxAge: 0,
    })
  }
}
