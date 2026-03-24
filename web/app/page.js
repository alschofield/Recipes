import Link from 'next/link'
import { getSession } from '../lib/server/session'

export default async function HomePage() {
  const session = await getSession()

  return (
    <div className="page-wrap home-hero-wrap">
      <section className="card home-hero" aria-label="Ingrediential intro">
        <p className="hero-kicker">Ingrediential</p>
        <h1>Cook great meals from what you already have.</h1>
        <p className="hero-copy muted">
          Ingredient-first search, reliable favorites sync, and fast recipe detail workflows for everyday cooking.
        </p>
        <div className="hero-actions">
          <Link href="/recipes" className="btn btn-primary">
            Start Discovering Recipes
          </Link>
          {!session && (
            <Link href="/signup" className="btn btn-secondary">
              Create Free Account
            </Link>
          )}
        </div>
      </section>
    </div>
  )
}
