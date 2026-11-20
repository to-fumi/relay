const KEY = "relay.devUserEmail";

/** 開発用の仮ログイン。バックエンドの AUTH_MODE=dev と組み合わせて使う。 */
export function getDevEmail(): string | null {
  try {
    return sessionStorage.getItem(KEY);
  } catch {
    return null;
  }
}

export function setDevEmail(email: string | null): void {
  try {
    if (email) sessionStorage.setItem(KEY, email);
    else sessionStorage.removeItem(KEY);
  } catch {
    // sessionStorage が使えない環境では何もしない
  }
}

export function getDevAccessToken(): string | undefined {
  const email = getDevEmail();
  return email ? `dev:${email}` : undefined;
}
