import { CalendarDays } from "lucide-react";
import { useState } from "react";
import { addDays, cn, formatShortDate, parseDateKey, todayKey } from "../../shared/lib";

type Preset = "tomorrow" | "week" | "custom";

type Props = {
  value: string | null;
  onChange: (v: string | null) => void;
};

const chip =
  "inline-flex h-10 items-center justify-center rounded-full px-4 font-medium text-sm ring-1 ring-inset transition";
const on = "bg-indigo-600 text-white ring-indigo-600";
const off =
  "bg-white text-zinc-700 ring-zinc-300 hover:bg-zinc-50 dark:bg-zinc-950 dark:text-zinc-300 dark:ring-zinc-700 dark:hover:bg-zinc-800";

/**
 * 返却期限の入力（明日・1週間・任意の日付）。選んでいるものをもう一度押すと、期限なしに戻る。
 * 右端に、選んだ期限の日付を出す。
 */
export function DueDateInput({ value, onChange }: Props) {
  const [preset, setPreset] = useState<Preset | null>(value ? "custom" : null);
  const today = todayKey();

  const toggle = (p: Exclude<Preset, "custom">, days: number) => {
    if (preset === p) {
      setPreset(null);
      onChange(null);
    } else {
      setPreset(p);
      onChange(addDays(today, days));
    }
  };

  return (
    <div className="flex items-center gap-2">
      <button
        type="button"
        aria-pressed={preset === "tomorrow"}
        onClick={() => toggle("tomorrow", 1)}
        className={cn(chip, preset === "tomorrow" ? on : off)}
      >
        明日
      </button>
      <button
        type="button"
        aria-pressed={preset === "week"}
        onClick={() => toggle("week", 7)}
        className={cn(chip, preset === "week" ? on : off)}
      >
        1週間
      </button>
      {/* 透明な date 入力を重ねて、押すとカレンダーが開くようにする */}
      <label className={cn(chip, "relative w-10 px-0", preset === "custom" ? on : off)} title="日付を選ぶ">
        <CalendarDays className="size-5" aria-hidden />
        <input
          type="date"
          aria-label="返却期限の日付を選ぶ"
          min={today}
          value={preset === "custom" ? (value ?? "") : ""}
          onChange={(e) => {
            const v = e.target.value || null;
            setPreset(v ? "custom" : null);
            onChange(v);
          }}
          className="absolute inset-0 cursor-pointer opacity-0"
        />
      </label>
      <span
        className={cn(
          "ml-auto shrink-0 font-semibold text-sm tabular-nums",
          value ? "text-indigo-600 dark:text-indigo-400" : "text-zinc-300 dark:text-zinc-600",
        )}
      >
        {value ? formatShortDate(parseDateKey(value)) : "—"}
      </span>
    </div>
  );
}
