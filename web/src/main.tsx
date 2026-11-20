import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app/App";
import { authMode, initMsal } from "./shared/auth";

const root = document.getElementById("root");
if (!root) throw new Error("#root が見つかりません");

if (authMode === "entra") {
  await initMsal();
}

createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
