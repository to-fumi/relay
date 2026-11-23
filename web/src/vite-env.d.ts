interface ImportMetaEnv {
  readonly VITE_AUTH_MODE?: "dev" | "oidc";
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
