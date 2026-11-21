import { InteractionType } from "@azure/msal-browser";
import { MsalAuthenticationTemplate, MsalProvider } from "@azure/msal-react";
import { type FormEvent, type ReactNode, useState } from "react";
import { Button, Card, Field } from "../ui";
import { authMode } from "./config";
import { getDevEmail, setDevEmail } from "./dev";
import { loginRequest, msalInstance } from "./msal";

export function AuthGate({ children }: { children: ReactNode }) {
  if (authMode === "entra") {
    return (
      <MsalProvider instance={msalInstance}>
        <MsalAuthenticationTemplate interactionType={InteractionType.Redirect} authenticationRequest={loginRequest}>
          {children}
        </MsalAuthenticationTemplate>
      </MsalProvider>
    );
  }
  return <DevAuthGate>{children}</DevAuthGate>;
}

function DevAuthGate({ children }: { children: ReactNode }) {
  const [email, setEmail] = useState(getDevEmail());
  const [input, setInput] = useState("");

  if (email) return <>{children}</>;

  const onSubmit = (e: FormEvent) => {
    e.preventDefault();
    const value = input.trim();
    if (!value) return;
    setDevEmail(value);
    setEmail(value);
  };

  return (
    <div className="flex min-h-dvh flex-col items-center justify-center px-4">
      <div className="mb-8 text-center">
        <Logo />
      </div>
      <Card className="w-full max-w-sm">
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
          <Button type="submit" block>
            ログイン
          </Button>
        </form>
      </Card>
      <p className="mt-4 rounded-full bg-amber-100 px-3 py-1 font-medium text-amber-800 text-xs dark:bg-amber-950 dark:text-amber-300">
        開発用ログイン（AUTH_MODE=dev）
      </p>
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
