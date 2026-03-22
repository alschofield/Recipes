import Link from 'next/link'
import { redirect } from 'next/navigation'
import { endpoints, serverGet } from '../../../lib/server/api'
import { clearSession, getSession } from '../../../lib/server/session'

function isAuthError(error) {
  if (error?.status === 401) return true
  const text = String(error?.message || '').toLowerCase()
  return text.includes('invalid token') || text.includes('expired token') || text.includes('authorization header') || text.includes('unauthorized')
}

export default async function AdminRecipesPage({ searchParams }) {
  const qp = await searchParams
  const session = await getSession()
  if (!session) redirect('/login?next=/admin/recipes')
  if (session.role !== 'admin') redirect('/recipes')

  const query = String(qp?.q || '').trim()
  const status = ['all', 'computed', 'pending', 'failed', 'missing'].includes(String(qp?.status || ''))
    ? String(qp?.status)
    : 'all'
  const source = ['all', 'database', 'llm'].includes(String(qp?.source || ''))
    ? String(qp?.source)
    : 'all'
  const sort = ['updated_desc', 'updated_asc', 'quality_desc', 'quality_asc', 'name_asc', 'name_desc'].includes(String(qp?.sort || ''))
    ? String(qp?.sort)
    : 'updated_desc'
  const page = Math.max(1, Number(qp?.page || 1) || 1)
  const minQuality = Math.max(0, Number(qp?.minQuality || 0) || 0)
  const maxQuality = Math.min(1, Number(qp?.maxQuality || 1) || 1)
  const needsReview = String(qp?.needsReview || '') === 'true'

  let data = null
  let dataError = ''
  try {
    data = await serverGet(
      endpoints.recipes,
      `/recipes/analysis/admin?q=${encodeURIComponent(query)}&status=${encodeURIComponent(status)}&source=${encodeURIComponent(source)}&sort=${encodeURIComponent(sort)}&page=${page}&pageSize=75&minQuality=${encodeURIComponent(minQuality)}&maxQuality=${encodeURIComponent(maxQuality)}&needsReview=${needsReview ? 'true' : 'false'}`,
      session.token,
    )
  } catch (e) {
    if (isAuthError(e)) {
      redirect('/login?next=/admin/recipes')
    }
    if (e?.status === 403 || String(e?.message || '').toLowerCase().includes('insufficient permissions')) {
      redirect('/recipes')
    }
    dataError = e.message || 'Could not load recipe admin dashboard.'
  }

  if (!data) {
    return (
      <div className="page-wrap status-box">
        <p className="error-text">Could not load recipe admin dashboard.</p>
        {dataError ? <p className="muted" style={{ marginBottom: '0.75rem' }}>{dataError}</p> : null}
        <Link className="btn btn-secondary" href="/recipes">Back to recipes</Link>
      </div>
    )
  }

  const overview = data.overview || {}
  const topRecipes = Array.isArray(data.topRecipes) ? data.topRecipes : []
  const lowRecipes = Array.isArray(data.lowRecipes) ? data.lowRecipes : []
  const byCuisine = Array.isArray(data.byCuisine) ? data.byCuisine : []
  const bySource = Array.isArray(data.bySource) ? data.bySource : []
  const staleQueue = Array.isArray(data.staleQueue) ? data.staleQueue : []
  const recipes = Array.isArray(data.recipes) ? data.recipes : []
  const total = Number(data.totalRecipes || 0)
  const pageSize = Number(data.pageSize || 75)
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  const prevPage = Math.max(1, page - 1)
  const nextPage = Math.min(totalPages, page + 1)

  return (
    <div className="page-wrap page-wide">
      <h1 style={{ marginBottom: '0.8rem' }}>Admin Recipes</h1>
      <p className="muted" style={{ marginBottom: '1rem' }}>
        View every recipe row in the database, then filter by analysis/quality/source for operations.
      </p>

      <div className="card" style={{ marginBottom: '1rem' }}>
        <h3 style={{ marginBottom: '0.65rem' }}>All recipes in database</h3>
        <form method="GET" style={{ display: 'grid', gap: '0.55rem', marginBottom: '0.8rem' }}>
          <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
            <input name="q" type="text" defaultValue={query} placeholder="Search recipe name or cuisine" style={{ flex: 1, minWidth: 220 }} />
            <select name="status" defaultValue={status} style={{ minWidth: 165 }}>
              <option value="all">All analysis statuses</option>
              <option value="computed">Computed</option>
              <option value="pending">Pending</option>
              <option value="failed">Failed</option>
              <option value="missing">Missing</option>
            </select>
            <select name="source" defaultValue={source} style={{ minWidth: 130 }}>
              <option value="all">All sources</option>
              <option value="database">database</option>
              <option value="llm">llm</option>
            </select>
            <select name="sort" defaultValue={sort} style={{ minWidth: 165 }}>
              <option value="updated_desc">Updated newest</option>
              <option value="updated_asc">Updated oldest</option>
              <option value="quality_desc">Quality high-low</option>
              <option value="quality_asc">Quality low-high</option>
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
              <span className="muted">Max quality</span>
              <input name="maxQuality" type="number" min="0" max="1" step="0.05" defaultValue={maxQuality} style={{ width: 90 }} />
            </label>
            <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
              <input name="needsReview" type="checkbox" value="true" defaultChecked={needsReview} />
              <span className="muted">Needs review only</span>
            </label>
            <button className="btn btn-primary" type="submit">Apply filters</button>
          </div>
        </form>

        <p className="muted" style={{ marginBottom: '0.6rem' }}>{total} recipe{total === 1 ? '' : 's'} visible</p>

        {recipes.length === 0 ? (
          <p className="muted">No recipes matched the current filters.</p>
        ) : (
          <div style={{ display: 'grid', gap: '0.45rem' }}>
            {recipes.map((item) => (
              <Link key={item.recipeId} href={`/recipes/${item.recipeId}`} style={{ textDecoration: 'none' }}>
                <div style={{ borderTop: '1px solid var(--border-subtle)', paddingTop: '0.55rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', gap: '0.5rem', flexWrap: 'wrap' }}>
                    <strong>{item.name}</strong>
                    <span className="mono muted">overall {(item.overallScore || 0).toFixed(3)}</span>
                  </div>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.4rem', marginTop: '0.3rem' }}>
                    <span className="tag">{item.analysisStatus}</span>
                    <span className="tag">{item.source}</span>
                    {item.cuisine ? <span className="tag">{item.cuisine}</span> : null}
                    <span className="tag">difficulty {item.difficulty}</span>
                    <span className="tag">ingredients {item.ingredientCount}</span>
                    {item.needsReview ? <span className="tag">needs review</span> : null}
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}

        <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '0.8rem' }}>
          <a
            className="btn btn-secondary"
            href={`/admin/recipes?q=${encodeURIComponent(query)}&status=${encodeURIComponent(status)}&source=${encodeURIComponent(source)}&sort=${encodeURIComponent(sort)}&minQuality=${encodeURIComponent(minQuality)}&maxQuality=${encodeURIComponent(maxQuality)}&needsReview=${needsReview ? 'true' : 'false'}&page=${prevPage}`}
            aria-disabled={page <= 1}
          >
            Prev
          </a>
          <span className="muted">Page {page} of {totalPages}</span>
          <a
            className="btn btn-secondary"
            href={`/admin/recipes?q=${encodeURIComponent(query)}&status=${encodeURIComponent(status)}&source=${encodeURIComponent(source)}&sort=${encodeURIComponent(sort)}&minQuality=${encodeURIComponent(minQuality)}&maxQuality=${encodeURIComponent(maxQuality)}&needsReview=${needsReview ? 'true' : 'false'}&page=${nextPage}`}
            aria-disabled={page >= totalPages}
          >
            Next
          </a>
        </div>
      </div>

      <div className="grid-2" style={{ marginBottom: '1rem' }}>
        <div className="card"><strong>{overview.totalRecipes || 0}</strong><p className="muted">Total recipes</p></div>
        <div className="card"><strong>{overview.computedAnalyses || 0}</strong><p className="muted">Computed analyses</p></div>
        <div className="card"><strong>{overview.pendingOrMissing || 0}</strong><p className="muted">Pending/missing analyses</p></div>
        <div className="card"><strong>{overview.failedAnalyses || 0}</strong><p className="muted">Failed analyses</p></div>
        <div className="card"><strong>{(overview.averageOverallScore || 0).toFixed(3)}</strong><p className="muted">Avg overall score</p></div>
        <div className="card"><strong>{(overview.averageIngredientCoverage || 0).toFixed(3)}</strong><p className="muted">Avg ingredient coverage</p></div>
      </div>

      <div className="grid-2" style={{ marginBottom: '1rem' }}>
        <div className="card">
          <h3 style={{ marginBottom: '0.6rem' }}>Top scored recipes</h3>
          {topRecipes.length === 0 ? <p className="muted">No data</p> : (
            <div style={{ display: 'grid', gap: '0.5rem' }}>
              {topRecipes.map((item) => (
                <Link key={item.recipeId} href={`/recipes/${item.recipeId}`} style={{ textDecoration: 'none' }}>
                  <div style={{ borderTop: '1px solid var(--border)', paddingTop: '0.5rem' }}>
                    <strong>{item.name}</strong>
                    <div className="muted">score {(item.overallScore || 0).toFixed(3)} | {item.source}</div>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>

        <div className="card">
          <h3 style={{ marginBottom: '0.6rem' }}>Lowest scored recipes</h3>
          {lowRecipes.length === 0 ? <p className="muted">No data</p> : (
            <div style={{ display: 'grid', gap: '0.5rem' }}>
              {lowRecipes.map((item) => (
                <Link key={item.recipeId} href={`/recipes/${item.recipeId}`} style={{ textDecoration: 'none' }}>
                  <div style={{ borderTop: '1px solid var(--border)', paddingTop: '0.5rem' }}>
                    <strong>{item.name}</strong>
                    <div className="muted">score {(item.overallScore || 0).toFixed(3)} | {item.source}</div>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="grid-2" style={{ marginBottom: '1rem' }}>
        <div className="card">
          <h3 style={{ marginBottom: '0.6rem' }}>Averages by cuisine</h3>
          {byCuisine.length === 0 ? <p className="muted">No data</p> : byCuisine.map((row) => (
            <div key={row.label} style={{ borderTop: '1px solid var(--border)', padding: '0.5rem 0', display: 'flex', justifyContent: 'space-between' }}>
              <span>{row.label}</span>
              <span className="muted">{row.recipeCount} / {(row.averageOverallScore || 0).toFixed(3)}</span>
            </div>
          ))}
        </div>

        <div className="card">
          <h3 style={{ marginBottom: '0.6rem' }}>Averages by source</h3>
          {bySource.length === 0 ? <p className="muted">No data</p> : bySource.map((row) => (
            <div key={row.label} style={{ borderTop: '1px solid var(--border)', padding: '0.5rem 0', display: 'flex', justifyContent: 'space-between' }}>
              <span>{row.label}</span>
              <span className="muted">{row.recipeCount} / {(row.averageOverallScore || 0).toFixed(3)}</span>
            </div>
          ))}
        </div>
      </div>

      <div className="card">
        <h3 style={{ marginBottom: '0.6rem' }}>Stale analysis queue</h3>
        {staleQueue.length === 0 ? <p className="muted">No stale entries.</p> : (
          <div style={{ display: 'grid', gap: '0.5rem' }}>
            {staleQueue.map((item) => (
              <Link key={`${item.recipeId}-${item.recipeUpdatedAt}`} href={`/recipes/${item.recipeId}`} style={{ textDecoration: 'none' }}>
                <div style={{ borderTop: '1px solid var(--border)', paddingTop: '0.5rem' }}>
                  <strong>{item.name}</strong>
                  <div className="muted">{item.analysisStatus} | updated {item.recipeUpdatedAt}</div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
