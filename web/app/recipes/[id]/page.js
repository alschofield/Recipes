import Link from 'next/link'
import { endpoints, serverGet } from '../../../lib/server/api'
import { getSession } from '../../../lib/server/session'
import { trackEvent } from '../../../lib/server/telemetry'

export default async function RecipeDetailPage({ params, searchParams }) {
  const session = await getSession()
  const { id } = await params
  const qp = await searchParams
  const from = String(qp?.from || '')
  const source = String(qp?.source || '')
  const mode = String(qp?.mode || '')
  const reason = String(qp?.reason || '')

  let recipe = null
  let error = ''
  try {
    recipe = await serverGet(endpoints.recipes, `/recipes/detail/${id}`, session?.token)

    if (from === 'search' || from === 'catalog') {
      trackEvent('recipe_detail_view', {
        recipeId: id,
        from,
        source,
        mode,
        reason,
      })

      if (source === 'llm') {
        trackEvent('fallback_result_engagement', {
          recipeId: id,
          from,
          mode,
          reason,
        })
      }
    }
  } catch (e) {
    error = e.message || 'Could not load recipe details.'
  }

  if (!recipe) {
    return (
      <div className="page-wrap status-box">
        <p className="error-text" style={{ marginBottom: '0.5rem' }}>{error || 'Recipe not found.'}</p>
        <div style={{ display: 'flex', gap: '0.55rem', flexWrap: 'wrap' }}>
          <Link className="btn btn-secondary" href={`/recipes/${id}`}>Retry</Link>
          <Link className="btn btn-secondary" href="/recipes">Back to search</Link>
        </div>
      </div>
    )
  }

  return (
    <div className="page-wrap recipe-detail-shell">
      <div className="card">
        <div className="print-only" style={{ marginBottom: '0.7rem' }}>
          <strong>Ingrediential</strong>
        </div>
        <div className="no-print" style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.9rem' }}>
          <Link className="btn btn-secondary" href="/recipes">Back to Discover</Link>
        </div>

        <h1 style={{ marginBottom: '0.4rem' }}>{recipe.name}</h1>
        {recipe.description && <p className="muted" style={{ marginBottom: '0.7rem' }}>{recipe.description}</p>}
        {!recipe.description && <p className="muted" style={{ marginBottom: '0.7rem' }}>Detailed recipe profile generated for your Ingrediential workflow.</p>}

        <div className="chip-list" style={{ marginBottom: '1rem' }}>
          {recipe.difficulty && <span className="tag">{recipe.difficulty}</span>}
          {recipe.cuisine && <span className="tag">{recipe.cuisine}</span>}
          {recipe.totalMinutes ? <span className="tag">{recipe.totalMinutes} min</span> : null}
          {recipe.servings ? <span className="tag">{recipe.servings} servings</span> : null}
          {recipe.analysis?.status ? <span className="tag">analysis: {recipe.analysis.status}</span> : null}
        </div>

        {recipe.analysis && (
          <>
            <h3 className="detail-heading">Quality Analysis</h3>
            <div className="analysis-grid" style={{ marginBottom: '1rem' }}>
              <div className="analysis-card"><strong>{(recipe.analysis.overallScore ?? 0).toFixed(3)}</strong><p className="muted">Overall</p></div>
              <div className="analysis-card"><strong>{(recipe.analysis.ingredientCoverageScore ?? 0).toFixed(3)}</strong><p className="muted">Ingredient coverage</p></div>
              <div className="analysis-card"><strong>{(recipe.analysis.nutritionBalanceScore ?? 0).toFixed(3)}</strong><p className="muted">Nutrition balance</p></div>
              <div className="analysis-card"><strong>{(recipe.analysis.flavourAlignmentScore ?? 0).toFixed(3)}</strong><p className="muted">Flavor alignment</p></div>
              <div className="analysis-card"><strong>{(recipe.analysis.noveltyScore ?? 0).toFixed(3)}</strong><p className="muted">Novelty</p></div>
            </div>
          </>
        )}

        <div className="detail-sections">
          <section className="card">
            <h3 className="detail-heading">Ingredients</h3>
            <ul style={{ paddingLeft: '1rem' }}>
              {(recipe.ingredients || recipe.matchedIngredients || []).map((item, idx) => (
                <li key={`${item}-${idx}`}>{typeof item === 'string' ? item : item.name}</li>
              ))}
            </ul>
          </section>

          <section className="card">
            <h3 className="detail-heading">Steps</h3>
            <ol style={{ paddingLeft: '1.1rem' }}>
              {(recipe.steps || []).map((step, idx) => (
                <li key={`${step}-${idx}`} style={{ marginBottom: '0.45rem' }}>{step}</li>
              ))}
              {!recipe.steps?.length && <li>No detailed steps available for this recipe source.</li>}
            </ol>
          </section>
        </div>
      </div>
    </div>
  )
}
