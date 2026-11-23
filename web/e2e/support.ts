import path from "node:path";
import type { APIRequestContext, Locator, Page, PlaywrightWorkerArgs, TestInfo } from "@playwright/test";

const repoRoot = path.resolve(import.meta.dirname, "../..");

/**
 * テストで使うユーザー。実行ごとに別のユーザーにして、前の実行の記録が画面に混ざらないようにする。
 * 表示名はメールアドレスの @ より前になるので、ドメインのほうで分ける。
 */
export function users(testInfo: TestInfo) {
  const run = `${testInfo.project.name}-${testInfo.retry}-${Date.now().toString(36)}`;
  return { sato: `sato@${run}.example`, suzuki: `suzuki@${run}.example` };
}

/** 開発用ログイン（AUTH_MODE=dev）で、email のユーザーとしてセッションの Cookie を入れる（ページを開く前に呼ぶ） */
export async function loginAs(page: Page, email: string): Promise<void> {
  const res = await page.request.post("/api/auth/dev-login", { data: { email } });
  if (!res.ok()) throw new Error(`dev-login failed: ${res.status()}`);
}

/** 画面を使わずに、email のユーザーとして API を呼ぶ（相手側の操作に使う） */
export async function apiAs(
  playwright: PlaywrightWorkerArgs["playwright"],
  baseURL: string | undefined,
  email: string,
): Promise<APIRequestContext> {
  const ctx = await playwright.request.newContext({ baseURL });
  const res = await ctx.post("/api/auth/dev-login", { data: { email } });
  if (!res.ok()) throw new Error(`dev-login failed: ${res.status()}`);
  return ctx;
}

/** QR コード（svg の title で探す） */
export function qr(page: Page, title: string): Locator {
  return page.locator("svg").filter({ has: page.locator("title", { hasText: title }) });
}

/** 画面のスクショを artifacts/screenshots/<light|dark>/<name>.png に保存する */
export async function shot(page: Page, testInfo: TestInfo, name: string): Promise<void> {
  await page.screenshot({
    path: path.join(repoRoot, "artifacts", "screenshots", testInfo.project.name, `${name}.png`),
    animations: "disabled",
  });
}
