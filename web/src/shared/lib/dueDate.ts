import { dayKey } from "./datetime";

const DAY_MS = 24 * 60 * 60 * 1000;

/** 今日の日付（ローカル時刻の YYYY-MM-DD） */
export const todayKey = () => dayKey(new Date());

/** YYYY-MM-DD をローカル時刻の 0 時の Date にする */
export function parseDateKey(key: string): Date {
  const [y = 0, m = 1, d = 1] = key.split("-").map(Number);
  return new Date(y, m - 1, d);
}

/** YYYY-MM-DD に n 日を足した日付 */
export function addDays(key: string, n: number): string {
  const d = parseDateKey(key);
  d.setDate(d.getDate() + n);
  return dayKey(d);
}

/** 今日から返却期限までの日数（今日が期限なら 0、過ぎていれば負） */
export function daysUntil(due: string, today = todayKey()): number {
  return Math.round((parseDateKey(due).getTime() - parseDateKey(today).getTime()) / DAY_MS);
}
