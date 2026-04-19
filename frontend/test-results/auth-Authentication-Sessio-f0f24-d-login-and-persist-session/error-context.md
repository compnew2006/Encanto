# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: auth.spec.ts >> Authentication & Session Persistence >> should allow valid login and persist session
- Location: tests/auth.spec.ts:15:2

# Error details

```
Error: expect(page).toHaveURL(expected) failed

Expected pattern: /\/chat/
Received string:  "http://localhost:5173/login"
Timeout: 5000ms

Call log:
  - Expect "toHaveURL" with timeout 5000ms
    9 × unexpected value "http://localhost:5173/login"

```

# Page snapshot

```yaml
- generic [ref=e3]:
  - banner [ref=e4]:
    - heading "Encanto Workspace" [level=1] [ref=e6]
  - main [ref=e7]:
    - generic [ref=e9]:
      - heading "Sign in" [level=2] [ref=e10]
      - generic [ref=e11]:
        - generic [ref=e12]:
          - generic [ref=e13]: Email
          - textbox "Email" [active] [ref=e14]
        - generic [ref=e15]:
          - generic [ref=e16]: Password
          - textbox "Password" [ref=e17]: password123
        - paragraph [ref=e18]: "Hint: admin@example.com / password123"
        - button "Sign In" [ref=e19]
```

# Test source

```ts
  1  | import { test, expect } from '@playwright/test';
  2  | 
  3  | test.describe('Authentication & Session Persistence', () => {
  4  | 
  5  | 	test('should reject invalid login', async ({ page }) => {
  6  | 		await page.goto('/login');
  7  | 		await page.fill('input[name="email"]', 'wrong@example.com');
  8  | 		await page.fill('input[name="password"]', 'badpass');
  9  | 		await page.click('button[type="submit"]');
  10 | 
  11 | 		const errorMsg = page.locator('text=Invalid email or password');
  12 | 		await expect(errorMsg).toBeVisible();
  13 | 	});
  14 | 
  15 | 	test('should allow valid login and persist session', async ({ page }) => {
  16 | 		// 1. Login
  17 | 		await page.goto('/login');
  18 | 		await page.fill('input[name="email"]', 'admin@example.com');
  19 | 		await page.fill('input[name="password"]', 'password123');
  20 | 		await page.click('button[type="submit"]');
  21 | 
  22 | 		// 2. Verify redirect and secure view
> 23 | 		await expect(page).toHaveURL(/\/chat/);
     |                      ^ Error: expect(page).toHaveURL(expected) failed
  24 | 		await expect(page.locator('text=Welcome to your Inbox')).toBeVisible();
  25 | 		await expect(page.locator('text=admin@example.com')).toBeVisible();
  26 | 
  27 | 		// 3. Verify session persistence on reload
  28 | 		await page.reload();
  29 | 		await expect(page.locator('text=Welcome to your Inbox')).toBeVisible();
  30 | 
  31 | 		// 4. Test Logout
  32 | 		await page.click('text=Logout');
  33 | 		await expect(page).toHaveURL(/\/login/);
  34 | 
  35 | 		// 5. Test protected route redirect
  36 | 		await page.goto('/chat');
  37 | 		await expect(page).toHaveURL(/\/login/);
  38 | 	});
  39 | });
  40 | 
```