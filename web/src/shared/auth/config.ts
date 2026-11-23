/** dev：開発用ログイン（メールアドレスを入れるだけ）。oidc：Google などでログインする（サーバー側で受ける） */
export type AuthMode = "dev" | "oidc";

export const authMode: AuthMode = import.meta.env.VITE_AUTH_MODE === "oidc" ? "oidc" : "dev";
