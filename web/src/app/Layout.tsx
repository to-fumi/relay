import { useQuery } from "@tanstack/react-query";
import { ArrowDownToLine, ArrowUpFromLine, ChartNetwork, History, LogOut, Settings } from "lucide-react";
import { Suspense } from "react";
import { NavLink, Outlet } from "react-router";
import { getMeOptions } from "../shared/api";
import { Logo, logout } from "../shared/auth";
import { cn } from "../shared/lib";

const tabs = [
  { to: "/", label: "渡す", icon: ArrowUpFromLine, end: true },
  { to: "/receive", label: "受け取る", icon: ArrowDownToLine },
  { to: "/history", label: "履歴", icon: History },
  { to: "/analytics", label: "分析", icon: ChartNetwork },
] as const;

export function Layout() {
  const me = useQuery(getMeOptions());
  const items = me.data?.isAdmin ? [...tabs, { to: "/admin", label: "管理", icon: Settings }] : tabs;

  return (
    <div className="mx-auto flex min-h-dvh max-w-lg flex-col">
      <header className="sticky top-0 z-10 flex items-center justify-between bg-zinc-50/80 px-4 py-3 backdrop-blur dark:bg-zinc-950/80">
        <div className="scale-75 origin-left">
          <Logo />
        </div>
        <div className="flex items-center gap-1">
          {me.data && (
            <span className="max-w-40 truncate text-sm text-zinc-600 dark:text-zinc-400">{me.data.displayName}</span>
          )}
          <button
            type="button"
            onClick={logout}
            className="grid size-10 place-items-center rounded-full text-zinc-500 hover:bg-zinc-200/60 dark:hover:bg-zinc-800"
            aria-label="ログアウト"
          >
            <LogOut className="size-5" />
          </button>
        </div>
      </header>

      <main className="flex-1 px-4 pt-2 pb-28">
        <Suspense fallback={<div className="py-20 text-center text-sm text-zinc-500">読み込み中…</div>}>
          <Outlet />
        </Suspense>
      </main>

      <nav className="fixed inset-x-0 bottom-0 z-10 border-zinc-200 border-t bg-white/90 pb-[env(safe-area-inset-bottom)] backdrop-blur dark:border-zinc-800 dark:bg-zinc-900/90">
        <ul className="mx-auto flex max-w-lg">
          {items.map(({ to, label, icon: Icon, ...rest }) => (
            <li key={to} className="flex-1">
              <NavLink
                to={to}
                end={"end" in rest}
                className={({ isActive }) =>
                  cn(
                    "flex flex-col items-center gap-1 py-2.5 font-medium text-[11px] transition-colors",
                    isActive
                      ? "text-indigo-600 dark:text-indigo-400"
                      : "text-zinc-500 hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-200",
                  )
                }
              >
                <Icon className="size-6" aria-hidden />
                {label}
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>
    </div>
  );
}
