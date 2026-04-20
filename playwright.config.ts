import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: 'tests',
	fullyParallel: false,
	workers: 1,
	retries: process.env.CI ? 2 : 0,
	use: {
		baseURL: 'http://127.0.0.1:5173',
		trace: 'retain-on-failure'
	},
	webServer: [
		{
			command: 'npm run dev:backend',
			url: 'http://127.0.0.1:58080/api/health',
			reuseExistingServer: !process.env.CI,
			timeout: 60_000
		},
		{
			command: 'npm run dev:frontend',
			url: 'http://127.0.0.1:5173/login',
			reuseExistingServer: !process.env.CI,
			timeout: 60_000
		}
	],
	projects: [
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] }
		},
		{
			name: 'firefox',
			use: { ...devices['Desktop Firefox'] }
		},
		{
			name: 'webkit',
			use: { ...devices['Desktop Safari'] }
		}
	]
});
