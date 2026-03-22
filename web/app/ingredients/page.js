import { redirect } from 'next/navigation'
import Link from 'next/link'
import { endpoints, serverGet, serverPost } from '../../lib/server/api'
import { getSession } from '../../lib/server/session'

export default async function IngredientsPage({ searchParams }) {
  const qp = await searchParams
  const session = await getSession()
  const query = String(qp?.q || '').trim()
  const status = ['all', 'enriched', 'pending', 'review_required'].includes(String(qp?.status || ''))
    ? String(qp?.status)
    : 'all'
  const page = Math.max(1, Number(qp?.page || 1) || 1)

  async function suggestAction(formData) {
    'use server'
    const auth = await getSession()
    if (!auth) redirect('/login?next=/ingredients')
    const name = String(formData.get('name') || '').trim()
    if (!name) redirect('/ingredients?error=Ingredient%20name%20is%20required')

    let destination
    try {
      const data = await serverPost(endpoints.recipes, '/ingredients/suggestions', { name }, auth.token)
      if (data.status === 'matched') {
        destination = `/ingredients?ok=${encodeURIComponent(`Already tracked as ${data.match.canonicalName}.`)}`
      } else {
        destination = '/ingredients?ok=Suggestion%20queued%20for%20review'
      }
    } catch (e) {
      destination = `/ingredients?error=${encodeURIComponent(e.message || 'Failed to submit suggestion')}`
    }
    redirect(destination)
  }

  async function voteAction(formData) {
    'use server'
    const auth = await getSession()
    if (!auth) redirect('/login?next=/ingredients')
    const candidateID = String(formData.get('candidateID') || '')
    const vote = Number(formData.get('vote') || 0)
    let destination
    try {
      await serverPost(endpoints.recipes, `/ingredients/candidates/${candidateID}/votes`, { vote }, auth.token)
      destination = '/ingredients?ok=Vote%20recorded'
    } catch (e) {
      destination = `/ingredients?error=${encodeURIComponent(e.message || 'Vote failed')}`
    }
    redirect(destination)
  }

  let candidates = []
  if (session) {
    try {
      const data = await serverGet(endpoints.recipes, '/ingredients/candidates?status=pending', session.token)
      candidates = Array.isArray(data) ? data : []
    } catch {}
  }

  let catalog = { items: [], total: 0, page, pageSize: 50 }
  let catalogError = ''
  try {
    const data = await serverGet(
      endpoints.recipes,
      `/ingredients/catalog?q=${encodeURIComponent(query)}&status=${encodeURIComponent(status)}&page=${page}&pageSize=50`,
      session?.token,
    )
    catalog = data || catalog
  } catch (e) {
    catalogError = e.message || 'Failed to load ingredient catalog.'
  }

  const error = qp?.error ? decodeURIComponent(qp.error) : ''
  const ok = qp?.ok ? decodeURIComponent(qp.ok) : ''
  const totalPages = Math.max(1, Math.ceil((catalog.total || 0) / (catalog.pageSize || 50)))
  const prevPage = Math.max(1, page - 1)
  const nextPage = Math.min(totalPages, page + 1)

  return (
    <div className="page-wrap">
      <h1 style={{ marginBottom: '0.75rem' }}>Ingredient Catalog</h1>
      <p className="muted" style={{ marginBottom: '1rem' }}>
        Browse all known ingredients and their analysis state. Sign in to suggest new ingredients and vote.
      </p>

      <form method="GET" className="card" style={{ marginBottom: '1rem' }}>
        <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
          <input name="q" type="text" defaultValue={query} placeholder="Search ingredient name" style={{ flex: 1, minWidth: 220 }} />
          <select name="status" defaultValue={status} style={{ minWidth: 160 }}>
            <option value="all">All statuses</option>
            <option value="enriched">Enriched</option>
            <option value="pending">Pending</option>
            <option value="review_required">Review required</option>
          </select>
          <button className="btn btn-primary" type="submit">Filter</button>
        </div>
      </form>

      <div className="card" style={{ marginBottom: '1rem' }}>
        <h3 style={{ marginBottom: '0.75rem' }}>Catalog results</h3>
        <p className="muted" style={{ marginBottom: '0.75rem' }}>
          {catalog.total || 0} ingredient{catalog.total === 1 ? '' : 's'}
        </p>

        {catalogError ? <p className="error-text" role="alert" style={{ marginBottom: '0.65rem' }}>{catalogError}</p> : null}

        {catalog.items?.length ? (
          <div style={{ marginBottom: '0.75rem' }}>
            <h4 style={{ marginBottom: '0.45rem' }}>Browse highlights</h4>
            <div className="h-carousel">
              {catalog.items.slice(0, 12).map((item) => (
                <Link key={`featured-${item.id}`} className="h-carousel-card" href={`/ingredients/${item.id}`}>
                  <strong>{item.canonicalName}</strong>
                  <div className="muted" style={{ fontSize: '0.8rem', marginTop: '0.25rem' }}>
                    {(item.qualityScore ?? 0).toFixed(3)} quality
                  </div>
                  <div style={{ display: 'flex', gap: '0.35rem', marginTop: '0.3rem', flexWrap: 'wrap' }}>
                    <span className="tag">{item.analysisStatus}</span>
                    <span className="tag">cov {item.sourceCoverage}</span>
                  </div>
                </Link>
              ))}
            </div>
          </div>
        ) : null}

        {catalog.items?.length ? (
          <div style={{ display: 'grid', gap: '0.55rem' }}>
            {catalog.items.map((item) => (
              <Link key={item.id} href={`/ingredients/${item.id}`} style={{ textDecoration: 'none' }}>
              <div style={{ borderTop: '1px solid var(--border)', paddingTop: '0.6rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', gap: '0.5rem', flexWrap: 'wrap' }}>
                  <strong>{item.canonicalName}</strong>
                  <span className="muted">quality {(item.qualityScore ?? 0).toFixed(3)}</span>
                </div>
                <div style={{ display: 'flex', gap: '0.4rem', flexWrap: 'wrap', marginTop: '0.35rem' }}>
                  <span className="tag">{item.analysisStatus}</span>
                  {item.category ? <span className="tag">{item.category}</span> : null}
                  <span className="tag">coverage {item.sourceCoverage}</span>
                  <span className="tag">aliases {item.aliasCount}</span>
                  {item.flavourMoleculeCount ? <span className="tag">molecules {item.flavourMoleculeCount}</span> : null}
                </div>
              </div>
              </Link>
            ))}
          </div>
        ) : (
          <p className="muted">No ingredients match this filter.</p>
        )}

        <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '0.9rem' }}>
          <a className="btn btn-secondary" href={`/ingredients?q=${encodeURIComponent(query)}&status=${encodeURIComponent(status)}&page=${prevPage}`} aria-disabled={page <= 1}>Prev</a>
          <span className="muted">Page {page} of {totalPages}</span>
          <a className="btn btn-secondary" href={`/ingredients?q=${encodeURIComponent(query)}&status=${encodeURIComponent(status)}&page=${nextPage}`} aria-disabled={page >= totalPages}>Next</a>
        </div>
      </div>

      {session ? (
        <>
          <h2 style={{ marginBottom: '0.75rem' }}>Ingredient Suggestions</h2>
          <p className="muted" style={{ marginBottom: '1rem' }}>Help improve ingredient quality by suggesting additions and voting on pending items.</p>

          <form action={suggestAction} className="card" style={{ marginBottom: '1rem' }}>
            <label htmlFor="ingredient-suggestion-input" style={{ display: 'block', marginBottom: '0.4rem' }}>Ingredient name</label>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <input id="ingredient-suggestion-input" name="name" type="text" placeholder="e.g., scallion" required />
              <button className="btn btn-primary" type="submit">Suggest</button>
            </div>
            {ok && <p style={{ color: 'var(--success)', marginTop: '0.5rem' }} aria-live="polite">{ok}</p>}
            {error && <p className="error-text" style={{ marginTop: '0.5rem' }} role="alert">{error}</p>}
          </form>

          <div className="card">
            <h3 style={{ marginBottom: '0.75rem' }}>Pending candidates</h3>
            {candidates.length === 0 ? (
              <p className="muted">No pending ingredient candidates.</p>
            ) : (
              candidates.map((c) => (
                <div key={c.id} style={{ borderTop: '1px solid var(--border)', padding: '0.7rem 0' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', gap: '0.75rem', alignItems: 'center' }}>
                    <div>
                      <strong>{c.rawName}</strong>
                      <span className="muted" style={{ marginLeft: '0.5rem', fontSize: '0.85rem' }}>
                        normalized: {c.normalizedName}
                      </span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                      <form action={voteAction}>
                        <input type="hidden" name="candidateID" value={c.id} />
                        <input type="hidden" name="vote" value="1" />
                        <button className="btn" type="submit" aria-label={`Upvote ${c.rawName}`}>+1</button>
                      </form>
                      <form action={voteAction}>
                        <input type="hidden" name="candidateID" value={c.id} />
                        <input type="hidden" name="vote" value="-1" />
                        <button className="btn" type="submit" aria-label={`Downvote ${c.rawName}`}>-1</button>
                      </form>
                      <span className="muted">score: {c.voteScore || 0}</span>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </>
      ) : (
        <div className="card">
          <p className="muted">Sign in to suggest new ingredients and vote on pending entries.</p>
        </div>
      )}
    </div>
  )
}
