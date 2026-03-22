import { redirect } from 'next/navigation'
import { endpoints, serverDelete, serverGet, serverPost } from '../../lib/server/api'
import { clearSession, getSession } from '../../lib/server/session'

function isAuthError(error) {
  if (error?.status === 401) return true
  const text = String(error?.message || '').toLowerCase()
  return text.includes('invalid token') || text.includes('expired token') || text.includes('authorization header') || text.includes('unauthorized')
}

export default async function FavoritesPage({ searchParams }) {
  const qp = await searchParams
  const session = await getSession()
  if (!session) redirect('/login?next=/favorites')

  async function addFavoriteAction(formData) {
    'use server'
    const auth = await getSession()
    if (!auth) redirect('/login?next=/favorites')
    const recipeID = String(formData.get('recipeID') || '').trim()
    if (!recipeID) return
    let destination
    try {
      await serverPost(endpoints.favorites, `/favorites/${auth.id}/${recipeID}`, {}, auth.token)
      destination = '/favorites?ok=added'
    } catch (e) {
      if (isAuthError(e)) {
        await clearSession()
        redirect('/login?next=/favorites')
      }
      destination = `/favorites?error=${encodeURIComponent(e.message || 'Failed to add favorite')}`
    }
    redirect(destination)
  }

  async function removeFavoriteAction(formData) {
    'use server'
    const auth = await getSession()
    if (!auth) redirect('/login?next=/favorites')
    const recipeID = String(formData.get('recipeID') || '')
    let destination
    try {
      await serverDelete(endpoints.favorites, `/favorites/${auth.id}/${recipeID}`, auth.token)
      destination = '/favorites?ok=removed'
    } catch (e) {
      if (isAuthError(e)) {
        await clearSession()
        redirect('/login?next=/favorites')
      }
      destination = `/favorites?error=${encodeURIComponent(e.message || 'Failed to remove favorite')}`
    }
    redirect(destination)
  }

  let favorites = []
  let loadError = ''
  try {
    const data = await serverGet(endpoints.favorites, `/favorites/${session.id}`, session.token)
    favorites = Array.isArray(data) ? data : []
  } catch (e) {
    if (isAuthError(e)) {
      redirect('/login?next=/favorites')
    }
    loadError = e.message || 'Failed to load favorites'
  }

  const error = qp?.error ? decodeURIComponent(qp.error) : loadError

  return (
    <div style={{ maxWidth: 700, margin: '0 auto' }}>
      <h1 style={{ marginBottom: '1rem' }}>Favorites</h1>

      <form action={addFavoriteAction} style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem' }}>
        <label htmlFor="favorite-recipe-id" className="sr-only">Recipe ID to add to favorites</label>
        <input id="favorite-recipe-id" name="recipeID" type="text" placeholder="Recipe ID to favorite" style={{ flex: 1 }} />
        <button type="submit" className="btn btn-primary">Add</button>
      </form>

      {error && <p className="error-text" style={{ marginBottom: '1rem' }}>{error}</p>}

      {favorites.length === 0 ? (
        <p className="muted">No favorites yet.</p>
      ) : (
        favorites.map((fav) => (
          <div key={fav.id} className="card" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <strong>{fav.recipeName || fav.recipeId}</strong>
              <span className="muted" style={{ fontSize: '0.8rem', marginLeft: '0.75rem' }}>
                {new Date(fav.createdAt).toLocaleDateString()}
              </span>
            </div>
            <form action={removeFavoriteAction}>
              <input type="hidden" name="recipeID" value={fav.recipeId} />
              <button className="btn btn-danger" type="submit">Remove</button>
            </form>
          </div>
        ))
      )}
    </div>
  )
}
