import { expect, test } from "@playwright/test";
import { apiAs, loginAs, qr, shot, users } from "./support";

// 受け渡しから返却までを、2 人（sato と suzuki）で一通り行う。
// 画面で操作するのは 1 人ずつで、相手の QR の読み取りは API で代わりにする（カメラで QR は読めないため）。
test("受け渡しから返却まで", async ({ page, playwright, baseURL }, testInfo) => {
  const u = users(testInfo);
  const satoApi = await apiAs(playwright, baseURL, u.sato);
  const suzukiApi = await apiAs(playwright, baseURL, u.suzuki);
  // ユーザーは初回の /me で登録される
  expect((await satoApi.get("/api/me")).ok()).toBe(true);
  expect((await suzukiApi.get("/api/me")).ok()).toBe(true);

  let handoffToken = "";
  let returnToken = "";

  await test.step("sato：品目・数量・返却期限を入れる", async () => {
    await loginAs(page, u.sato);
    await page.goto("/");
    await page.getByLabel("品目").fill("LANケーブル");
    await page.getByRole("button", { name: "1つ増やす" }).click();
    await page.getByRole("button", { name: "2週間" }).click();
    await expect(page.getByLabel("数量")).toHaveValue("2");
    await shot(page, testInfo, "01-issue-form");
  });

  await test.step("sato：QR を発行する", async () => {
    const issued = page.waitForResponse(
      (r) => r.url().endsWith("/api/handoff-tokens") && r.request().method() === "POST",
    );
    await page.getByRole("button", { name: "QRを発行" }).click();
    handoffToken = (await (await issued).json()).token;
    await expect(qr(page, "受け渡し用のQRコード")).toBeVisible();
    await shot(page, testInfo, "02-issue-qr");
  });

  await test.step("suzuki が読み取ると、sato の画面が完了になる", async () => {
    const res = await suzukiApi.post("/api/handoffs", { data: { token: handoffToken } });
    expect(res.status()).toBe(201);
    await expect(page.getByText("完了")).toBeVisible();
    await expect(page.getByText("suzuki")).toBeVisible();
    await shot(page, testInfo, "03-issue-done");
  });

  await test.step("sato：待ちに入る", async () => {
    await page.goto("/waiting");
    await expect(page.getByText("LANケーブル")).toBeVisible();
    await expect(page.getByText("suzuki")).toBeVisible();
    await expect(page.getByRole("button", { name: "返却QRを読み取る" })).toBeVisible();
    await shot(page, testInfo, "04-waiting");
  });

  await test.step("suzuki：受け取る（カメラ）", async () => {
    await loginAs(page, u.suzuki);
    await page.goto("/receive");
    await expect(page.locator("video")).toBeVisible();
    await shot(page, testInfo, "05-receive");
  });

  await test.step("suzuki：返すに入る", async () => {
    await page.goto("/return");
    await expect(page.getByRole("button", { name: "LANケーブルの返却QRを表示" })).toBeVisible();
    await expect(page.getByText("sato")).toBeVisible();
    await shot(page, testInfo, "06-return-list");
  });

  await test.step("suzuki：返却QRを出す", async () => {
    const issued = page.waitForResponse(
      (r) => r.url().endsWith("/api/return-tokens") && r.request().method() === "POST",
    );
    await page.getByRole("button", { name: "LANケーブルの返却QRを表示" }).click();
    returnToken = (await (await issued).json()).token;
    await expect(qr(page, "返却用のQRコード")).toBeVisible();
    await shot(page, testInfo, "07-return-qr");
  });

  await test.step("sato が読み取ると、suzuki の画面が完了になる", async () => {
    const res = await satoApi.post("/api/returns", { data: { token: returnToken } });
    expect(res.status()).toBe(200);
    await expect(page.getByText("完了")).toBeVisible();
    await shot(page, testInfo, "08-return-done");
  });

  await test.step("sato：履歴に送った・返ってきたが並ぶ", async () => {
    await loginAs(page, u.sato);
    await page.goto("/history");
    await expect(page.getByLabel("送った")).toBeVisible();
    await expect(page.getByLabel("返ってきた")).toBeVisible();
    await shot(page, testInfo, "09-history");
  });
});

// 対面でないとき：借りた側が選んでまとめて返却を送り、貸した側がまとめて承認する。
test("対面でない返却をまとめて送る", async ({ page, playwright, baseURL }, testInfo) => {
  const u = users(testInfo);
  const satoApi = await apiAs(playwright, baseURL, u.sato);
  const suzukiApi = await apiAs(playwright, baseURL, u.suzuki);
  expect((await satoApi.get("/api/me")).ok()).toBe(true);
  expect((await suzukiApi.get("/api/me")).ok()).toBe(true);

  const due = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10);
  for (const item of ["LANケーブル", "モバイルモニター", "HDMIアダプター"]) {
    const issued = await satoApi.post("/api/handoff-tokens", { data: { item, qty: 1, dueDate: due } });
    const res = await suzukiApi.post("/api/handoffs", { data: { token: (await issued.json()).token } });
    expect(res.status()).toBe(201);
  }

  await test.step("suzuki：返すで 2 つ選ぶ", async () => {
    await loginAs(page, u.suzuki);
    await page.goto("/return");
    await page.getByRole("checkbox", { name: /LANケーブル/ }).check();
    await page.getByRole("checkbox", { name: /モバイルモニター/ }).check();
    await expect(page.getByRole("button", { name: "選んだものの返却を送信" })).toBeVisible();
    await shot(page, testInfo, "10-return-select");
  });

  await test.step("suzuki：まとめて送ると、承認待ちになる", async () => {
    await page.getByRole("button", { name: "選んだものの返却を送信" }).click();
    await expect(page.getByText("返却の承認待ち")).toHaveCount(2);
    await expect(page.getByRole("checkbox")).toHaveCount(1);
    await expect(page.getByRole("button", { name: "選んだものの返却を送信" })).toBeHidden();
    await shot(page, testInfo, "11-return-requested");
  });

  await test.step("sato：待ちで、届いたものを選ぶ", async () => {
    await loginAs(page, u.sato);
    await page.goto("/waiting");
    await expect(page.getByRole("checkbox")).toHaveCount(2);
    await page.getByRole("checkbox", { name: /LANケーブル/ }).check();
    await page.getByRole("checkbox", { name: /モバイルモニター/ }).check();
    await shot(page, testInfo, "12-waiting-select");
  });

  await test.step("sato：まとめて承認すると、待ちから消える", async () => {
    await page.getByRole("button", { name: "選んだものの返却を承認" }).click();
    await expect(page.getByText("LANケーブル")).toBeHidden();
    await expect(page.getByText("モバイルモニター")).toBeHidden();
    await expect(page.getByText("HDMIアダプター")).toBeVisible();
    await expect(page.getByRole("button", { name: "返却QRを読み取る" })).toBeVisible();
    await shot(page, testInfo, "13-waiting-confirmed");
  });
});
