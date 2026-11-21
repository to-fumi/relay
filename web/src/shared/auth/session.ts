import { authMode } from "./config";
import { setDevEmail } from "./dev";
import { reloginEntra } from "./msal";

let redirecting = false;

/**
 * ログインの期限が切れたとき（API が 401 を返したときなど）に呼ぶ。
 * 保存済みのトークンを消して、ログインし直してもらう。何度呼ばれても 1 回だけ行う。
 */
export function expireSession(): void {
  if (redirecting) return;
  redirecting = true;
  if (authMode === "entra") {
    void reloginEntra();
    return;
  }
  setDevEmail(null);
  window.location.reload();
}
