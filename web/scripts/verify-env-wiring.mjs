#!/usr/bin/env node

import fs from 'node:fs'
import path from 'node:path'

function parseEnvFile(filePath) {
  const content = fs.readFileSync(filePath, 'utf8')
  const out = {}
  for (const line of content.split(/\r?\n/)) {
    const raw = line.trim()
    if (!raw || raw.startsWith('#')) continue
    const idx = raw.indexOf('=')
    if (idx <= 0) continue
    const key = raw.slice(0, idx).trim()
    const value = raw.slice(idx + 1).trim()
    out[key] = value
  }
  return out
}

function trimSlash(value) {
  return String(value || '').replace(/\/$/, '')
}

function endpoint(sharedBase, explicitURL, fallbackPort, apiURL) {
  if (sharedBase) return sharedBase
  if (explicitURL) return trimSlash(explicitURL)
  return `${apiURL}:${fallbackPort}`
}

function main() {
  const root = process.cwd()
  const envArg = process.argv[2] || '.env.production.example'
  const envPath = path.resolve(root, envArg)

  if (!fs.existsSync(envPath)) {
    console.error(`env file not found: ${envPath}`)
    process.exit(1)
  }

  const env = parseEnvFile(envPath)
  const apiBase = trimSlash(env.NEXT_PUBLIC_API_BASE_URL || '')
  const apiURL = env.NEXT_PUBLIC_API_URL || 'http://localhost'
  const recipesPort = env.NEXT_PUBLIC_API_RECIPES_PORT || '8081'
  const usersPort = env.NEXT_PUBLIC_API_USERS_PORT || '8082'
  const favoritesPort = env.NEXT_PUBLIC_API_FAVORITES_PORT || '8080'

  const endpoints = {
    recipes: endpoint(apiBase, env.NEXT_PUBLIC_API_RECIPES_URL, recipesPort, apiURL),
    users: endpoint(apiBase, env.NEXT_PUBLIC_API_USERS_URL, usersPort, apiURL),
    favorites: endpoint(apiBase, env.NEXT_PUBLIC_API_FAVORITES_URL, favoritesPort, apiURL),
  }

  const issues = []
  if (!apiBase && !env.NEXT_PUBLIC_API_URL && !env.NEXT_PUBLIC_API_RECIPES_URL) {
    issues.push('No API base routing configured. Set NEXT_PUBLIC_API_BASE_URL or explicit service URLs.')
  }

  if (apiBase && Object.values(endpoints).some((value) => value !== apiBase)) {
    issues.push('Mixed endpoint resolution while NEXT_PUBLIC_API_BASE_URL is set.')
  }

  console.log('Resolved endpoint wiring:')
  console.log(JSON.stringify(endpoints, null, 2))

  if (issues.length) {
    console.error('\nWiring issues:')
    for (const issue of issues) console.error(`- ${issue}`)
    process.exit(2)
  }

  console.log('\nWiring check passed.')
}

main()
