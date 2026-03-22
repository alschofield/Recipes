const { defineConfig, devices } = require('@playwright/test')

const browserChannel = process.env.PLAYWRIGHT_CHANNEL || (process.platform === 'win32' ? 'msedge' : undefined)

const projects = [
  {
    name: 'desktop-chromium',
    use: {
      ...devices['Desktop Chrome'],
      channel: browserChannel,
    },
  },
]

if (process.env.PLAYWRIGHT_INCLUDE_WEBKIT === '1') {
  projects.push({ name: 'mobile-safari', use: { ...devices['iPhone 13'] } })
}

module.exports = defineConfig({
  testDir: './tests/e2e',
  timeout: 60000,
  fullyParallel: true,
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
  },
  projects,
  webServer: {
    command: 'pnpm exec next dev --turbopack',
    url: 'http://localhost:3000',
    reuseExistingServer: true,
    timeout: 120000,
  },
})
