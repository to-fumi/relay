export type AuthMode = "dev" | "entra";

export const authMode: AuthMode = import.meta.env.VITE_AUTH_MODE === "entra" ? "entra" : "dev";

export const entraConfig = {
  clientId: import.meta.env.VITE_ENTRA_CLIENT_ID ?? "",
  tenantId: import.meta.env.VITE_ENTRA_TENANT_ID ?? "",
  apiScope: import.meta.env.VITE_ENTRA_API_SCOPE ?? "",
};
