import Link from 'next/link'
import { endpoints, serverGet } from '../../../lib/server/api'
import { getSession } from '../../../lib/server/session'

export default async function IngredientDetailPage({ params }) {
  const { id } = await params
  const session = await getSession()

  let ingredient = null
  let error = ''
  try {
    ingredient = await serverGet(endpoints.recipes, `/ingredients/detail/${id}`, session?.token)
  } catch (e) {
    error = e.message || 'Could not load ingredient details.'
  }

  if (!ingredient) {
    return (
      <div className="page-wrap status-box">
        <p className="error-text">{error || 'Ingredient not found.'}</p>
        <Link className="btn btn-secondary" href="/ingredients" style={{ marginTop: '0.6rem' }}>Back to ingredients</Link>
      </div>
    )
  }

  return (
    <div className="page-wrap">
      <div className="card">
        <div className="no-print" style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.9rem' }}>
          <Link className="btn btn-secondary" href="/ingredients">Back</Link>
        </div>

        <h1 style={{ marginBottom: '0.45rem' }}>{ingredient.canonicalName}</h1>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.45rem', marginBottom: '0.95rem' }}>
          <span className="tag">{ingredient.analysisStatus}</span>
          <span className="tag">quality {(ingredient.qualityScore || 0).toFixed(3)}</span>
          <span className="tag">coverage {ingredient.sourceCoverage || 0}</span>
          <span className="tag">recipes {ingredient.recipeCount || 0}</span>
          {ingredient.category ? <span className="tag">{ingredient.category}</span> : null}
          {ingredient.naturalSource ? <span className="tag">{ingredient.naturalSource}</span> : null}
        </div>

        {ingredient.analysisNotes ? (
          <p className="muted" style={{ marginBottom: '0.8rem' }}>{ingredient.analysisNotes}</p>
        ) : null}

        <h3 style={{ marginBottom: '0.5rem' }}>Aliases</h3>
        {Array.isArray(ingredient.aliases) && ingredient.aliases.length ? (
          <div style={{ display: 'flex', gap: '0.4rem', flexWrap: 'wrap' }}>
            {ingredient.aliases.map((alias) => <span key={alias} className="tag">{alias}</span>)}
          </div>
        ) : (
          <p className="muted">No aliases recorded.</p>
        )}
      </div>
    </div>
  )
}
