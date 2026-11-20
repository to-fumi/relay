import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Scanner } from "@yudiel/react-qr-scanner";
import { listMyHandoffsQueryKey, receiveHandoffMutation } from "../../shared/api";
import { ErrorMessage } from "../../shared/ui";

export function ReceivePage() {
  const queryClient = useQueryClient();
  const receive = useMutation({
    ...receiveHandoffMutation(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: listMyHandoffsQueryKey() }),
  });

  const onScan = (token: string) => {
    if (receive.isPending || receive.isSuccess) return;
    receive.mutate({ body: { token } });
  };

  return (
    <section>
      <h2>受け取る</h2>
      {receive.data ? (
        <div>
          <p>
            {receive.data.from.displayName} から {receive.data.item} × {receive.data.qty} を受け取りました
          </p>
          <button type="button" onClick={() => receive.reset()}>
            続けて読み取る
          </button>
        </div>
      ) : (
        <Scanner onScan={(codes) => codes[0] && onScan(codes[0].rawValue)} />
      )}
      <ErrorMessage error={receive.error} />
    </section>
  );
}
