import { ChevronLeft, ChevronRight, X } from "lucide-react";
import { useState } from "react";
import { cn, dayKey, parseDateKey } from "../../shared/lib";

const weekdays = ["日", "月", "火", "水", "木", "金", "土"];

type Props = {
  value: string | null;
  /** 選べる最初の日（YYYY-MM-DD）。これより前は押せない */
  min: string;
  onSelect: (v: string) => void;
  onClose: () => void;
};

/** 下からせり出すカレンダー。日を押すと選んで閉じる。左下の「本日」で今日を選ぶ */
export function CalendarSheet({ value, min, onSelect, onClose }: Props) {
  const start = parseDateKey(value ?? min);
  const [month, setMonth] = useState(() => new Date(start.getFullYear(), start.getMonth(), 1));
  const minMonth = parseDateKey(min);
  const canPrev =
    month.getFullYear() > minMonth.getFullYear() ||
    (month.getFullYear() === minMonth.getFullYear() && month.getMonth() > minMonth.getMonth());

  const daysInMonth = new Date(month.getFullYear(), month.getMonth() + 1, 0).getDate();
  const cells: (string | null)[] = [
    ...Array<null>(month.getDay()).fill(null),
    ...Array.from({ length: daysInMonth }, (_, i) => dayKey(new Date(month.getFullYear(), month.getMonth(), i + 1))),
  ];
  const shift = (n: number) => setMonth(new Date(month.getFullYear(), month.getMonth() + n, 1));
  const pick = (v: string) => {
    onSelect(v);
    onClose();
  };

  return (
    <div className="fixed inset-0 z-30 flex items-end justify-center">
      {/* 後ろを暗くする幕。押すと閉じる */}
      <button
        type="button"
        aria-label="閉じる"
        tabIndex={-1}
        className="absolute inset-0 bg-black/40"
        onClick={onClose}
      />
      <div
        className="relative w-full max-w-lg rounded-t-3xl bg-white p-6 pb-[calc(1.5rem+env(safe-area-inset-bottom))] shadow-xl dark:bg-zinc-900"
        role="dialog"
        aria-modal="true"
        aria-label="返却期限を選ぶ"
      >
        <div className="mb-4 flex items-center justify-between">
          <button
            type="button"
            aria-label="前の月"
            onClick={() => shift(-1)}
            disabled={!canPrev}
            className="grid size-11 place-items-center rounded-full hover:bg-zinc-100 disabled:opacity-30 dark:hover:bg-zinc-800"
          >
            <ChevronLeft className="size-5" />
          </button>
          <p className="font-semibold text-lg tabular-nums">
            {month.getFullYear()}年{month.getMonth() + 1}月
          </p>
          <button
            type="button"
            aria-label="次の月"
            onClick={() => shift(1)}
            className="grid size-11 place-items-center rounded-full hover:bg-zinc-100 dark:hover:bg-zinc-800"
          >
            <ChevronRight className="size-5" />
          </button>
        </div>

        <div className="grid grid-cols-7 gap-1 text-center">
          {weekdays.map((w, i) => (
            <span
              key={w}
              className={cn(
                "pb-2 text-xs",
                i === 0 ? "text-red-500" : i === 6 ? "text-blue-500" : "text-zinc-400 dark:text-zinc-500",
              )}
            >
              {w}
            </span>
          ))}
          {cells.map((d, i) =>
            d === null ? (
              // biome-ignore lint/suspicious/noArrayIndexKey: 月初より前の空きマス
              <span key={`blank-${i}`} />
            ) : (
              <button
                key={d}
                type="button"
                disabled={d < min}
                onClick={() => pick(d)}
                aria-pressed={d === value}
                className={cn(
                  "mx-auto grid size-11 place-items-center rounded-full text-base tabular-nums transition disabled:text-zinc-300 dark:disabled:text-zinc-700",
                  d === value ? "bg-indigo-600 font-semibold text-white" : "hover:bg-zinc-100 dark:hover:bg-zinc-800",
                  d === min &&
                    d !== value &&
                    "font-semibold text-indigo-600 ring-1 ring-indigo-600 ring-inset dark:text-indigo-400",
                )}
              >
                {parseDateKey(d).getDate()}
              </button>
            ),
          )}
        </div>

        <div className="mt-5 flex items-center justify-between">
          <button
            type="button"
            onClick={() => pick(min)}
            className="h-11 rounded-full px-5 font-semibold text-indigo-600 text-sm ring-1 ring-indigo-600 ring-inset hover:bg-indigo-50 dark:text-indigo-400 dark:hover:bg-indigo-950"
          >
            本日
          </button>
          <button
            type="button"
            aria-label="閉じる"
            onClick={onClose}
            className="grid size-11 place-items-center rounded-full text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"
          >
            <X className="size-5" />
          </button>
        </div>
      </div>
    </div>
  );
}
