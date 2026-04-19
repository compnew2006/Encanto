import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: 'tests',
	testMatch: /(.+\.)?(test|spec)\.[jt]s/,
	workers: 1,
	use: {
		baseURL: 'http://127.0.0.1:4173',
		trace: 'on-first-retry'
	},
	webServer: [
		{
			command: 'cd ../backend && PORT=18080 go run .',
			url: 'http://127.0.0.1:18080/api/auth/ws-token',
			reuseExistingServer: false,
			timeout: 120_000
		},
		{
			command: 'PUBLIC_API_BASE=http://127.0.0.1:18080 npm run dev -- --host 127.0.0.1 --port 4173',
			url: 'http://127.0.0.1:4173/login',
			reuseExistingServer: false,
			timeout: 120_000
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
