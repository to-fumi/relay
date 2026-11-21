import { useQuery } from "@tanstack/react-query";
import { ArrowRightLeft, History, Undo2 } from "lucide-react";
import { Suspense } from "react";
import { NavLink, Outlet, useLocation } from "react-router";
import { getMeOptions, listMyOutstandingHandoffsOptions } from "../shared/api";
import { Logo } from "../shared/auth";
import { cn, daysUntil } from "../shared/lib";

// paths：そのタブを選択中として表示するパス（タブの中の切り替え先も含める）
const tabs = [
  { to: "/", label: "受け渡し", icon: ArrowRightLeft, paths: ["/", "/receive"] },
  { to: "/return", label: "返却", icon: Undo2, paths: ["/return", "/waiting"] },
  { to: "/history", label: "履歴", icon: History, paths: ["/history"] },
];

export function Layout() {
  const me = useQuery(getMeOptions());
  const outstanding = useQuery(listMyOutstandingHandoffsOptions());
  const { pathname } = useLocation();
  // 返却期限を過ぎているもの（返す・待ち）の数を返却タブに表示する
  const overdue = (outstanding.data ?? []).filter((h) => h.dueDate && daysUntil(h.dueDate) < 0).length;

  return (
    <div className="mx-auto flex min-h-dvh max-w-lg flex-col">
      <header className="sticky top-0 z-10 flex h-16 items-center justify-between bg-zinc-50/80 px-5 backdrop-blur dark:bg-zinc-950/80">
        <div className="origin-left scale-75">
          <Logo />
        </div>
        {me.data && (
          <span className="max-w-48 truncate text-sm text-zinc-500 dark:text-zinc-400">{me.data.displayName}</span>
        )}
      </header>

      {/* 下の余白は、下部タブと右下のボタン（Fab）に隠れないようにとる */}
      <main className="flex-1 px-5 pt-4 pb-48">
        <Suspense fallback={<div className="py-20 text-center text-sm text-zinc-500">読み込み中…</div>}>
          <Outlet />
        </Suspense>
      </main>

      <nav className="fixed inset-x-0 bottom-0 z-10 border-zinc-200 border-t bg-white/90 pb-[env(safe-area-inset-bottom)] backdrop-blur dark:border-zinc-800 dark:bg-zinc-900/90">
        <ul className="mx-auto flex max-w-lg">
          {tabs.map(({ to, label, icon: Icon, paths }) => (
            <li key={to} className="flex-1">
              <NavLink
                to={to}
                className={() => {
                  const isActive = paths.includes(pathname);
                  return cn(
                    "flex flex-col items-center gap-1 py-2.5 font-medium text-[11px] transition-colors",
                    isActive
                      ? "text-indigo-600 dark:text-indigo-400"
                      : "text-zinc-500 hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-200",
                  );
                }}
              >
                <span className="relative">
                  <Icon className="size-6" aria-hidden />
                  {to === "/return" && overdue > 0 && (
                    <span className="-top-1.5 -right-2.5 absolute min-w-4 rounded-full bg-red-600 px-1 text-center font-bold text-[10px] text-white leading-4 tabular-nums">
                      {overdue}
                      <span className="sr-only">件が期限切れ</span>
                    </span>
                  )}
                </span>
                {label}
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>
    </div>
  );
}
