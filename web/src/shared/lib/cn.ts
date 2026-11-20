import { twMerge } from "tailwind-merge";

/** className を結合する。falsy は無視し、衝突する Tailwind のクラスは後ろを優先する（例：p-5 と p-0 なら p-0）。 */
export const cn = (...classes: Array<string | false | null | undefined>) => twMerge(classes.filter(Boolean).join(" "));
