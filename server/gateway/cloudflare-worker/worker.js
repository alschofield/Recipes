export default {
  async fetch(request, env) {
    const url = new URL(request.url)

    let origin
    if (url.pathname.startsWith('/recipes') || url.pathname.startsWith('/ingredients')) {
      origin = env.RECIPES_ORIGIN
    } else if (url.pathname.startsWith('/users')) {
      origin = env.USERS_ORIGIN
    } else if (url.pathname.startsWith('/favorites')) {
      origin = env.FAVORITES_ORIGIN
    } else {
      return new Response('Not found', { status: 404 })
    }

    if (!origin) {
      return new Response('Gateway origin not configured', { status: 500 })
    }

    const upstream = new URL(url.pathname + url.search, origin)
    const upstreamRequest = new Request(upstream, request)

    upstreamRequest.headers.set('X-Forwarded-Host', url.host)
    upstreamRequest.headers.set('X-Forwarded-Proto', 'https')

    const response = await fetch(upstreamRequest)
    return new Response(response.body, response)
  },
}
