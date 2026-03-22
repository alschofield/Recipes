import Link from 'next/link'
import { redirect } from 'next/navigation'
import { endpoints, serverGet, serverPost } from '../../../lib/server/api'
import { clearSession, getSession } from '../../../lib/server/session'

function isAuthError(error) {
  if (error?.status === 401) return true
  const text = String(error?.message || '').toLowerCase()
  return text.includes('invalid token') || text.includes('expired token') || text.includes('authorization header') || text.includes('unauthorized')
}

export default async function AdminIngredientsPage({ searchParams }) {
  const qp = await searchParams
  const session = await getSession()
  if (!session) redirect('/login?next=/admin/ingredients')
  if (session.role !== 'admin') redirect('/recipes')

  const query = String(qp?.q || '').trim()
  const status = ['all', 'enriched', 'pending', 'review_required'].includes(String(qp?.status || ''))
    ? String(qp?.status)
    : 'all'
  const sort = ['quality_desc', 'quality_asc', 'coverage_desc', 'coverage_asc', 'name_asc', 'name_desc'].includes(String(qp?.sort || ''))
    ? String(qp?.sort)
    : 'quality_desc'
  const page = Math.max(1, Number(qp?.page || 1) || 1)
  const minQuality = Math.max(0, Number(qp?.minQuality || 0) || 0)
  const minCoverage = Math.max(0, Number(qp?.minCoverage || 0) || 0)
  const needsReview = String(qp?.needsReview || '') === 'true'

  async function resolveAction(formData) {
    'use server'
    const auth = await getSession()
    if (!auth || auth.role !== 'admin') redirect('/recipes')
    const candidateID = String(formData.get('candidateID') || '')
    const action = String(formData.get('action') || '')

    let destination
    try {
      await serverPost(endpoints.recipes, `/ingredients/candidates/${candidateID}/resolve`, { action }, auth.token)
      destination = '/admin/ingredients?ok=Candidate%20updated'
    } catch (e) {
      if (isAuthError(e)) {
        await clearSession()
        redirect('/login?next=/admin/ingredients')
      }
      if (e?.status === 403) {
        redirect('/recipes')
      }
      destination = `/admin/ingredients?error=${encodeURIComponent(e.message || 'Failed to resolve candidate')}`
    }
    redirect(destination)
  }

  let candidateData = []
  try {
    candidateData = await serverGet(endpoints.recipes, '/ingredients/candidates?status=pending', session.token)
  } catch (e) {
    if (isAuthError(e)) {
      redirect('/login?next=/admin/ingredients')
    }
    if (e?.status === 403) {
      redirect('/recipes')
    }
    candidateData = []
  }

  let metricData = null
  try {
    metricData = await serverGet(endpoints.recipes, '/ingredients/metrics', session.token)
  } catch (e) {
    if (isAuthError(e)) {
      redirect('/login?next=/admin/ingredients')
    }
    if (e?.status === 403) {
      redirect('/recipes')
    }
    metricData = null
  }

  let catalog = { items: [], total: 0, page, pageSize: 75 }
  let catalogError = ''
  try {
    const data = await serverGet(
      endpoints.recipes,
      `/ingredients/catalog?q=${encodeURIComponent(query)}&status=${encodeURIComponent(status)}&sort=${encodeURIComponent(sort)}&page=${page}&pageSize=75&minQuality=${encodeURIComponent(minQuality)}&minCoverage=${encodeURIComponent(minCoverage)}&needsReview=${needsReview ? 'true' : 'false'}`,
      session.token,
    )
    catalog = data || catalog
  } catch (e) {
    if (isAuthError(e)) {
      redirect('/login?next=/admin/ingredients')
    }
    if (e?.status === 403) {
      redirect('/recipes')
    }
    catalogError = e.message || 'Failed to load ingredient catalog'
  }

  const candidates = Array.isArray(candidateData) ? candidateData : []
  const metrics = metricData
  const items = Array.isArray(catalog?.items) ? catalog.items : []
  const totalPages = Math.max(1, Math.ceil((catalog?.total || 0) / (catalog?.pageSize || 75)))
  const prevPage = Math.max(1, page - 1)
  const nextPage = Math.min(totalPages, page + 1)
  const error = qp?.error ? decodeURIComponent(qp.error) : ''
  const ok = qp?.ok ? decodeURIComponent(qp.ok) : ''

  return (
    <div className="page-wrap page-wide">
      <h1 style={{ marginBottom: '1rem' }}>Admin Ingredients</h1>
      {error && <p className="error-text" role="alert">{error}</p>}
      {ok && <p style={{ color: 'var(--success)' }}>{ok}</p>}

      {metrics && (
        <div className="grid-2" style={{ marginBottom: '1rem' }}>
          <div className="card"><strong>{metrics.canonicalIngredients}</strong><p className="muted">Canonical</p></div>
          <div className="card"><strong>{metrics.aliases}</strong><p className="muted">Aliases</p></div>
          <div className="card"><strong>{metrics.pendingCandidates}</strong><p className="muted">Pending</p></div>
          <div className="card"><strong>{metrics.avgPendingAgeHours?.toFixed ? metrics.avgPendingAgeHours.toFixed(1) : metrics.avgPendingAgeHours}</strong><p className="muted">Avg pending age (hrs)</p></div>
        </div>
      )}

      <div className="card" style={{ marginBottom: '1rem' }}>
        <h3 style={{ marginBottom: '0.65rem' }}>All ingredients in database</h3>
        <form method="GET" style={{ display: 'grid', gap: '0.55rem', marginBottom: '0.8rem' }}>
          <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
            <input name="q" type="text" placeholder="Search by ingredient or category" defaultValue={query} style={{ flex: 1, minWidth: 220 }} />
            <select name="status" defaultValue={status} style={{ minWidth: 165 }}>
              <option value="all">All statuses</option>
              <option value="enriched">Enriched</option>
              <option value="pending">Pending</option>
              <option value="review_required">Review required</option>
            </select>
            <select name="sort" defaultValue={sort} style={{ minWidth: 165 }}>
              <option value="quality_desc">Quality high-low</option>
              <option value="quality_asc">Quality low-high</option>
              <option value="coverage_desc">Coverage high-low</option>
              <option value="coverage_asc">Coverage low-high</option>
              <option value="name_asc">Name A-Z</option>
              <option value="name_desc">Name Z-A</option>
            </select>
          </div>
          <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
            <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
              <span className="muted">Min quality</span>
              <input name="minQuality" type="number" min="0" max="1" step="0.05" defaultValue={minQuality} style={{ width: 90 }} />
            </label>
            <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
              <span className="muted">Min coverage</span>
              <input name="minCoverage" type="number" min="0" max="5" step="1" defaultValue={minCoverage} style={{ width: 80 }} />
            </label>
            <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
              <input name="needsReview" type="checkbox" value="true" defaultChecked={needsReview} />
              <span className="muted">Needs review only</span>
            </label>
            <button className="btn btn-primary" type="submit">Apply filters</button>
          </div>
        </form>

        <p className="muted" style={{ marginBottom: '0.6rem' }}>
          {catalog?.total || 0} ingredient{catalog?.total === 1 ? '' : 's'} visible
        </p>

        {catalogError ? <p className="error-text" role="alert" style={{ marginBottom: '0.6rem' }}>{catalogError}</p> : null}

        {items.length === 0 ? (
          <p className="muted">No ingredients matched the current filters.</p>
        ) : (
          <div className="admin-catalog-grid">
            {items.map((item) => (
              <Link key={item.id} href={`/ingredients/${item.id}`} className="admin-catalog-link">
                <div className="admin-catalog-card">
                  <div style={{ display: 'flex', justifyContent: 'space-between', gap: '0.6rem', flexWrap: 'wrap' }}>
                    <strong>{item.canonicalName}</strong>
                    <span className="mono muted">q {(item.qualityScore || 0).toFixed(3)}</span>
                  </div>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.4rem', marginTop: '0.3rem' }}>
                    <span className="tag">{item.analysisStatus}</span>
                    <span className="tag">coverage {item.sourceCoverage}</span>
                    <span className="tag">aliases {item.aliasCount}</span>
                    {item.category ? <span className="tag">{item.category}</span> : null}
                    {item.flavourMoleculeCount ? <span className="tag">molecules {item.flavourMoleculeCount}</span> : null}
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}

        <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '0.8rem' }}>
          <a
            className="btn btn-secondary"
            href={`/admin/ingredients?q=${encodeURIComponent(query)}&status=${encodeURIComponent(status)}&sort=${encodeURIComponent(sort)}&minQuality=${encodeURIComponent(minQuality)}&minCoverage=${encodeURIComponent(minCoverage)}&needsReview=${needsReview ? 'true' : 'false'}&page=${prevPage}`}
            aria-disabled={page <= 1}
          >
            Prev
          </a>
          <span className="muted">Page {page} of {totalPages}</span>
          <a
            className="btn btn-secondary"
            href={`/admin/ingredients?q=${encodeURIComponent(query)}&status=${encodeURIComponent(status)}&sort=${encodeURIComponent(sort)}&minQuality=${encodeURIComponent(minQuality)}&minCoverage=${encodeURIComponent(minCoverage)}&needsReview=${needsReview ? 'true' : 'false'}&page=${nextPage}`}
            aria-disabled={page >= totalPages}
          >
            Next
          </a>
        </div>
      </div>

      <div className="card">
        <h3 style={{ marginBottom: '0.75rem' }}>Pending candidates</h3>
        {candidates.length === 0 ? (
          <p className="muted">No pending candidates.</p>
        ) : candidates.map((c) => (
          <div key={c.id} style={{ borderTop: '1px solid var(--border)', padding: '0.75rem 0' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', gap: '1rem', alignItems: 'center' }}>
              <div>
                <strong>{c.rawName}</strong>
                <div className="muted" style={{ fontSize: '0.85rem' }}>
                  normalized: {c.normalizedName} | source: {c.source} | votes: {c.voteScore || 0}
                </div>
              </div>
              <div style={{ display: 'flex', gap: '0.4rem' }}>
                <form action={resolveAction}>
                  <input type="hidden" name="candidateID" value={c.id} />
                  <input type="hidden" name="action" value="approve_canonical" />
                  <button className="btn btn-primary" type="submit">Approve Canonical</button>
                </form>
                <form action={resolveAction}>
                  <input type="hidden" name="candidateID" value={c.id} />
                  <input type="hidden" name="action" value="approve_alias" />
                  <button className="btn" type="submit">Approve Alias</button>
                </form>
                <form action={resolveAction}>
                  <input type="hidden" name="candidateID" value={c.id} />
                  <input type="hidden" name="action" value="reject" />
                  <button className="btn btn-danger" type="submit">Reject</button>
                </form>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
