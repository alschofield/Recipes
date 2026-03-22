import Link from 'next/link'
import { redirect } from 'next/navigation'
import { endpoints, serverPost } from '../../lib/server/api'
import { getSession, setSession } from '../../lib/server/session'

export default async function LoginPage({ searchParams }) {
  const qp = await searchParams
  const session = await getSession()
  if (session) {
    redirect('/recipes')
  }

  async function loginAction(formData) {
    'use server'
    const username = String(formData.get('username') || '')
    const password = String(formData.get('password') || '')
    const next = String(formData.get('next') || '/recipes')

    let destination
    try {
      const data = await serverPost(endpoints.users, '/users/login', { username, password })
      await setSession({
        token: data.token,
        id: data.id,
        username: data.username,
        email: data.email,
        role: data.role,
      })
      destination = next.startsWith('/') ? next : '/recipes'
    } catch (error) {
      const message = encodeURIComponent(error.message || 'Login failed')
      destination = `/login?error=${message}`
    }
    redirect(destination)
  }

  const error = qp?.error ? decodeURIComponent(qp.error) : ''
  const next = qp?.next || '/recipes'

  return (
    <div style={{ maxWidth: 400, margin: '2rem auto' }}>
      <h1 style={{ marginBottom: '1rem' }}>Login</h1>
      <form action={loginAction}>
        <input type="hidden" name="next" value={next} />
        <div style={{ marginBottom: '1rem' }}>
          <label htmlFor="login-username" style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.9rem' }}>Username</label>
          <input id="login-username" name="username" type="text" required />
        </div>
        <div style={{ marginBottom: '1rem' }}>
          <label htmlFor="login-password" style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.9rem' }}>Password</label>
          <input id="login-password" name="password" type="password" required />
        </div>
        {error && <p className="error-text" style={{ marginBottom: '0.75rem' }}>{error}</p>}
        <button type="submit" className="btn btn-primary" style={{ width: '100%' }}>Login</button>
      </form>
      <p style={{ marginTop: '1rem', fontSize: '0.9rem', textAlign: 'center' }}>
        Don&apos;t have an account? <Link href="/signup">Sign up</Link>
      </p>
    </div>
  )
}
