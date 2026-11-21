import { CalendarClock, CircleAlert, CircleCheck } from "lucide-react";
import { cn, daysUntil, formatShortDate, parseDateKey } from "../lib";

type Props = {
  /** 返却期限（YYYY-MM-DD） */
  dueDate: string;
  /** 返却された日時。あれば返却済みとして表示する */
  returnedAt?: string;
  className?: string;
};

/** 返却期限を日付と色で表示する（期限切れは赤、今日・明日は黄、それ以外は灰、返却済みは緑）。 */
export function DueBadge({ dueDate, returnedAt, className }: Props) {
  const date = formatShortDate(parseDateKey(dueDate));
  if (returnedAt) {
    return (
      <span className={cn(base, "bg-emerald-50 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300", className)}>
        <CircleCheck className="size-3.5" aria-hidden />
        <span className="sr-only">返却済み</span>
        {date}
      </span>
    );
  }
  const days = daysUntil(dueDate);
  const Icon = days < 0 ? CircleAlert : CalendarClock;
  const tone =
    days < 0
      ? "bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300"
      : days <= 1
        ? "bg-amber-50 text-amber-800 dark:bg-amber-950 dark:text-amber-300"
        : "bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300";
  return (
    <span className={cn(base, tone, className)}>
      <Icon className="size-3.5" aria-hidden />
      <span className="sr-only">{days < 0 ? "期限切れ 返却期限" : "返却期限"}</span>
      {date}
    </span>
  );
}

const base = "inline-flex shrink-0 items-center gap-1 rounded-full px-2 py-0.5 font-medium text-xs tabular-nums";
