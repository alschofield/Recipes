import { redirect } from 'next/navigation'
import { endpoints, serverPut } from '../../lib/server/api'
import { clearSession, getSession, setSession } from '../../lib/server/session'

function isAuthError(error) {
  if (error?.status === 401) return true
  const text = String(error?.message || '').toLowerCase()
  return text.includes('invalid token') || text.includes('expired token') || text.includes('authorization header') || text.includes('unauthorized')
}

export default async function AccountPage({ searchParams }) {
  const qp = await searchParams
  const session = await getSession()
  if (!session) redirect('/login?next=/account')

  async function updateAccountAction(formData) {
    'use server'
    const auth = await getSession()
    if (!auth) redirect('/login?next=/account')

    const username = String(formData.get('username') || '')
    const email = String(formData.get('email') || '')
    const currentPassword = String(formData.get('currentPassword') || '')
    const newPassword = String(formData.get('newPassword') || '')

    const payload = {}
    if (username && username !== auth.username) payload.username = username
    if (email && email !== auth.email) payload.email = email
    if (newPassword) {
      payload.currentPassword = currentPassword
      payload.newPassword = newPassword
    }

    if (Object.keys(payload).length === 0) {
      redirect('/account?error=No%20changes%20to%20save')
    }

    let destination
    try {
      const data = await serverPut(endpoints.users, `/users/${auth.id}`, payload, auth.token)
      await setSession({
        token: auth.token,
        id: data.id || auth.id,
        username: data.username,
        email: data.email,
        role: data.role || auth.role,
      })
      destination = '/account?ok=Profile%20updated'
    } catch (e) {
      if (isAuthError(e)) {
        await clearSession()
        redirect('/login?next=/account')
      }
      if (e?.status === 403) {
        redirect('/recipes')
      }
      destination = `/account?error=${encodeURIComponent(e.message || 'Update failed')}`
    }
    redirect(destination)
  }

  async function logoutAction() {
    'use server'
    await clearSession()
    redirect('/')
  }

  const error = qp?.error ? decodeURIComponent(qp.error) : ''
  const ok = qp?.ok ? decodeURIComponent(qp.ok) : ''

  return (
    <div style={{ maxWidth: 500, margin: '0 auto' }}>
      <div className="page-header">
        <h1>Profile</h1>
        <p className="muted">Manage your Ingrediential identity, credentials, and account access.</p>
      </div>

      <div className="card" style={{ marginBottom: '1.5rem' }}>
        <h3 style={{ marginBottom: '1rem' }}>Profile details</h3>
        <form action={updateAccountAction}>
          <div style={{ marginBottom: '1rem' }}>
            <label htmlFor="account-username" style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.9rem' }}>Username</label>
            <input id="account-username" name="username" type="text" defaultValue={session.username} />
          </div>
          <div style={{ marginBottom: '1rem' }}>
            <label htmlFor="account-email" style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.9rem' }}>Email</label>
            <input id="account-email" name="email" type="email" defaultValue={session.email} />
          </div>
          <div style={{ marginBottom: '1rem' }}>
            <label htmlFor="account-current-password" style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.9rem' }}>Current Password</label>
            <input id="account-current-password" name="currentPassword" type="password" placeholder="Required to change password" />
          </div>
          <div style={{ marginBottom: '1rem' }}>
            <label htmlFor="account-new-password" style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.9rem' }}>New Password</label>
            <input id="account-new-password" name="newPassword" type="password" placeholder="Leave blank to keep current" />
          </div>
          {error && <p className="error-text" style={{ marginBottom: '0.5rem' }}>{error}</p>}
          {ok && <p style={{ color: 'var(--success)', fontSize: '0.85rem', marginBottom: '0.5rem' }}>{ok}</p>}
          <button type="submit" className="btn btn-primary" style={{ width: '100%' }}>Update Profile</button>
        </form>
      </div>

      <div className="card">
        <h3 style={{ marginBottom: '0.75rem' }}>Security snapshot</h3>
        <p className="muted" style={{ fontSize: '0.9rem' }}>
          Role: <strong>{session.role}</strong>
        </p>
        <form action={logoutAction}>
          <button className="btn btn-secondary" style={{ marginTop: '0.6rem' }} type="submit">Logout</button>
        </form>
      </div>
    </div>
  )
}
