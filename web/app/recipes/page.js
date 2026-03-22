import Link from 'next/link'
import { redirect } from 'next/navigation'
import { endpoints, serverGet, serverPost } from '../../lib/server/api'
import { getSession } from '../../lib/server/session'

export default async function RecipesPage({ searchParams }) {
  const qp = await searchParams
  const session = await getSession()
  const ingredientsRaw = String(qp?.ingredients || '')
  const mode = qp?.mode === 'inclusive' ? 'inclusive' : 'strict'
  const query = String(qp?.q || '').trim()
  const source = ['all', 'database', 'llm'].includes(String(qp?.source || ''))
    ? String(qp?.source)
    : 'all'
  const sort = ['updated_desc', 'updated_asc', 'quality_desc', 'quality_asc', 'name_asc', 'name_desc'].includes(String(qp?.sort || ''))
    ? String(qp?.sort)
    : 'updated_desc'
  const page = Math.max(1, Number(qp?.page || 1) || 1)

  const ingredients = ingredientsRaw
    .split(',')
    .map((s) => s.trim().toLowerCase())
    .filter(Boolean)

  let results = null
  let catalog = { items: [], total: 0, page, pageSize: 20 }
  let error = ''

  if (ingredients.length > 0) {
    try {
      results = await serverPost(endpoints.recipes, '/recipes/search', {
        ingredients,
        mode,
        pagination: { page: 1, pageSize: 20 },
      }, session?.token)
    } catch (e) {
      error = e.message || 'Search failed'
    }
  } else {
    try {
      catalog = await serverGet(
        endpoints.recipes,
        `/recipes/catalog?q=${encodeURIComponent(query)}&source=${encodeURIComponent(source)}&sort=${encodeURIComponent(sort)}&page=${page}&pageSize=20`,
        session?.token,
      )
    } catch (e) {
      error = e.message || 'Failed to load recipe catalog'
    }
  }

  async function favoriteAction(formData) {
    'use server'
    const auth = await getSession()
    if (!auth) redirect('/login?next=/recipes')

    const recipeID = String(formData.get('recipeID') || '')
    try {
      await serverPost(endpoints.favorites, `/favorites/${auth.id}/${recipeID}`, {}, auth.token)
    } catch {}

    const q = String(formData.get('ingredients') || '')
    const m = String(formData.get('mode') || 'strict')
    redirect(`/recipes?ingredients=${encodeURIComponent(q)}&mode=${encodeURIComponent(m)}`)
  }

  return (
    <div className="page-wrap">
      <h1 style={{ marginBottom: '0.7rem' }}>Recipe Search</h1>
      <p className="muted" style={{ marginBottom: '1rem' }}>Enter ingredients as a comma-separated list.</p>

      <form method="GET" className="card pantry-wrap" style={{ marginBottom: '1rem' }}>
        <label htmlFor="ingredients" style={{ display: 'block', marginBottom: '0.4rem' }}>Ingredients</label>
        <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
          <input
            id="ingredients"
            name="ingredients"
            type="text"
            defaultValue={ingredientsRaw}
            placeholder="chicken, rice, garlic"
            style={{ flex: 1, minWidth: 220 }}
          />
          <select name="mode" defaultValue={mode} style={{ width: 150 }}>
            <option value="strict">Strict</option>
            <option value="inclusive">Inclusive</option>
          </select>
          <button type="submit" className="btn btn-primary">Search</button>
        </div>
      </form>

      <form method="GET" className="card pantry-wrap" style={{ marginBottom: '1rem' }}>
        <label htmlFor="q" style={{ display: 'block', marginBottom: '0.4rem' }}>Browse all recipes</label>
        <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
          <input id="q" name="q" type="text" defaultValue={query} placeholder="Search by recipe name or cuisine" style={{ flex: 1, minWidth: 220 }} />
          <select name="source" defaultValue={source} style={{ minWidth: 140 }}>
            <option value="all">All sources</option>
            <option value="database">database</option>
            <option value="llm">llm</option>
          </select>
          <select name="sort" defaultValue={sort} style={{ minWidth: 170 }}>
            <option value="updated_desc">Updated newest</option>
            <option value="updated_asc">Updated oldest</option>
            <option value="quality_desc">Quality high-low</option>
            <option value="quality_asc">Quality low-high</option>
            <option value="name_asc">Name A-Z</option>
            <option value="name_desc">Name Z-A</option>
          </select>
          <button type="submit" className="btn btn-secondary">Browse</button>
        </div>
      </form>

      {error && (
        <div className="status-box" role="alert" style={{ marginBottom: '0.9rem' }}>
          <p className="error-text">{error}</p>
        </div>
      )}

      {results && (
        <>
          <p className="muted" style={{ marginBottom: '1rem' }}>
            {results.pagination.total} result{results.pagination.total !== 1 ? 's' : ''} (mode: {results.mode})
          </p>
          {results.results.length === 0 ? (
            <div className="status-box">
              <p className="muted">No matching recipes found.</p>
            </div>
          ) : (
            <div className="grid-2">
              {results.results.map((recipe) => (
                <div className="card recipe-card" key={recipe.id}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '0.5rem' }}>
                    <div>
                      <strong style={{ fontSize: '1.05rem' }}>{recipe.name}</strong>
                      <span className="muted" style={{ fontSize: '0.8rem', marginLeft: '0.5rem' }}>{recipe.source}</span>
                    </div>
                  </div>
                  <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', marginBottom: '0.75rem' }}>
                    <span className="tag">{Math.round((recipe.matchPercent ?? 0) * 100)}% match</span>
                    {recipe.difficulty && <span className="tag">{recipe.difficulty}</span>}
                    {recipe.cuisine && <span className="tag">{recipe.cuisine}</span>}
                    {recipe.totalMinutes > 0 && <span className="tag">{recipe.totalMinutes} min</span>}
                  </div>
                  <div className="recipe-card-footer no-print">
                    <Link href={`/recipes/${recipe.id}`} className="btn btn-secondary">Details</Link>
                    {session && (
                      <form action={favoriteAction}>
                        <input type="hidden" name="recipeID" value={recipe.id} />
                        <input type="hidden" name="ingredients" value={ingredientsRaw} />
                        <input type="hidden" name="mode" value={mode} />
                        <button className="btn btn-primary" type="submit">Add Favorite</button>
                      </form>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}

      {!results && !error && (
        <>
          <p className="muted" style={{ marginBottom: '1rem' }}>
            {catalog.total} recipe{catalog.total === 1 ? '' : 's'} available
          </p>
          {catalog.items?.length === 0 ? (
            <div className="status-box">
              <p className="muted">No recipes found for this filter.</p>
            </div>
          ) : (
            <div className="grid-2">
              {catalog.items.map((recipe) => (
                <div className="card recipe-card" key={recipe.id}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '0.5rem' }}>
                    <div>
                      <strong style={{ fontSize: '1.05rem' }}>{recipe.name}</strong>
                      <span className="muted" style={{ fontSize: '0.8rem', marginLeft: '0.5rem' }}>{recipe.source}</span>
                    </div>
                  </div>
                  {recipe.description ? <p className="muted" style={{ marginBottom: '0.55rem' }}>{recipe.description}</p> : null}
                  <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', marginBottom: '0.75rem' }}>
                    <span className="tag">q {(recipe.qualityScore ?? 0).toFixed(3)}</span>
                    {recipe.difficulty && <span className="tag">{recipe.difficulty}</span>}
                    {recipe.cuisine && <span className="tag">{recipe.cuisine}</span>}
                    {recipe.totalMinutes > 0 && <span className="tag">{recipe.totalMinutes} min</span>}
                    {recipe.servings > 0 && <span className="tag">{recipe.servings} servings</span>}
                  </div>
                  <div className="recipe-card-footer no-print">
                    <Link href={`/recipes/${recipe.id}`} className="btn btn-secondary">Details</Link>
                    {session && (
                      <form action={favoriteAction}>
                        <input type="hidden" name="recipeID" value={recipe.id} />
                        <input type="hidden" name="ingredients" value={ingredientsRaw} />
                        <input type="hidden" name="mode" value={mode} />
                        <button className="btn btn-primary" type="submit">Add Favorite</button>
                      </form>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}

          {catalog.total > 0 && (
            <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '1rem' }}>
              <a className="btn btn-secondary" href={`/recipes?q=${encodeURIComponent(query)}&source=${encodeURIComponent(source)}&sort=${encodeURIComponent(sort)}&page=${Math.max(1, page - 1)}`} aria-disabled={page <= 1}>Prev</a>
              <span className="muted">Page {page} of {Math.max(1, Math.ceil((catalog.total || 0) / (catalog.pageSize || 20)))}</span>
              <a className="btn btn-secondary" href={`/recipes?q=${encodeURIComponent(query)}&source=${encodeURIComponent(source)}&sort=${encodeURIComponent(sort)}&page=${Math.min(Math.max(1, Math.ceil((catalog.total || 0) / (catalog.pageSize || 20))), page + 1)}`} aria-disabled={page >= Math.max(1, Math.ceil((catalog.total || 0) / (catalog.pageSize || 20)))}>Next</a>
            </div>
          )}
        </>
      )}
    </div>
  )
}
