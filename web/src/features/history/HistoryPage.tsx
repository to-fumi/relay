import { useQuery } from "@tanstack/react-query";
import { ArrowDownLeft, ArrowUpRight, Inbox } from "lucide-react";
import { useMemo, useState } from "react";
import { getMeOptions, type Handoff, listMyHandoffsOptions } from "../../shared/api";
import { cn, dayKey, formatDay, formatTime } from "../../shared/lib";
import { Alert, Card, EmptyState, PageHeader } from "../../shared/ui";

type Filter = "all" | "out" | "in";

const filters: { value: Filter; label: string }[] = [
  { value: "all", label: "すべて" },
  { value: "out", label: "渡した" },
  { value: "in", label: "受け取った" },
];

export function HistoryPage() {
  const me = useQuery(getMeOptions());
  const handoffs = useQuery(listMyHandoffsOptions({ query: { limit: 200 } }));
  const [filter, setFilter] = useState<Filter>("all");

  const groups = useMemo(() => {
    const myId = me.data?.id;
    const rows = (handoffs.data ?? []).filter((h) => {
      if (filter === "out") return h.from.id === myId;
      if (filter === "in") return h.to.id === myId;
      return true;
    });
    const map = new Map<string, Handoff[]>();
    for (const h of rows) {
      const key = dayKey(h.createdAt);
      map.set(key, [...(map.get(key) ?? []), h]);
    }
    return [...map.entries()];
  }, [handoffs.data, me.data?.id, filter]);

  return (
    <section>
      <PageHeader title="履歴" description="自分が渡した・受け取った記録です。" />

      <div className="mb-4 inline-flex rounded-xl bg-zinc-200/70 p-1 dark:bg-zinc-800" role="tablist">
        {filters.map((f) => (
          <button
            key={f.value}
            type="button"
            role="tab"
            aria-selected={filter === f.value}
            onClick={() => setFilter(f.value)}
            className={cn(
              "rounded-lg px-4 py-1.5 font-medium text-sm transition",
              filter === f.value
                ? "bg-white text-zinc-900 shadow-sm dark:bg-zinc-950 dark:text-zinc-100"
                : "text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100",
            )}
          >
            {f.label}
          </button>
        ))}
      </div>

      {handoffs.isError && <Alert>履歴を取得できませんでした。</Alert>}

      {handoffs.isSuccess && groups.length === 0 && (
        <Card>
          <EmptyState icon={Inbox} title="記録はまだありません">
            「渡す」でQRを発行するか、「受け取る」で相手のQRを読み取ると、ここに表示されます。
          </EmptyState>
        </Card>
      )}

      <div className="space-y-6">
        {groups.map(([key, items]) => (
          <div key={key}>
            <h2 className="mb-2 px-1 font-semibold text-sm text-zinc-500 dark:text-zinc-400">
              {formatDay(items[0]?.createdAt ?? key)}
            </h2>
            <Card className="divide-y divide-zinc-100 p-0 dark:divide-zinc-800">
              {items.map((h) => (
                <HistoryRow key={h.id} handoff={h} outgoing={h.from.id === me.data?.id} />
              ))}
            </Card>
          </div>
        ))}
      </div>
    </section>
  );
}

function HistoryRow({ handoff: h, outgoing }: { handoff: Handoff; outgoing: boolean }) {
  const Icon = outgoing ? ArrowUpRight : ArrowDownLeft;
  const partner = outgoing ? h.to : h.from;
  return (
    <div className="flex items-center gap-3 px-4 py-3.5">
      <div
        className={cn(
          "grid size-10 shrink-0 place-items-center rounded-full",
          outgoing
            ? "bg-indigo-50 text-indigo-600 dark:bg-indigo-950 dark:text-indigo-400"
            : "bg-emerald-50 text-emerald-600 dark:bg-emerald-950 dark:text-emerald-400",
        )}
      >
        <Icon className="size-5" aria-label={outgoing ? "渡した" : "受け取った"} />
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate font-medium">{h.item}</p>
        <p className="truncate text-sm text-zinc-500 dark:text-zinc-400">
          {outgoing ? `${partner.displayName} へ渡した` : `${partner.displayName} から受け取った`}
        </p>
      </div>
      <div className="shrink-0 text-right">
        <p className="font-semibold tabular-nums">×{h.qty}</p>
        <p className="text-xs text-zinc-500 tabular-nums dark:text-zinc-400">{formatTime(h.createdAt)}</p>
      </div>
    </div>
  );
}
