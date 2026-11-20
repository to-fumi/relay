import { type AuthenticationResult, EventType, PublicClientApplication } from "@azure/msal-browser";
import { entraConfig } from "./config";

export const msalInstance = new PublicClientApplication({
  auth: {
    clientId: entraConfig.clientId,
    authority: `https://login.microsoftonline.com/${entraConfig.tenantId}`,
    redirectUri: window.location.origin,
  },
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

export async function getEntraAccessToken(): Promise<string | undefined> {
  const account = msalInstance.getActiveAccount();
  if (!account) return undefined;
  const result = await msalInstance.acquireTokenSilent({ ...loginRequest, account });
  return result.accessToken;
}
