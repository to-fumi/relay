import type { HTMLAttributes } from "react";
import { cn } from "../lib";

export function Card({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        "rounded-2xl bg-white p-5 shadow-sm ring-1 ring-zinc-200 dark:bg-zinc-900 dark:ring-zinc-800",
        className,
      )}
      {...props}
    />
  );
}
