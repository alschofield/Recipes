import Link from 'next/link'
import { redirect } from 'next/navigation'
import { endpoints, serverGet, serverPost } from '../../lib/server/api'
import { getSession } from '../../lib/server/session'
import { trackEvent } from '../../lib/server/telemetry'
import PantryComposer from './PantryComposer'

export default async function RecipesPage({ searchParams }) {
  const qp = await searchParams
  const session = await getSession()
  const ingredientsRaw = String(qp?.ingredients || '')
  const mode = qp?.mode === 'inclusive' ? 'inclusive' : 'strict'
  const query = String(qp?.q || '').trim()
  const density = ['compact', 'cozy', 'spacious'].includes(String(qp?.density || ''))
    ? String(qp?.density)
    : 'cozy'
  const source = ['all', 'database', 'llm'].includes(String(qp?.source || ''))
    ? String(qp?.source)
    : 'all'
  const sort = ['updated_desc', 'updated_asc', 'quality_desc', 'quality_asc', 'name_asc', 'name_desc'].includes(String(qp?.sort || ''))
    ? String(qp?.sort)
    : 'updated_desc'
  const page = Math.max(1, Number(qp?.page || 1) || 1)
  const complexQuery = String(qp?.complex || '').toLowerCase() === 'true'

  const ingredients = ingredientsRaw
    .split(',')
    .map((s) => s.trim().toLowerCase())
    .filter(Boolean)
  const complex = complexQuery || ingredients.length >= 10

  let results = null
  let catalog = { items: [], total: 0, page, pageSize: 20 }
  let error = ''

  if (ingredients.length > 0) {
    try {
      results = await serverPost(endpoints.recipes, '/recipes/search', {
        ingredients,
        mode,
        complex,
        pagination: { page: 1, pageSize: 20 },
      }, session?.token)

      const llmCount = Array.isArray(results?.results)
        ? results.results.filter((item) => item.source === 'llm').length
        : 0
      const dbCount = Array.isArray(results?.results)
        ? results.results.filter((item) => item.source === 'database').length
        : 0

      trackEvent('search_intent', {
        mode,
        complex,
        ingredientCount: ingredients.length,
        totalResults: Number(results?.pagination?.total || 0),
        llmResults: llmCount,
        dbResults: dbCount,
      })
    } catch (e) {
      error = e.message || 'Search failed'
      trackEvent('search_error', {
        mode,
        complex,
        ingredientCount: ingredients.length,
        message: error,
      })
    }
  } else {
    try {
      catalog = await serverGet(
        endpoints.recipes,
        `/recipes/catalog?q=${encodeURIComponent(query)}&source=${encodeURIComponent(source)}&sort=${encodeURIComponent(sort)}&page=${page}&pageSize=20`,
        session?.token,
      )

      if (query) {
        trackEvent('catalog_query', {
          query,
          source,
          sort,
          page,
          totalResults: Number(catalog?.total || 0),
        })
      }
    } catch (e) {
      error = e.message || 'Failed to load recipe catalog'
      if (query) {
        trackEvent('catalog_query_error', {
          query,
          source,
          sort,
          page,
          message: error,
        })
      }
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
    const c = String(formData.get('complex') || '').toLowerCase() === 'true'
    const d = String(formData.get('density') || 'cozy')
    redirect(`/recipes?ingredients=${encodeURIComponent(q)}&mode=${encodeURIComponent(m)}&complex=${c ? 'true' : 'false'}&density=${encodeURIComponent(d)}`)
  }

  const hasMixedSources = Boolean(
    results?.results?.some((item) => item.source === 'database')
      && results?.results?.some((item) => item.source === 'llm'),
  )

  return (
    <div className="page-wrap search-stack">
      <div className="page-header">
        <h1>Discover Recipes</h1>
        <p className="muted">Tell Ingrediential what is in your kitchen and get ranked cooking options in seconds.</p>
      </div>

      <form method="GET" className="card pantry-wrap">
        <p className="field-label" style={{ fontSize: '0.86rem' }}>Ingredient-first search</p>
        <PantryComposer defaultRaw={ingredientsRaw} />
        <fieldset className="search-controls" aria-label="Search controls">
          <legend className="sr-only">Search controls</legend>
          <div style={{ minWidth: 150 }}>
            <label htmlFor="mode" className="field-label">Match mode</label>
            <select id="mode" name="mode" defaultValue={mode} style={{ width: '100%' }}>
              <option value="strict">Strict</option>
              <option value="inclusive">Inclusive</option>
            </select>
          </div>
          <label htmlFor="complexToggle" style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <input id="complexToggle" type="checkbox" name="complex" value="true" defaultChecked={complex} />
            Complex mode
          </label>
          <div style={{ minWidth: 140 }}>
            <label htmlFor="density" className="field-label">Result density</label>
            <select id="density" name="density" defaultValue={density} style={{ width: '100%' }}>
              <option value="compact">Compact</option>
              <option value="cozy">Cozy</option>
              <option value="spacious">Spacious</option>
            </select>
          </div>
          <button type="submit" className="btn btn-primary">Find Matches</button>
        </fieldset>
        <p className="muted" style={{ marginTop: '0.5rem', marginBottom: 0 }}>
          Complex mode auto-enables at 10+ ingredients.
        </p>
        <div className="hint-grid" style={{ marginTop: '0.55rem' }}>
          <p><strong>Strict:</strong> returns recipes where all required ingredients are present.</p>
          <p><strong>Inclusive:</strong> allows missing ingredients and shows what to add.</p>
          <p><strong>Complex:</strong> prompts more advanced generated recipes (auto on at 10+ ingredients).</p>
        </div>
      </form>

      <form method="GET" className="card pantry-wrap">
        <p className="field-label" style={{ fontSize: '0.86rem' }}>Library browser</p>
        <label htmlFor="q" style={{ display: 'block', marginBottom: '0.4rem' }}>Browse recipe library</label>
        <div className="search-controls">
          <input id="q" name="q" type="text" defaultValue={query} placeholder="Search by recipe name or cuisine" style={{ flex: 1, minWidth: 220 }} />
          <div style={{ minWidth: 140 }}>
            <label htmlFor="source" className="field-label">Source</label>
            <select id="source" name="source" defaultValue={source} style={{ minWidth: 140 }}>
              <option value="all">All sources</option>
              <option value="database">database</option>
              <option value="llm">llm</option>
            </select>
          </div>
          <div style={{ minWidth: 170 }}>
            <label htmlFor="sort" className="field-label">Sort</label>
            <select id="sort" name="sort" defaultValue={sort} style={{ minWidth: 170 }}>
              <option value="updated_desc">Updated newest</option>
              <option value="updated_asc">Updated oldest</option>
              <option value="quality_desc">Quality high-low</option>
              <option value="quality_asc">Quality low-high</option>
              <option value="name_asc">Name A-Z</option>
              <option value="name_desc">Name Z-A</option>
            </select>
          </div>
          <div style={{ minWidth: 140 }}>
            <label htmlFor="browseDensity" className="field-label">Density</label>
            <select id="browseDensity" name="density" defaultValue={density} style={{ width: 140 }}>
              <option value="compact">Compact</option>
              <option value="cozy">Cozy</option>
              <option value="spacious">Spacious</option>
            </select>
          </div>
          <button type="submit" className="btn btn-secondary">Apply Filters</button>
        </div>
      </form>

      {error && (
        <div className="status-box" role="alert" style={{ marginBottom: '0.9rem' }}>
          <p className="error-text" style={{ marginBottom: '0.45rem' }}>{error}</p>
          <div style={{ display: 'flex', gap: '0.55rem', flexWrap: 'wrap' }}>
            <Link
              className="btn btn-secondary"
              href={ingredientsRaw
                ? `/recipes?ingredients=${encodeURIComponent(ingredientsRaw)}&mode=${encodeURIComponent(mode)}&complex=${complex ? 'true' : 'false'}`
                : `/recipes?q=${encodeURIComponent(query)}&source=${encodeURIComponent(source)}&sort=${encodeURIComponent(sort)}&page=${page}`}
            >
              Retry
            </Link>
            {ingredientsRaw ? (
              <Link className="btn btn-secondary" href="/recipes">Browse catalog</Link>
            ) : null}
          </div>
        </div>
      )}

      {results && (
        <>
          <div className="result-meta" aria-live="polite">
            <span className="tag">{results.pagination.total} result{results.pagination.total !== 1 ? 's' : ''}</span>
            <span className="tag">mode: {results.mode}</span>
            <span className="tag">density: {density}</span>
          </div>
          {hasMixedSources && (
            <div className="status-box" style={{ marginBottom: '0.85rem' }}>
              <p className="muted" style={{ marginBottom: '0.4rem' }}>
                Mixed results: database recipes are blended with generated fallback recipes.
              </p>
              <div className="chip-list" aria-label="source legend" role="list">
                <span className="tag">database: canonical recipes</span>
                <span className="tag">llm: generated fallback recipes</span>
                <span className="tag">match %: ingredient overlap confidence hint</span>
              </div>
            </div>
          )}
          {results.results.length === 0 ? (
            <div className="status-box">
              <p className="muted" style={{ marginBottom: '0.4rem' }}>No direct matches yet.</p>
              {mode === 'strict' ? (
                <p className="muted">
                  Try switching to <strong>Inclusive</strong> mode or remove 1-2 restrictive ingredients.
                </p>
              ) : (
                  <p className="muted">Try broader ingredients or enable Complex mode for a wider generation pass.</p>
                )}
              </div>
          ) : (
            <div className="grid-2">
              {results.results.map((recipe) => (
                <div className={`card recipe-card recipe-density-${density}`} key={recipe.id}>
                  <div className="recipe-card-header">
                    <div>
                      <strong className="recipe-title">{recipe.name}</strong>
                      <span className="muted" style={{ fontSize: '0.8rem', marginLeft: '0.5rem' }}>{String(recipe.source || '').toUpperCase()}</span>
                      {recipe.source === 'llm' && <span className="tag tag-review">generated (reviewable)</span>}
                    </div>
                  </div>
                  <div className="chip-list" style={{ marginBottom: '0.75rem' }}>
                    <span className="tag">{Math.round((recipe.matchPercent ?? 0) * 100)}% match</span>
                    {recipe.blendSlot ? <span className="tag">blend slot {recipe.blendSlot}</span> : null}
                    {recipe.rankingReason ? <span className="tag">{String(recipe.rankingReason).replaceAll('_', ' ')}</span> : null}
                    {recipe.difficulty && <span className="tag">{recipe.difficulty}</span>}
                    {recipe.cuisine && <span className="tag">{recipe.cuisine}</span>}
                    {recipe.totalMinutes > 0 && <span className="tag">{recipe.totalMinutes} min</span>}
                  </div>
                  <div className="trust-list">
                    <p><strong>Why this result:</strong> {recipe.source === 'llm' ? 'fallback-generated recipe for your pantry profile' : 'database-ranked recipe match'}</p>
                    {Array.isArray(recipe.matchedIngredients) && recipe.matchedIngredients.length > 0 ? (
                      <p><strong>Matched:</strong> {recipe.matchedIngredients.slice(0, 6).join(', ')}</p>
                    ) : null}
                    {Array.isArray(recipe.missingIngredients) && recipe.missingIngredients.length > 0 ? (
                      <p><strong>Missing:</strong> {recipe.missingIngredients.slice(0, 6).join(', ')}</p>
                    ) : null}
                  </div>
                  <div className="recipe-card-footer no-print">
                    <Link aria-label={`View details for ${recipe.name}`} href={`/recipes/${recipe.id}?from=search&source=${encodeURIComponent(recipe.source || '')}&mode=${encodeURIComponent(mode)}&complex=${complex ? 'true' : 'false'}&reason=${encodeURIComponent(recipe.rankingReason || '')}`} className="btn btn-secondary">Details</Link>
                    {session && (
                      <form action={favoriteAction}>
                        <input type="hidden" name="recipeID" value={recipe.id} />
                        <input type="hidden" name="ingredients" value={ingredientsRaw} />
                        <input type="hidden" name="mode" value={mode} />
                        <input type="hidden" name="complex" value={complex ? 'true' : 'false'} />
                        <input type="hidden" name="density" value={density} />
                        <button className="btn btn-primary" type="submit" aria-label={`Save ${recipe.name} to favorites`}>Save Recipe</button>
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
          <div className="result-meta" aria-live="polite">
            <span className="tag">{catalog.total} recipe{catalog.total === 1 ? '' : 's'} available</span>
            <span className="tag">source: {source}</span>
            <span className="tag">sort: {String(sort).replace('_', ' ')}</span>
          </div>
          {catalog.items?.length === 0 ? (
            <div className="status-box">
              <p className="muted">No recipes found for this filter.</p>
            </div>
          ) : (
            <div className="grid-2">
              {catalog.items.map((recipe) => (
                <div className={`card recipe-card recipe-density-${density}`} key={recipe.id}>
                  <div className="recipe-card-header">
                    <div>
                      <strong className="recipe-title">{recipe.name}</strong>
                      <span className="muted" style={{ fontSize: '0.8rem', marginLeft: '0.5rem' }}>{String(recipe.source || '').toUpperCase()}</span>
                    </div>
                  </div>
                  {recipe.description ? <p className="muted" style={{ marginBottom: '0.55rem' }}>{recipe.description}</p> : null}
                  <div className="chip-list" style={{ marginBottom: '0.75rem' }}>
                    <span className="tag">q {(recipe.qualityScore ?? 0).toFixed(3)}</span>
                    {recipe.difficulty && <span className="tag">{recipe.difficulty}</span>}
                    {recipe.cuisine && <span className="tag">{recipe.cuisine}</span>}
                    {recipe.totalMinutes > 0 && <span className="tag">{recipe.totalMinutes} min</span>}
                    {recipe.servings > 0 && <span className="tag">{recipe.servings} servings</span>}
                  </div>
                  <div className="recipe-card-footer no-print">
                    <Link aria-label={`View details for ${recipe.name}`} href={`/recipes/${recipe.id}?from=catalog&source=${encodeURIComponent(recipe.source || '')}&sort=${encodeURIComponent(sort)}&page=${page}`} className="btn btn-secondary">Details</Link>
                    {session && (
                      <form action={favoriteAction}>
                        <input type="hidden" name="recipeID" value={recipe.id} />
                        <input type="hidden" name="ingredients" value={ingredientsRaw} />
                        <input type="hidden" name="mode" value={mode} />
                        <input type="hidden" name="complex" value={complex ? 'true' : 'false'} />
                        <input type="hidden" name="density" value={density} />
                        <button className="btn btn-primary" type="submit" aria-label={`Save ${recipe.name} to favorites`}>Save Recipe</button>
                      </form>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}

          {catalog.total > 0 && (
            <div className="pagination-row">
              {page <= 1 ? (
                <span className="btn btn-secondary btn-disabled" aria-disabled="true">Prev</span>
              ) : (
                <Link className="btn btn-secondary" href={`/recipes?q=${encodeURIComponent(query)}&source=${encodeURIComponent(source)}&sort=${encodeURIComponent(sort)}&density=${encodeURIComponent(density)}&page=${Math.max(1, page - 1)}`}>Prev</Link>
              )}
              <span className="muted">Page {page} of {Math.max(1, Math.ceil((catalog.total || 0) / (catalog.pageSize || 20)))}</span>
              {page >= Math.max(1, Math.ceil((catalog.total || 0) / (catalog.pageSize || 20))) ? (
                <span className="btn btn-secondary btn-disabled" aria-disabled="true">Next</span>
              ) : (
                <Link className="btn btn-secondary" href={`/recipes?q=${encodeURIComponent(query)}&source=${encodeURIComponent(source)}&sort=${encodeURIComponent(sort)}&density=${encodeURIComponent(density)}&page=${Math.min(Math.max(1, Math.ceil((catalog.total || 0) / (catalog.pageSize || 20))), page + 1)}`}>Next</Link>
              )}
            </div>
          )}
        </>
      )}
    </div>
  )
}
