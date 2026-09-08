import { test, expect } from '@playwright/test'

const testEmail    = `e2e-${Date.now()}@test.inferroute.dev`
const testPassword = 'testpassword123'

test.describe('Authentication flow', () => {
  test('registers a new account', async ({ page }) => {
    await page.goto('/register')
    await page.fill('input[type="email"]',    testEmail)
    await page.fill('input[type="password"]', testPassword)
    await page.click('button[type="submit"]')
    await expect(page).toHaveURL('/')
  })

  test('logs in with existing credentials', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[type="email"]',    testEmail)
    await page.fill('input[type="password"]', testPassword)
    await page.click('button[type="submit"]')
    await expect(page).toHaveURL('/')
    await expect(page.locator('text=Dashboard')).toBeVisible()
  })

  test('creates an API key', async ({ page }) => {
    // Assume already logged in via localStorage (or re-login)
    await page.goto('/login')
    await page.fill('input[type="email"]',    testEmail)
    await page.fill('input[type="password"]', testPassword)
    await page.click('button[type="submit"]')
    await page.waitForURL('/')

    await page.goto('/keys')
    await page.fill('input[placeholder*="Key name"]', 'e2e-test-key')
    await page.click('button:has-text("Create key")')

    const keyDisplay = page.locator('code')
    await expect(keyDisplay).toBeVisible()
    const rawKey = await keyDisplay.textContent()
    expect(rawKey).toMatch(/^ir_live_/)
  })

  test('views worker health on workers page', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[type="email"]',    testEmail)
    await page.fill('input[type="password"]', testPassword)
    await page.click('button[type="submit"]')
    await page.waitForURL('/')

    await page.goto('/workers')
    await expect(page.locator('text=Workers')).toBeVisible()
  })

  test('redirects to login when not authenticated', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveURL('/login')
  })
})
