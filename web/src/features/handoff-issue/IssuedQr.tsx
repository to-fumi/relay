import { RotateCw, X } from "lucide-react";
import { Card, DueBadge, IconButton, TokenQr } from "../../shared/ui";

type Props = {
  token: string;
  expiresAt: string;
  item: string;
  qty: number;
  dueDate: string | null;
  onReissue: () => void;
  onReset: () => void;
  reissuing: boolean;
};

export function IssuedQr({ token, expiresAt, item, qty, dueDate, onReissue, onReset, reissuing }: Props) {
  return (
    <Card className="flex flex-col items-center gap-6">
      <div className="flex w-full items-center justify-between gap-3">
        <p className="min-w-0 truncate font-semibold text-xl">{item}</p>
        <div className="flex shrink-0 items-center gap-2">
          <span className="font-semibold text-xl tabular-nums">×{qty}</span>
          {dueDate && <DueBadge dueDate={dueDate} className="text-sm" />}
        </div>
      </div>

      <TokenQr token={token} expiresAt={expiresAt} title="受け渡し用のQRコード" />

      <div className="flex w-full justify-between">
        <IconButton label="やり直す" onClick={onReset}>
          <X className="size-5" aria-hidden />
        </IconButton>
        <IconButton label="再発行" variant="primary" onClick={onReissue} disabled={reissuing}>
          <RotateCw className="size-5" aria-hidden />
        </IconButton>
      </div>
    </Card>
  );
}
