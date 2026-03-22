import Link from 'next/link'
import { redirect } from 'next/navigation'
import { endpoints, serverPost } from '../../lib/server/api'
import { getSession, setSession } from '../../lib/server/session'

export default async function SignupPage({ searchParams }) {
  const qp = await searchParams
  const session = await getSession()
  if (session) {
    redirect('/recipes')
  }

  async function signupAction(formData) {
    'use server'
    const username = String(formData.get('username') || '')
    const email = String(formData.get('email') || '')
    const password = String(formData.get('password') || '')

    let destination
    try {
      await serverPost(endpoints.users, '/users/new', { username, email, password })
      const data = await serverPost(endpoints.users, '/users/login', { username, password })
      await setSession({
        token: data.token,
        id: data.id,
        username: data.username,
        email: data.email,
        role: data.role,
      })
      destination = '/recipes'
    } catch (error) {
      const message = encodeURIComponent(error.message || 'Signup failed')
      destination = `/signup?error=${message}`
    }
    redirect(destination)
  }

  const error = qp?.error ? decodeURIComponent(qp.error) : ''

  return (
    <div style={{ maxWidth: 400, margin: '2rem auto' }}>
      <h1 style={{ marginBottom: '1rem' }}>Create Account</h1>
      <form action={signupAction}>
        <div style={{ marginBottom: '1rem' }}>
          <label htmlFor="signup-username" style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.9rem' }}>Username</label>
          <input id="signup-username" name="username" type="text" required />
        </div>
        <div style={{ marginBottom: '1rem' }}>
          <label htmlFor="signup-email" style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.9rem' }}>Email</label>
          <input id="signup-email" name="email" type="email" required />
        </div>
        <div style={{ marginBottom: '1rem' }}>
          <label htmlFor="signup-password" style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.9rem' }}>Password</label>
          <input id="signup-password" name="password" type="password" required minLength={12} />
          <span className="muted" style={{ fontSize: '0.8rem' }}>Min 12 characters</span>
        </div>
        {error && <p className="error-text" style={{ marginBottom: '0.75rem' }}>{error}</p>}
        <button type="submit" className="btn btn-primary" style={{ width: '100%' }}>Create Account</button>
      </form>
      <p style={{ marginTop: '1rem', fontSize: '0.9rem', textAlign: 'center' }}>
        Already have an account? <Link href="/login">Login</Link>
      </p>
    </div>
  )
}
