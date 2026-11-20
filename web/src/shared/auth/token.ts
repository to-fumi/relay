import { authMode } from "./config";
import { getDevAccessToken } from "./dev";
import { getEntraAccessToken } from "./msal";

export function getAccessToken(): Promise<string | undefined> {
  return authMode === "entra" ? getEntraAccessToken() : Promise.resolve(getDevAccessToken());
}
