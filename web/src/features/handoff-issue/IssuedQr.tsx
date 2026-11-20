import { RotateCw, X } from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { cn, useCountdown } from "../../shared/lib";
import { Button, Card } from "../../shared/ui";

type Props = {
  token: string;
  expiresAt: string;
  item: string;
  qty: number;
  onReissue: () => void;
  onReset: () => void;
  reissuing: boolean;
};

export function IssuedQr({ token, expiresAt, item, qty, onReissue, onReset, reissuing }: Props) {
  const remaining = useCountdown(expiresAt);
  const expired = remaining === 0;

  return (
    <Card className="flex flex-col items-center gap-5 text-center">
      <div>
        <p className="font-semibold text-xl">{item}</p>
        <p className="mt-0.5 text-zinc-500 dark:text-zinc-400">
          数量 <span className="font-semibold text-zinc-900 tabular-nums dark:text-zinc-100">{qty}</span>
        </p>
      </div>

      <div className="relative">
        <div
          className={cn("rounded-2xl bg-white p-4 ring-1 ring-zinc-200 transition", expired && "opacity-15 blur-[2px]")}
        >
          {/* 誤り訂正レベルは L（最も荒くなる）。画面表示なので汚れや欠けは考えなくてよい */}
          <QRCodeSVG value={token} level="L" marginSize={0} className="size-64" title="受け渡し用のQRコード" />
        </div>
        {expired && (
          <div className="absolute inset-0 grid place-items-center">
            <span className="rounded-full bg-zinc-900 px-4 py-1.5 font-semibold text-sm text-white">期限切れ</span>
          </div>
        )}
      </div>

      {expired ? (
        <p className="text-sm text-zinc-500 dark:text-zinc-400">もう一度発行してください</p>
      ) : (
        <div className="w-full max-w-64">
          <p className="mb-2 font-medium text-sm">相手に読み取ってもらってください</p>
          <div className="flex items-center gap-3">
            <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-zinc-200 dark:bg-zinc-800">
              <div
                className="h-full rounded-full bg-indigo-600 transition-[width] duration-300"
                style={{ width: `${Math.min(100, (remaining / 60) * 100)}%` }}
              />
            </div>
            <span className="shrink-0 text-sm text-zinc-500 tabular-nums dark:text-zinc-400">残り{remaining}秒</span>
          </div>
        </div>
      )}

      <div className="flex w-full gap-2">
        <Button variant="secondary" className="flex-1" onClick={onReset}>
          <X className="size-5" aria-hidden />
          やり直す
        </Button>
        <Button className="flex-1" onClick={onReissue} disabled={reissuing}>
          <RotateCw className="size-5" aria-hidden />
          再発行
        </Button>
      </div>
    </Card>
  );
}
