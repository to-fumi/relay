import type { ButtonHTMLAttributes } from "react";
import { cn } from "../lib";

/**
 * 画面の主な操作を、右下（下部タブの上）に浮かせて出す丸いボタン。アイコンだけを入れ、label は読み上げに使う。
 * type="submit" にすれば、置いた場所の form を送信できる。
 */
export function Fab({
  label,
  className,
  type = "button",
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { label: string }) {
  return (
    <div className="pointer-events-none fixed inset-x-0 bottom-[calc(5.5rem+env(safe-area-inset-bottom))] z-20">
      <div className="mx-auto flex max-w-lg justify-end px-5">
        <button
          type={type}
          aria-label={label}
          title={label}
          className={cn(
            "pointer-events-auto grid size-16 place-items-center rounded-full bg-indigo-600 text-white shadow-indigo-600/30 shadow-lg transition hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-indigo-600 focus-visible:outline-offset-2 active:scale-95 disabled:bg-zinc-300 disabled:shadow-none dark:disabled:bg-zinc-700 [&>svg]:size-7",
            className,
          )}
          {...props}
        />
      </div>
    </div>
  );
}
