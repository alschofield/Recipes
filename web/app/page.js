import Link from 'next/link'
import { getSession } from '../lib/server/session'

export default async function HomePage() {
  const session = await getSession()

  return (
    <div className="page-wrap" style={{ textAlign: 'center', paddingTop: '1.8rem' }}>
      <h1 style={{ fontSize: '2.3rem', marginBottom: '0.75rem' }}>Find recipes from your ingredients</h1>
      <p className="muted" style={{ marginBottom: '2rem', fontSize: '1.1rem' }}>
        Enter what you have. We find what you can make.
      </p>

      <Link href="/recipes" className="btn btn-primary" style={{ fontSize: '1rem', padding: '0.75rem 2rem' }}>
        Search Recipes
      </Link>

      {!session && (
        <p style={{ marginTop: '2rem', fontSize: '0.95rem' }}>
          <Link href="/signup">Create an account</Link> to save favorites and manage your profile.
        </p>
      )}
    </div>
  )
}
