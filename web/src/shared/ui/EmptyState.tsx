import type { LucideIcon } from "lucide-react";
import type { ReactNode } from "react";

export function EmptyState({ icon: Icon, title, children }: { icon: LucideIcon; title: string; children?: ReactNode }) {
  return (
    <div className="flex flex-col items-center px-6 py-14 text-center">
      <div className="mb-4 grid size-14 place-items-center rounded-2xl bg-zinc-100 text-zinc-400 dark:bg-zinc-800 dark:text-zinc-500">
        <Icon className="size-7" aria-hidden />
      </div>
      <p className="font-semibold text-zinc-900 dark:text-zinc-100">{title}</p>
      {children && <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">{children}</p>}
    </div>
  );
}
