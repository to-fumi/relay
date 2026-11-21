import { useQuery } from "@tanstack/react-query";
import { ArrowDownLeft, ArrowUpRight, CornerDownLeft, Inbox, type LucideIcon, Rows3, Undo2 } from "lucide-react";
import { useMemo, useState } from "react";
import { getMeOptions, type Handoff, listMyHandoffsOptions } from "../../shared/api";
import { cn, dayKey, formatDay, formatTime } from "../../shared/lib";
import { Alert, Card, EmptyState, PageHeader } from "../../shared/ui";

type Filter = "all" | "out" | "in";

const filters: { value: Filter; label: string; icon: LucideIcon }[] = [
  { value: "all", label: "すべて", icon: Rows3 },
  { value: "out", label: "自分から", icon: ArrowUpRight },
  { value: "in", label: "自分へ", icon: ArrowDownLeft },
];

/** 操作の種類。受け渡しの記録 1 件から、送った・受け取った・返した・返ってきたの操作を作る */
type Kind = "sent" | "received" | "returned" | "gotBack";

type Event = {
  key: string;
  kind: Kind;
  at: string;
  handoff: Handoff;
};

const kinds: Record<Kind, { label: string; icon: LucideIcon; tone: string; outgoing: boolean }> = {
  sent: {
    label: "送った",
    icon: ArrowUpRight,
    tone: "bg-indigo-50 text-indigo-600 dark:bg-indigo-950 dark:text-indigo-400",
    outgoing: true,
  },
  received: {
    label: "受け取った",
    icon: ArrowDownLeft,
    tone: "bg-emerald-50 text-emerald-600 dark:bg-emerald-950 dark:text-emerald-400",
    outgoing: false,
  },
  returned: {
    label: "返した",
    icon: Undo2,
    tone: "bg-amber-50 text-amber-600 dark:bg-amber-950 dark:text-amber-400",
    outgoing: true,
  },
  gotBack: {
    label: "返ってきた",
    icon: CornerDownLeft,
    tone: "bg-amber-50 text-amber-600 dark:bg-amber-950 dark:text-amber-400",
    outgoing: false,
  },
};

function toEvents(handoffs: Handoff[], myId: string | undefined): Event[] {
  const events: Event[] = [];
  for (const h of handoffs) {
    const mine = h.from.id === myId;
    events.push({ key: `${h.id}:handoff`, kind: mine ? "sent" : "received", at: h.createdAt, handoff: h });
    if (h.returnedAt) {
      events.push({ key: `${h.id}:return`, kind: mine ? "gotBack" : "returned", at: h.returnedAt, handoff: h });
    }
  }
  return events.sort((a, b) => b.at.localeCompare(a.at));
}

export function HistoryPage() {
  const me = useQuery(getMeOptions());
  const handoffs = useQuery(listMyHandoffsOptions({ query: { limit: 200 } }));
  const [filter, setFilter] = useState<Filter>("all");

  const groups = useMemo(() => {
    const events = toEvents(handoffs.data ?? [], me.data?.id).filter((e) => {
      if (filter === "out") return kinds[e.kind].outgoing;
      if (filter === "in") return !kinds[e.kind].outgoing;
      return true;
    });
    const map = new Map<string, Event[]>();
    for (const e of events) {
      const key = dayKey(e.at);
      map.set(key, [...(map.get(key) ?? []), e]);
    }
    return [...map.entries()];
  }, [handoffs.data, me.data?.id, filter]);

  return (
    <section>
      <PageHeader title="履歴" />

      <div className="mb-6 grid grid-cols-3 rounded-2xl bg-zinc-200/70 p-1 dark:bg-zinc-800" role="tablist">
        {filters.map((f) => (
          <button
            key={f.value}
            type="button"
            role="tab"
            aria-selected={filter === f.value}
            aria-label={f.label}
            title={f.label}
            onClick={() => setFilter(f.value)}
            className={cn(
              "grid place-items-center rounded-xl py-2.5 transition",
              filter === f.value
                ? "bg-white text-zinc-900 shadow-sm dark:bg-zinc-950 dark:text-zinc-100"
                : "text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100",
            )}
          >
            <f.icon className="size-5" aria-hidden />
          </button>
        ))}
      </div>

      {handoffs.isError && <Alert>取得できませんでした</Alert>}

      {handoffs.isSuccess && groups.length === 0 && (
        <Card>
          <EmptyState icon={Inbox} title="なし" />
        </Card>
      )}

      <div className="space-y-8">
        {groups.map(([key, events]) => (
          <div key={key}>
            <h2 className="mb-2 px-1 font-semibold text-sm text-zinc-500 dark:text-zinc-400">
              {formatDay(events[0]?.at ?? key)}
            </h2>
            <Card className="divide-y divide-zinc-100 p-0 dark:divide-zinc-800">
              {events.map((e) => (
                <EventRow key={e.key} event={e} />
              ))}
            </Card>
          </div>
        ))}
      </div>
    </section>
  );
}

/** 1 行目：品目 → 数量、2 行目：相手 → 時刻 */
function EventRow({ event: { kind, at, handoff: h } }: { event: Event }) {
  const k = kinds[kind];
  const partner = kind === "sent" || kind === "gotBack" ? h.to : h.from;
  return (
    <div className="flex items-center gap-4 px-5 py-4">
      <div className={cn("grid size-10 shrink-0 place-items-center rounded-full", k.tone)}>
        <k.icon className="size-5" aria-label={k.label} />
      </div>
      <div className="min-w-0 flex-1 space-y-1">
        <div className="flex items-baseline justify-between gap-3">
          <p className="truncate font-medium">{h.item}</p>
          <span className="shrink-0 font-semibold tabular-nums">×{h.qty}</span>
        </div>
        <div className="flex items-center justify-between gap-3">
          <p className="truncate text-sm text-zinc-500 dark:text-zinc-400">{partner.displayName}</p>
          <span className="shrink-0 text-xs text-zinc-500 tabular-nums dark:text-zinc-400">{formatTime(at)}</span>
        </div>
      </div>
    </div>
  );
}
