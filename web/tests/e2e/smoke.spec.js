const { test, expect } = require('@playwright/test')

test.describe('Public route smoke', () => {
  test('home renders Ingrediential hero', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByRole('heading', { name: /cook great meals from what you already have/i })).toBeVisible()
    await expect(page.getByRole('link', { name: /start discovering recipes/i })).toBeVisible()
  })

  test('discover page renders core sections', async ({ page }) => {
    await page.goto('/recipes')
    await expect(page.getByRole('heading', { name: /discover recipes/i })).toBeVisible()
    await expect(page.getByText(/ingredient-first search/i)).toBeVisible()
    await expect(page.getByText(/library browser/i)).toBeVisible()
  })

  test('auth pages render', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByRole('heading', { name: /welcome back/i })).toBeVisible()

    await page.goto('/signup')
    await expect(page.getByRole('heading', { name: /create your ingrediential account/i })).toBeVisible()
  })
})

test('health endpoint returns ok', async ({ request }) => {
  const response = await request.get('/api/health')
  expect(response.ok()).toBeTruthy()
  const json = await response.json()
  expect(json).toHaveProperty('ok', true)
  expect(json).toHaveProperty('service', 'web')
})
