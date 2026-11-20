const locale = "ja-JP";

const dateTimeFormat = new Intl.DateTimeFormat(locale, {
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
});
const timeFormat = new Intl.DateTimeFormat(locale, { hour: "2-digit", minute: "2-digit" });
const dayFormat = new Intl.DateTimeFormat(locale, { month: "long", day: "numeric", weekday: "short" });
const dayKeyFormat = new Intl.DateTimeFormat("sv-SE", { year: "numeric", month: "2-digit", day: "2-digit" });

/** 例：2026/11/21 09:05 */
export const formatDateTime = (d: Date | string) => dateTimeFormat.format(new Date(d));
/** 例：09:05 */
export const formatTime = (d: Date | string) => timeFormat.format(new Date(d));
/** 例：11月21日(土) */
export const formatDay = (d: Date | string) => dayFormat.format(new Date(d));
/** 日付ごとにまとめるためのキー（ローカル時刻の YYYY-MM-DD） */
export const dayKey = (d: Date | string) => dayKeyFormat.format(new Date(d));
