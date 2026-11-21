import { ChevronRight } from "lucide-react";
import type { ReactNode } from "react";
import type { Handoff } from "../../shared/api";
import { DueBadge } from "../../shared/ui";

/** 1 行目：品目 → 数量、2 行目：相手 → 返却期限。trailing があれば右端に出す */
export function LoanRow({
  handoff: h,
  partner,
  trailing,
}: {
  handoff: Handoff;
  partner: string;
  trailing?: ReactNode;
}) {
  return (
    <div className="flex items-center gap-3 px-5 py-4">
      <div className="min-w-0 flex-1 space-y-1">
        <div className="flex items-baseline justify-between gap-3">
          <p className="truncate font-medium">{h.item}</p>
          <span className="shrink-0 font-semibold tabular-nums">×{h.qty}</span>
        </div>
        <div className="flex items-center justify-between gap-3">
          <p className="truncate text-sm text-zinc-500 dark:text-zinc-400">{partner}</p>
          {h.dueDate && <DueBadge dueDate={h.dueDate} />}
        </div>
      </div>
      {trailing && (
        <span className="flex shrink-0 items-center text-indigo-600 dark:text-indigo-400" aria-hidden>
          {trailing}
          <ChevronRight className="size-4" />
        </span>
      )}
    </div>
  );
}
