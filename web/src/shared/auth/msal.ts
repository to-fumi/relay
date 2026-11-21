import {
  type AuthenticationResult,
  EventType,
  InteractionRequiredAuthError,
  PublicClientApplication,
} from "@azure/msal-browser";
import { entraConfig } from "./config";

export const msalInstance = new PublicClientApplication({
  auth: {
    clientId: entraConfig.clientId,
    authority: `https://login.microsoftonline.com/${entraConfig.tenantId}`,
    redirectUri: window.location.origin,
  },
  // タブを閉じれば消える sessionStorage に置く
  cache: { cacheLocation: "sessionStorage" },
});

export const loginRequest = { scopes: [entraConfig.apiScope] };

export async function initMsal(): Promise<void> {
  await msalInstance.initialize();
  const result = await msalInstance.handleRedirectPromise();
  if (result?.account) {
    msalInstance.setActiveAccount(result.account);
  }
  msalInstance.addEventCallback((event) => {
    if (event.eventType === EventType.LOGIN_SUCCESS) {
      const payload = event.payload as AuthenticationResult | null;
      if (payload?.account) msalInstance.setActiveAccount(payload.account);
    }
  });
  const [first] = msalInstance.getAllAccounts();
  if (!msalInstance.getActiveAccount() && first) {
    msalInstance.setActiveAccount(first);
  }
}

/**
 * API 用のアクセストークンを返す。期限が近いものは MSAL がリフレッシュトークンで自動で更新する。
 * 更新できない（リフレッシュトークンも期限切れなど）ときは、保存済みのトークンを消してログインし直す。
 */
export async function getEntraAccessToken(): Promise<string | undefined> {
  const account = msalInstance.getActiveAccount();
  if (!account) return undefined;
  try {
    const result = await msalInstance.acquireTokenSilent({ ...loginRequest, account });
    return result.accessToken;
  } catch (e) {
    if (e instanceof InteractionRequiredAuthError) {
      await reloginEntra();
      return undefined;
    }
    throw e;
  }
}

/** 保存済みのトークンとアカウントを消し、Entra ID のログイン画面へ移る。 */
export async function reloginEntra(): Promise<void> {
  await msalInstance.clearCache();
  await msalInstance.loginRedirect(loginRequest);
}
