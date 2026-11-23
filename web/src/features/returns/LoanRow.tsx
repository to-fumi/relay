import { Hourglass, Undo2 } from "lucide-react";
import { type ReactNode, useId } from "react";
import type { Handoff } from "../../shared/api";
import { cn } from "../../shared/lib";
import { DueBadge } from "../../shared/ui";

type Props = {
  handoff: Handoff;
  partner: string;
  /** 自分が借りた側（返す）か、貸した側（待ち）か。対面でない返却の状態の出し方が変わる */
  side: "borrower" | "lender";
  /** 選べる行にするとき。選べない行は左を空けて、ほかの行とそろえる */
  select?: { checked: boolean; onChange: (checked: boolean) => void };
  /** 左に選ぶ欄を置く一覧か */
  selectable?: boolean;
  /** 右端の操作 */
  action?: ReactNode;
};

/** 1 行目：品目 → 数量、2 行目：相手 → 返却期限。左に選ぶ欄、右端に操作を置ける */
export function LoanRow({ handoff: h, partner, side, select, selectable, action }: Props) {
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
        {/* 相手の名前を切らないように、入りきらなければ状態と期限を次の行の右に回す */}
        <span className="flex flex-wrap items-center justify-between gap-x-3 gap-y-1.5">
          <span className="max-w-full truncate text-sm text-zinc-500 dark:text-zinc-400">{partner}</span>
          <span className="ml-auto flex shrink-0 items-center gap-1.5">
            {h.returnRequestedAt && <RequestedChip side={side} />}
            {h.dueDate && <DueBadge dueDate={h.dueDate} />}
          </span>
        </span>
      </Info>
      {action}
    </div>
  );
}

/**
 * 対面でない返却が送られて、貸した側の承認を待っていること。
 * 借りた側には「承認待ち」、貸した側には「返却あり」（届いたものを確かめて承認する）と出す。
 */
function RequestedChip({ side }: { side: "borrower" | "lender" }) {
  const { Icon, label, tone } =
    side === "borrower"
      ? {
          Icon: Hourglass,
          label: "承認待ち",
          tone: "bg-zinc-100 text-zinc-700 ring-zinc-300 dark:bg-zinc-800 dark:text-zinc-200 dark:ring-zinc-600",
        }
      : {
          Icon: Undo2,
          label: "返却あり",
          tone: "bg-indigo-50 text-indigo-700 ring-indigo-200 dark:bg-indigo-950 dark:text-indigo-300 dark:ring-indigo-800",
        };
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full px-2 py-0.5 font-semibold text-xs ring-1 ring-inset",
        tone,
      )}
    >
      <Icon className="size-3.5" aria-hidden />
      {label}
    </span>
  );
}
