import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, CornerDownLeft, Package, ScanLine, User } from "lucide-react";
import { listMyHandoffsQueryKey, listMyOutstandingHandoffsQueryKey, returnHandoffMutation } from "../../shared/api";
import { formatTime, problemStatus } from "../../shared/lib";
import { Alert, Card, Fab, IconButton, IconRow, IconRows } from "../../shared/ui";
import { QrScanner } from "../../shared/ui/QrScanner";

const errorMessages: Record<number, string> = {
  400: "無効、または期限切れのQRです",
  403: "あなたが送ったものではありません",
  404: "見つかりません",
  409: "返却済みです",
};

/** 待ち：相手の返却QRを読み取って、返却を記録する */
export function ReturnScan({ onBack }: { onBack: () => void }) {
  const queryClient = useQueryClient();
  const ret = useMutation({
    ...returnHandoffMutation(),
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: listMyHandoffsQueryKey() }),
        queryClient.invalidateQueries({ queryKey: listMyOutstandingHandoffsQueryKey() }),
      ]),
  });

  const onScan = (token: string) => {
    if (ret.isPending || ret.isSuccess) return;
    ret.mutate({ body: { token } });
  };

  if (ret.data) {
    const h = ret.data;
    return (
      <>
        <Card className="space-y-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-amber-600 dark:text-amber-400">
              <CornerDownLeft className="size-8" aria-hidden />
              <span className="font-bold text-lg">返却</span>
            </div>
            {h.returnedAt && (
              <span className="text-sm text-zinc-500 tabular-nums dark:text-zinc-400">{formatTime(h.returnedAt)}</span>
            )}
          </div>
          <IconRows>
            <IconRow icon={User} label="返した人">
              {h.to.displayName}
            </IconRow>
            <IconRow icon={Package} label="品目と数量">
              {h.item} <span className="tabular-nums">×{h.qty}</span>
            </IconRow>
          </IconRows>
          <IconButton label="一覧へ" onClick={onBack}>
            <ArrowLeft className="size-5" aria-hidden />
          </IconButton>
        </Card>
        <Fab label="続けて読み取る" onClick={() => ret.reset()}>
          <ScanLine aria-hidden />
        </Fab>
      </>
    );
  }

  const status = problemStatus(ret.error);
  return (
    <div className="space-y-6">
      <QrScanner onScan={onScan} paused={ret.isPending} />
      {ret.isError && <Alert>{(status && errorMessages[status]) ?? "記録できませんでした"}</Alert>}
      <div className="flex justify-between">
        <IconButton label="一覧へ" onClick={onBack}>
          <ArrowLeft className="size-5" aria-hidden />
        </IconButton>
        {ret.isError && (
          <Fab label="もう一度読み取る" onClick={() => ret.reset()}>
            <ScanLine aria-hidden />
          </Fab>
        )}
      </div>
    </div>
  );
}
