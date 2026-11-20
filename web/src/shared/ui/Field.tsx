import type { InputHTMLAttributes, ReactNode } from "react";
import { cn } from "../lib";

export const inputClass =
  "block h-12 w-full rounded-xl border-0 bg-white px-4 text-base text-zinc-900 ring-1 ring-zinc-300 ring-inset placeholder:text-zinc-400 focus:ring-2 focus:ring-indigo-600 focus:outline-none dark:bg-zinc-950 dark:text-zinc-100 dark:ring-zinc-700";

type Props = InputHTMLAttributes<HTMLInputElement> & { label: ReactNode };

export function Field({ label, className, id, ...props }: Props) {
  return (
    <label className="block" htmlFor={id}>
      <span className="mb-1.5 block font-medium text-sm text-zinc-700 dark:text-zinc-300">{label}</span>
      <input id={id} className={cn(inputClass, className)} {...props} />
    </label>
  );
}
