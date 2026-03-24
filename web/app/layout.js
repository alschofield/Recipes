import './globals.css'
import Link from 'next/link'
import { redirect } from 'next/navigation'
import { IBM_Plex_Mono, Plus_Jakarta_Sans, Space_Grotesk } from 'next/font/google'
import { clearSession, getSession } from '../lib/server/session'

const headingFont = Space_Grotesk({
  subsets: ['latin'],
  variable: '--font-heading',
  weight: ['500', '600', '700'],
})

const bodyFont = Plus_Jakarta_Sans({
  subsets: ['latin'],
  variable: '--font-body',
  weight: ['400', '500', '600', '700'],
})

const monoFont = IBM_Plex_Mono({
  subsets: ['latin'],
  variable: '--font-mono',
  weight: ['400', '500'],
})

export const metadata = {
  title: 'Ingrediential',
  description: 'Cook smarter with ingredient-first recipe search',
}

export const viewport = {
  themeColor: '#ad3f2f',
}

export default async function RootLayout({ children }) {
  const session = await getSession()

  async function logoutAction() {
    'use server'
    await clearSession()
    redirect('/')
  }

  return (
    <html lang="en">
      <body className={`${headingFont.variable} ${bodyFont.variable} ${monoFont.variable}`}>
        <div className="app-shell">
          <a className="skip-link" href="#main-content">Skip to main content</a>
          <aside className="app-sidebar" aria-label="Sidebar">
            <Link href="/" className="brand">Ingrediential</Link>
            <p className="muted" style={{ marginTop: '0.5rem', fontSize: '0.83rem' }}>Ingredient-first cooking studio</p>
            <nav className="side-nav" aria-label="Primary">
              <Link href="/">Home</Link>
              <Link href="/recipes">Recipes</Link>
              {session && <Link href="/ingredients">Ingredients</Link>}
              {session && <Link href="/favorites">Favorites</Link>}
              {session && <Link href="/account">Account</Link>}
            </nav>
            {session?.role === 'admin' && (
              <>
                <p className="side-nav-label">Admin</p>
                <nav className="side-nav" aria-label="Admin">
                  <Link href="/admin/ingredients">Ingredients</Link>
                  <Link href="/admin/recipes">Recipes</Link>
                </nav>
              </>
            )}
          </aside>
          <div className="app-frame">
            <header className="app-topbar" role="banner">
              <div className="topbar-title-wrap">
                <p className="topbar-kicker">Ingrediential V1</p>
                <p className="topbar-title">Plan less. Cook better. Waste less.</p>
              </div>
              <div className="user-actions">
              {session ? (
                <>
                  <span className="muted">{session.username}</span>
                  <form action={logoutAction}>
                    <button className="btn btn-secondary" type="submit">Logout</button>
                  </form>
                </>
              ) : (
                <>
                  <Link href="/login" className="btn btn-secondary">Login</Link>
                  <Link href="/signup" className="btn btn-primary">Sign Up</Link>
                </>
              )}
              </div>
            </header>
            <main id="main-content" className="app-main" role="main">{children}</main>
          </div>
        </div>
      </body>
    </html>
  )
}
