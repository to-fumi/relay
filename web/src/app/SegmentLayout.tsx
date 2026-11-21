import type { LucideIcon } from "lucide-react";
import { NavLink, Outlet } from "react-router";
import { cn } from "../shared/lib";

export type Segment = {
  to: string;
  label: string;
  icon: LucideIcon;
  end?: boolean;
  count?: number;
  /** count を赤くする（期限切れがあるとき） */
  alert?: boolean;
};

/** 下部タブの中を、上の切り替えでさらに分ける */
export function SegmentLayout({ label, segments }: { label: string; segments: Segment[] }) {
  return (
    <section>
      <nav
        className="mb-8 grid rounded-2xl bg-zinc-200/70 p-1 dark:bg-zinc-800"
        style={{ gridTemplateColumns: `repeat(${segments.length}, minmax(0, 1fr))` }}
        aria-label={label}
      >
        {segments.map((s) => (
          <NavLink
            key={s.to}
            to={s.to}
            end={s.end}
            className={({ isActive }) =>
              cn(
                "flex items-center justify-center gap-1.5 rounded-xl py-2.5 font-semibold text-sm transition",
                isActive
                  ? "bg-white text-zinc-900 shadow-sm dark:bg-zinc-950 dark:text-zinc-100"
                  : "text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100",
              )
            }
          >
            <s.icon className="size-4" aria-hidden />
            {s.label}
            {s.count !== undefined && s.count > 0 && (
              <span
                className={cn(
                  "min-w-5 rounded-full px-1.5 text-xs tabular-nums",
                  s.alert
                    ? "bg-red-600 text-white"
                    : "bg-zinc-300/80 text-zinc-700 dark:bg-zinc-700 dark:text-zinc-200",
                )}
              >
                {s.count}
              </span>
            )}
          </NavLink>
        ))}
      </nav>
      <Outlet />
    </section>
  );
}
