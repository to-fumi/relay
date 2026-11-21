import { useQuery, useQueryClient } from "@tanstack/react-query";
import { CalendarClock, CircleCheck, Package, Plus, RotateCw, User, X } from "lucide-react";
import { useEffect } from "react";
import { listMyHandoffsOptions, listMyOutstandingHandoffsQueryKey } from "../../shared/api";
import { formatTime, useCountdown } from "../../shared/lib";
import { Card, DueBadge, Fab, IconButton, IconRow, IconRows, TokenQr } from "../../shared/ui";

/** 期限が切れたあとも、ぎりぎりで読み取られた分を拾うために少しだけ確認を続ける */
const GRACE_MS = 5000;

type Props = {
  token: string;
  expiresAt: string;
  /** QR の ID。読み取られて記録されると、その記録の jti と一致する */
  jti: string | undefined;
  item: string;
  qty: number;
  dueDate: string | null;
  onReissue: () => void;
  onReset: () => void;
  reissuing: boolean;
};

/**
 * 発行した QR を表示する。相手が読み取ると、その記録を見つけて「完了」に切り替える。
 * 有効な QR が同時に 2 つできないように、再発行は期限が切れてからだけにする。
 */
export function IssuedQr({ token, expiresAt, jti, item, qty, dueDate, onReissue, onReset, reissuing }: Props) {
  const queryClient = useQueryClient();
  const remaining = useCountdown(expiresAt);
  const expired = remaining === 0;
  const watching = jti !== undefined && Date.now() < new Date(expiresAt).getTime() + GRACE_MS;

  // 読み取られたかを確かめるため、自分の記録を取り直す
  const recent = useQuery({
    ...listMyHandoffsOptions({ query: { limit: 20 } }),
    refetchInterval: (q) => (watching && !q.state.data?.some((h) => h.jti === jti) ? 2000 : false),
  });
  const received = recent.data?.find((h) => h.jti === jti);

  useEffect(() => {
    if (received?.dueDate) queryClient.invalidateQueries({ queryKey: listMyOutstandingHandoffsQueryKey() });
  }, [received, queryClient]);

  if (received) {
    return (
      <>
        <Card className="space-y-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-emerald-600 dark:text-emerald-400">
              <CircleCheck className="size-8" aria-hidden />
              <span className="font-bold text-lg">完了</span>
            </div>
            <span className="text-sm text-zinc-500 tabular-nums dark:text-zinc-400">
              {formatTime(received.createdAt)}
            </span>
          </div>
          <IconRows>
            <IconRow icon={User} label="受け取った人">
              {received.to.displayName}
            </IconRow>
            <IconRow icon={Package} label="品目と数量">
              {received.item} <span className="tabular-nums">×{received.qty}</span>
            </IconRow>
            {received.dueDate && (
              <IconRow icon={CalendarClock} label="返却期限">
                <DueBadge dueDate={received.dueDate} className="text-sm" />
              </IconRow>
            )}
          </IconRows>
        </Card>
        <Fab label="次を送る" onClick={onReset}>
          <Plus aria-hidden />
        </Fab>
      </>
    );
  }

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
        <IconButton
          label={expired ? "再発行" : "期限が切れると再発行できます"}
          variant="primary"
          onClick={onReissue}
          disabled={!expired || reissuing}
        >
          <RotateCw className="size-5" aria-hidden />
        </IconButton>
      </div>
    </Card>
  );
}
