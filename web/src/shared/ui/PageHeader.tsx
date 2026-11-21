import type { ReactNode } from "react";

/** 画面の見出し。左にタイトル、右に操作（あれば）を置く。 */
export function PageHeader({ title, action }: { title: string; action?: ReactNode }) {
  return (
    <div className="mb-8 flex min-h-11 items-center justify-between gap-4">
      <h1 className="font-bold text-3xl tracking-tight">{title}</h1>
      {action}
    </div>
  );
}
