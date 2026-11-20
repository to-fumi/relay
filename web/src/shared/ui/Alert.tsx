import { CircleAlert } from "lucide-react";
import type { ReactNode } from "react";

export function Alert({ children }: { children: ReactNode }) {
  return (
    <div
      role="alert"
      className="flex items-start gap-2 rounded-xl bg-red-50 p-3 text-red-800 text-sm ring-1 ring-red-200 dark:bg-red-950/50 dark:text-red-200 dark:ring-red-900"
    >
      <CircleAlert className="mt-0.5 size-4 shrink-0" aria-hidden />
      <div>{children}</div>
    </div>
  );
}
