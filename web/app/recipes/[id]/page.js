import Link from 'next/link'
import { endpoints, serverGet } from '../../../lib/server/api'
import { getSession } from '../../../lib/server/session'

export default async function RecipeDetailPage({ params }) {
  const session = await getSession()
  const { id } = await params

  let recipe = null
  let error = ''
  try {
    recipe = await serverGet(endpoints.recipes, `/recipes/detail/${id}`, session?.token)
  } catch (e) {
    error = e.message || 'Could not load recipe details.'
  }

  if (!recipe) {
    return (
      <div className="page-wrap status-box">
        <p className="error-text">{error || 'Recipe not found.'}</p>
        <Link className="btn btn-secondary" style={{ marginTop: '0.6rem' }} href="/recipes">Back to search</Link>
      </div>
    )
  }

  return (
    <div className="page-wrap">
      <div className="card">
        <div className="no-print" style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.9rem' }}>
          <Link className="btn btn-secondary" href="/recipes">Back</Link>
        </div>

        <h1 style={{ marginBottom: '0.4rem' }}>{recipe.name}</h1>
        {recipe.description && <p className="muted" style={{ marginBottom: '0.7rem' }}>{recipe.description}</p>}

        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.45rem', marginBottom: '1rem' }}>
          {recipe.difficulty && <span className="tag">{recipe.difficulty}</span>}
          {recipe.cuisine && <span className="tag">{recipe.cuisine}</span>}
          {recipe.totalMinutes ? <span className="tag">{recipe.totalMinutes} min</span> : null}
          {recipe.servings ? <span className="tag">{recipe.servings} servings</span> : null}
          {recipe.analysis?.status ? <span className="tag">analysis: {recipe.analysis.status}</span> : null}
        </div>

        {recipe.analysis && (
          <>
            <h3 style={{ marginBottom: '0.5rem' }}>Quality Analysis</h3>
            <div className="grid-2" style={{ marginBottom: '1rem' }}>
              <div className="card"><strong>{(recipe.analysis.overallScore ?? 0).toFixed(3)}</strong><p className="muted">Overall</p></div>
              <div className="card"><strong>{(recipe.analysis.ingredientCoverageScore ?? 0).toFixed(3)}</strong><p className="muted">Ingredient coverage</p></div>
              <div className="card"><strong>{(recipe.analysis.nutritionBalanceScore ?? 0).toFixed(3)}</strong><p className="muted">Nutrition balance</p></div>
              <div className="card"><strong>{(recipe.analysis.flavourAlignmentScore ?? 0).toFixed(3)}</strong><p className="muted">Flavor alignment</p></div>
              <div className="card"><strong>{(recipe.analysis.noveltyScore ?? 0).toFixed(3)}</strong><p className="muted">Novelty</p></div>
            </div>
          </>
        )}

        <h3 style={{ marginBottom: '0.5rem' }}>Ingredients</h3>
        <ul style={{ paddingLeft: '1rem', marginBottom: '1rem' }}>
          {(recipe.ingredients || recipe.matchedIngredients || []).map((item, idx) => (
            <li key={`${item}-${idx}`}>{typeof item === 'string' ? item : item.name}</li>
          ))}
        </ul>

        <h3 style={{ marginBottom: '0.5rem' }}>Steps</h3>
        <ol style={{ paddingLeft: '1.1rem' }}>
          {(recipe.steps || []).map((step, idx) => (
            <li key={`${step}-${idx}`} style={{ marginBottom: '0.45rem' }}>{step}</li>
          ))}
          {!recipe.steps?.length && <li>No detailed steps available for this recipe source.</li>}
        </ol>
      </div>
    </div>
  )
}
