import { QRCodeSVG } from "qrcode.react";
import { cn, useCountdown } from "../lib";

type Props = {
  token: string;
  expiresAt: string;
  /** QR の代替テキスト */
  title: string;
};

/** 署名付きトークンの QR と、有効期限までの残り時間（バーと秒数）を表示する。期限が切れたら QR をぼかす。 */
export function TokenQr({ token, expiresAt, title }: Props) {
  const remaining = useCountdown(expiresAt);
  const expired = remaining === 0;

  return (
    <>
      <div className="relative">
        <div
          className={cn("rounded-2xl bg-white p-4 ring-1 ring-zinc-200 transition", expired && "opacity-15 blur-[2px]")}
        >
          {/* 誤り訂正レベルは L（最も荒くなる）。画面表示なので汚れや欠けは考えなくてよい */}
          <QRCodeSVG value={token} level="L" marginSize={0} className="size-64" title={title} />
        </div>
        {expired && (
          <div className="absolute inset-0 grid place-items-center">
            <span className="rounded-full bg-zinc-900 px-4 py-1.5 font-semibold text-sm text-white">期限切れ</span>
          </div>
        )}
      </div>

      <div className="flex w-full max-w-72 items-center gap-3">
        <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-zinc-200 dark:bg-zinc-800">
          <div
            className="h-full rounded-full bg-indigo-600 transition-[width] duration-300"
            style={{ width: `${Math.min(100, (remaining / 60) * 100)}%` }}
          />
        </div>
        <span className="w-8 shrink-0 text-right text-sm text-zinc-500 tabular-nums dark:text-zinc-400">
          {remaining}s
        </span>
      </div>
    </>
  );
}
