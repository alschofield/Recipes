import { NextResponse } from 'next/server'

const SESSION_COOKIE = 'recipes_token'
const ROLE_COOKIE = 'recipes_role'

const protectedPaths = ['/favorites', '/account', '/admin']

export function proxy(request) {
  const { pathname } = request.nextUrl
  const needsAuth = protectedPaths.some((prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`))
  if (!needsAuth) return NextResponse.next()

  const token = request.cookies.get(SESSION_COOKIE)?.value
  if (!token) {
    const url = request.nextUrl.clone()
    url.pathname = '/login'
    url.searchParams.set('next', pathname)
    return NextResponse.redirect(url)
  }

  if (pathname.startsWith('/admin')) {
    const role = request.cookies.get(ROLE_COOKIE)?.value
    if (role !== 'admin') {
      const url = request.nextUrl.clone()
      url.pathname = '/recipes'
      return NextResponse.redirect(url)
    }
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/favorites/:path*', '/account/:path*', '/admin/:path*'],
}
