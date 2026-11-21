import type { LucideIcon } from "lucide-react";
import type { ReactNode } from "react";

/** 左にアイコン、右に値を置く 1 行。label は読み上げ用。 */
export function IconRow({ icon: Icon, label, children }: { icon: LucideIcon; label: string; children: ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-4 px-4 py-3">
      <dt className="text-zinc-400 dark:text-zinc-500">
        <Icon className="size-5" aria-hidden />
        <span className="sr-only">{label}</span>
      </dt>
      <dd className="min-w-0 truncate text-right font-medium">{children}</dd>
    </div>
  );
}

/** IconRow をまとめる枠 */
export function IconRows({ children }: { children: ReactNode }) {
  return (
    <dl className="w-full divide-y divide-zinc-100 rounded-xl bg-zinc-50 text-sm dark:divide-zinc-800 dark:bg-zinc-950">
      {children}
    </dl>
  );
}
