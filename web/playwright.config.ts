import { defineConfig, devices } from "@playwright/test";

/**
 * E2E テスト。ビルドした SPA を API サーバーから配信し（本番と同じ形）、AUTH_MODE=dev の仮ログインで画面を操作する。
 * 実行は `pnpm task e2e`（先に `pnpm task db:up` で PostgreSQL を起動しておく）。
 * 画面のスクショは artifacts/screenshots/<project>/ に保存する（CI では PR にコメントで載せる）。
 */
const port = Number(process.env.RELAY_E2E_PORT ?? 8090);
const dbPort = process.env.RELAY_DB_PORT ?? "5432";

export default defineConfig({
  testDir: "./e2e",
  // 1 つのテストの中で 2 人の操作を順に行うので、テストの間では並べない
  fullyParallel: false,
  workers: 1,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [["github"], ["list"]] : [["list"]],
  timeout: 60_000,
  use: {
    baseURL: `http://localhost:${port}`,
    locale: "ja-JP",
    timezoneId: "Asia/Tokyo",
    trace: "retain-on-failure",
    // 受け取る画面のカメラは、Chromium の偽のカメラ映像で代わりにする
    permissions: ["camera"],
    launchOptions: {
      args: ["--use-fake-ui-for-media-stream", "--use-fake-device-for-media-stream"],
      executablePath: process.env.RELAY_E2E_CHROMIUM || undefined,
    },
  },
  projects: [
    // iPhone の画面の大きさ（390×844）。ライトとダークの両方で撮る
    {
      name: "light",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 390, height: 844 },
        deviceScaleFactor: 2,
        hasTouch: true,
        colorScheme: "light",
      },
    },
    {
      name: "dark",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 390, height: 844 },
        deviceScaleFactor: 2,
        hasTouch: true,
        colorScheme: "dark",
      },
    },
  ],
  webServer: {
    command: "go run ./cmd/relay",
    cwd: "../backend-go",
    url: `http://localhost:${port}/healthz`,
    timeout: 180_000,
    reuseExistingServer: !process.env.CI,
    stdout: "ignore",
    stderr: "pipe",
    env: {
      ADDR: `:${port}`,
      DATABASE_URL: process.env.DATABASE_URL ?? `postgres://relay:relay@localhost:${dbPort}/relay?sslmode=disable`,
      AUTH_MODE: "dev",
      HANDOFF_TOKEN_SECRET: "e2e-secret-e2e-secret-e2e-secret-e2e-secret",
      HANDOFF_TOKEN_TTL: "60s",
      STATIC_DIR: "../web/dist",
    },
  },
});
