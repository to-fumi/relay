import { authMode } from "./config";
import { setDevEmail } from "./dev";
import { msalInstance } from "./msal";

export function logout(): void {
  if (authMode === "entra") {
    void msalInstance.logoutRedirect();
    return;
  }
  setDevEmail(null);
  window.location.assign("/");
}
