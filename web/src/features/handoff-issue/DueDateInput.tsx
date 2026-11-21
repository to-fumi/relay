import { CalendarDays } from "lucide-react";
import { useState } from "react";
import { addDays, cn, formatLongDate, parseDateKey, todayKey } from "../../shared/lib";
import { CalendarSheet } from "./CalendarSheet";

type Preset = "today" | "nextDay" | "twoWeeks" | "custom";

const presets: { value: Exclude<Preset, "custom">; label: string; dueFrom: (today: string) => string }[] = [
  { value: "today", label: "本日", dueFrom: (t) => t },
  { value: "nextDay", label: "翌日", dueFrom: (t) => addDays(t, 1) },
  { value: "twoWeeks", label: "2週間", dueFrom: (t) => addDays(t, 14) },
];

type Props = {
  value: string | null;
  onChange: (v: string | null) => void;
};

const chip =
  "inline-flex h-10 items-center justify-center rounded-full px-3 font-medium text-sm ring-1 ring-inset transition";
const on = "bg-indigo-600 text-white ring-indigo-600";
const off =
  "bg-white text-zinc-700 ring-zinc-300 hover:bg-zinc-50 dark:bg-zinc-950 dark:text-zinc-300 dark:ring-zinc-700 dark:hover:bg-zinc-800";

/**
 * 返却期限の入力（本日・翌日・2週間・カレンダーで任意の日付）。選んでいるものをもう一度押すと、期限なしに戻る。
 * 上の段にチップ、下の段に選んだ期限の年月日を出す。
 */
export function DueDateInput({ value, onChange }: Props) {
  const [preset, setPreset] = useState<Preset | null>(value ? "custom" : null);
  const [calendarOpen, setCalendarOpen] = useState(false);
  const today = todayKey();

  const toggle = (p: (typeof presets)[number]) => {
    if (preset === p.value) {
      setPreset(null);
      onChange(null);
    } else {
      setPreset(p.value);
      onChange(p.dueFrom(today));
    }
  };

  return (
    <div className="space-y-2">
      <div className="flex flex-wrap items-center gap-1.5">
        {presets.map((p) => (
          <button
            key={p.value}
            type="button"
            aria-pressed={preset === p.value}
            onClick={() => toggle(p)}
            className={cn(chip, preset === p.value ? on : off)}
          >
            {p.label}
          </button>
        ))}
        <button
          type="button"
          aria-label="日付を選ぶ"
          aria-pressed={preset === "custom"}
          onClick={() => setCalendarOpen(true)}
          className={cn(chip, "w-10 px-0", preset === "custom" ? on : off)}
        >
          <CalendarDays className="size-5" aria-hidden />
        </button>
      </div>
      <p
        className={cn(
          "pl-1 font-semibold text-sm tabular-nums",
          value ? "text-indigo-600 dark:text-indigo-400" : "text-zinc-300 dark:text-zinc-600",
        )}
      >
        {value ? formatLongDate(parseDateKey(value)) : "—"}
      </p>
      {calendarOpen && (
        <CalendarSheet
          value={value}
          min={today}
          onSelect={(v) => {
            setPreset("custom");
            onChange(v);
          }}
          onClose={() => setCalendarOpen(false)}
        />
      )}
    </div>
  );
}
