import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type IScannerError, Scanner } from "@yudiel/react-qr-scanner";
import { CircleCheck, ScanLine } from "lucide-react";
import { useState } from "react";
import { listMyHandoffsQueryKey, receiveHandoffMutation } from "../../shared/api";
import { formatDateTime, problemStatus } from "../../shared/lib";
import { Alert, Button, Card, PageHeader } from "../../shared/ui";

// ライブラリの既定値は高さの最小が 640px で、640x480 のカメラが起動しないため、希望値だけを指定する
const cameraConstraints: MediaTrackConstraints = {
  facingMode: "environment",
  width: { ideal: 1280 },
  height: { ideal: 720 },
};

const errorMessages: Record<number, string> = {
  400: "QRが無効か、有効期限が切れています。相手にもう一度発行してもらってください。",
  409: "このQRはすでに読み取られています。",
  422: "自分が発行したQRは読み取れません。",
};

export function ReceivePage() {
  const queryClient = useQueryClient();
  const [cameraError, setCameraError] = useState<IScannerError | null>(null);
  const receive = useMutation({
    ...receiveHandoffMutation(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: listMyHandoffsQueryKey() }),
  });

  const onScan = (token: string) => {
    if (receive.isPending || receive.isSuccess) return;
    receive.mutate({ body: { token } });
  };

  if (receive.data) {
    const h = receive.data;
    return (
      <section>
        <PageHeader title="受け取る" />
        <Card className="flex flex-col items-center gap-4 py-8 text-center">
          <div className="grid size-16 place-items-center rounded-full bg-emerald-100 text-emerald-600 dark:bg-emerald-950 dark:text-emerald-400">
            <CircleCheck className="size-9" aria-hidden />
          </div>
          <div>
            <p className="font-bold text-xl">受け取りました</p>
            <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">{formatDateTime(h.createdAt)}</p>
          </div>
          <dl className="w-full divide-y divide-zinc-100 rounded-xl bg-zinc-50 text-left text-sm dark:divide-zinc-800 dark:bg-zinc-950">
            <Row label="渡した人" value={h.from.displayName} />
            <Row label="品目" value={h.item} />
            <Row label="数量" value={String(h.qty)} />
          </dl>
          <Button block onClick={() => receive.reset()}>
            <ScanLine className="size-5" aria-hidden />
            続けて読み取る
          </Button>
        </Card>
      </section>
    );
  }

  const status = problemStatus(receive.error);

  return (
    <section>
      <PageHeader title="受け取る" description="相手が表示したQRを、枠の中に映してください。" />
      <div className="overflow-hidden rounded-2xl bg-black shadow-sm ring-1 ring-zinc-200 dark:ring-zinc-800">
        <Scanner
          constraints={cameraConstraints}
          onScan={(codes) => codes[0] && onScan(codes[0].rawValue)}
          onError={(e) => {
            console.error("scanner error", e.kind, e.message, e.cause);
            setCameraError(e);
          }}
          paused={receive.isPending}
          components={{ finder: true }}
          styles={{ container: { aspectRatio: "1 / 1" } }}
        />
      </div>
      {cameraError && (
        <div className="mt-4">
          <Alert>カメラを起動できませんでした。ブラウザのカメラの許可を確認してください。（{cameraError.kind}）</Alert>
        </div>
      )}
      {receive.isPending && <p className="mt-4 text-center text-sm text-zinc-500">記録しています…</p>}
      {receive.isError && (
        <div className="mt-4 space-y-3">
          <Alert>{(status && errorMessages[status]) ?? "記録できませんでした。もう一度お試しください。"}</Alert>
          <Button variant="secondary" block onClick={() => receive.reset()}>
            もう一度読み取る
          </Button>
        </div>
      )}
    </section>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between gap-4 px-4 py-3">
      <dt className="text-zinc-500 dark:text-zinc-400">{label}</dt>
      <dd className="truncate font-medium">{value}</dd>
    </div>
  );
}
