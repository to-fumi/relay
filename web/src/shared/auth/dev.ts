const KEY = "relay.devSession";
/** 開発用ログインの有効期間。本番の Entra ID と同じく、期限が切れたらログインし直す */
const TTL_MS = 12 * 60 * 60 * 1000;

type DevSession = { email: string; expiresAt: number };

/** 開発用の仮ログイン。バックエンドの AUTH_MODE=dev と組み合わせて使う。期限切れのものは読むときに消す。 */
export function getDevEmail(): string | null {
  try {
    const raw = sessionStorage.getItem(KEY);
    if (!raw) return null;
    const s = JSON.parse(raw) as Partial<DevSession>;
    if (typeof s.email !== "string" || typeof s.expiresAt !== "number" || s.expiresAt <= Date.now()) {
      sessionStorage.removeItem(KEY);
      return null;
    }
    return s.email;
  } catch {
    return null;
  }
}

export function setDevEmail(email: string | null): void {
  try {
    if (email) sessionStorage.setItem(KEY, JSON.stringify({ email, expiresAt: Date.now() + TTL_MS }));
    else sessionStorage.removeItem(KEY);
  } catch {
    // sessionStorage が使えない環境では何もしない
  }
}

export function getDevAccessToken(): string | undefined {
  const email = getDevEmail();
  return email ? `dev:${email}` : undefined;
}
