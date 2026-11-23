import { Send } from "lucide-react";
import { type ReactNode, useId } from "react";
import type { Handoff } from "../../shared/api";
import { DueBadge } from "../../shared/ui";

type Props = {
  handoff: Handoff;
  partner: string;
  /** 選べる行にするとき。選べない行は左を空けて、ほかの行とそろえる */
  select?: { checked: boolean; onChange: (checked: boolean) => void };
  /** 左に選ぶ欄を置く一覧か */
  selectable?: boolean;
  /** 右端の操作 */
  action?: ReactNode;
};

/** 1 行目：品目 → 数量、2 行目：相手 → 返却期限。左に選ぶ欄、右端に操作を置ける */
export function LoanRow({ handoff: h, partner, select, selectable, action }: Props) {
  const id = useId();
  const Info = select ? "label" : "div";
  return (
    <div className="flex items-center gap-4 px-5 py-4">
      {select ? (
        <input
          id={id}
          type="checkbox"
          checked={select.checked}
          onChange={(e) => select.onChange(e.target.checked)}
          className="size-5 shrink-0 accent-indigo-600"
        />
      ) : (
        selectable && <span className="size-5 shrink-0" aria-hidden />
      )}
      <Info htmlFor={select ? id : undefined} className="min-w-0 flex-1 space-y-1">
        <span className="flex items-baseline justify-between gap-3">
          <span className="truncate font-medium">{h.item}</span>
          <span className="shrink-0 font-semibold tabular-nums">×{h.qty}</span>
        </span>
        <span className="flex items-center justify-between gap-3">
          <span className="truncate text-sm text-zinc-500 dark:text-zinc-400">{partner}</span>
          <span className="flex shrink-0 items-center gap-1.5">
            {h.returnRequestedAt && <RequestedBadge />}
            {h.dueDate && <DueBadge dueDate={h.dueDate} />}
          </span>
        </span>
      </Info>
      {action}
    </div>
  );
}

/** 対面でない返却が送られて、貸した側の承認待ちであること */
function RequestedBadge() {
  return (
    <span className="inline-flex items-center rounded-full bg-indigo-50 px-2 py-0.5 text-indigo-700 text-xs dark:bg-indigo-950 dark:text-indigo-300">
      <Send className="size-3.5" aria-hidden />
      <span className="sr-only">返却の承認待ち</span>
    </span>
  );
}
