import { InteractionType } from "@azure/msal-browser";
import { MsalAuthenticationTemplate, MsalProvider } from "@azure/msal-react";
import { type FormEvent, type ReactNode, useState } from "react";
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
    <form onSubmit={onSubmit} style={{ padding: 16 }}>
      <h1>Relay（開発用ログイン）</h1>
      <input type="email" value={input} onChange={(e) => setInput(e.target.value)} placeholder="you@example.com" />
      <button type="submit">ログイン</button>
    </form>
  );
}
