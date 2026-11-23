import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LogIn } from "lucide-react";
import { type FormEvent, type ReactNode, useState } from "react";
import { devLoginMutation, getMeOptions, getMeQueryKey } from "../api";
import { problemStatus } from "../lib";
import { Alert, Button, Card, Field } from "../ui";
import { authMode } from "./config";

/**
 * ログインしていれば children を出し、していなければログインの画面を出す。
 * ログインしているかは /me が 401 を返すかで決める（セッションは HttpOnly の Cookie なので、画面からは見えない）。
 */
export function AuthGate({ children }: { children: ReactNode }) {
  const me = useQuery({ ...getMeOptions(), retry: (count, error) => problemStatus(error) !== 401 && count < 3 });
  if (me.isSuccess) return <>{children}</>;
  if (me.isPending) return null;
  if (problemStatus(me.error) === 401) return <LoginScreen />;
  return (
    <Centered>
      <Alert>接続できませんでした</Alert>
    </Centered>
  );
}

const loginErrors: Record<string, string> = {
  forbidden: "このアカウントでは使えません",
  failed: "ログインできませんでした",
};

/** サーバーがログインに失敗して戻したときの ?login=… を読み、URL からは消す */
function takeLoginError(): string | undefined {
  const url = new URL(window.location.href);
  const result = url.searchParams.get("login");
  if (!result) return undefined;
  url.searchParams.delete("login");
  window.history.replaceState(null, "", url);
  return loginErrors[result] ?? loginErrors.failed;
}

function LoginScreen() {
  const [error] = useState(takeLoginError);
  return (
    <Centered>
      <Card className="w-full max-w-sm space-y-4">
        {error && <Alert>{error}</Alert>}
        {authMode === "oidc" ? <OidcLogin /> : <DevLogin />}
      </Card>
      {authMode === "dev" && (
        <p className="mt-4 rounded-full bg-amber-100 px-3 py-1 font-medium text-amber-800 text-xs dark:bg-amber-950 dark:text-amber-300">
          開発用ログイン（AUTH_MODE=dev）
        </p>
      )}
    </Centered>
  );
}

/** サーバーの /api/auth/login へ移ると、Google のログイン画面へリダイレクトされる */
function OidcLogin() {
  return (
    <a
      href="/api/auth/login"
      className="flex w-full items-center justify-center gap-2 rounded-xl bg-indigo-600 px-4 py-3 font-semibold text-white shadow-sm transition hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-indigo-600 focus-visible:outline-offset-2"
    >
      <LogIn className="size-5" aria-hidden />
      Google でログイン
    </a>
  );
}

function DevLogin() {
  const queryClient = useQueryClient();
  const [input, setInput] = useState("");
  const login = useMutation({
    ...devLoginMutation(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: getMeQueryKey() }),
  });

  const onSubmit = (e: FormEvent) => {
    e.preventDefault();
    const email = input.trim();
    if (email) login.mutate({ body: { email } });
  };

  return (
    <form onSubmit={onSubmit} className="space-y-4">
      <Field
        id="dev-email"
        label="メールアドレス"
        type="email"
        autoComplete="email"
        value={input}
        onChange={(e) => setInput(e.target.value)}
        placeholder="you@example.com"
        required
      />
      {login.isError && <Alert>ログインできませんでした</Alert>}
      <Button type="submit" block disabled={login.isPending}>
        ログイン
      </Button>
    </form>
  );
}

function Centered({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-dvh flex-col items-center justify-center px-4">
      <div className="mb-8 text-center">
        <Logo />
      </div>
      {children}
    </div>
  );
}

export function Logo() {
  return (
    <span className="inline-flex items-center gap-2 font-bold text-2xl tracking-tight">
      <span className="grid size-9 place-items-center rounded-xl bg-indigo-600 text-white">
        <svg viewBox="0 0 24 24" className="size-5" fill="none" stroke="currentColor" strokeWidth={2.5} aria-hidden>
          <path d="M4 17 L10 7 L14 13 L20 7" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </span>
      Relay
    </span>
  );
}
