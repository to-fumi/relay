import { useQuery } from "@tanstack/react-query";
import { getMeOptions, listMyOutstandingHandoffsOptions } from "../../shared/api";

/** 返却期限付きでまだ返却されていないものを、自分が返すもの・返ってくるのを待つものに分ける */
export function useOutstanding() {
  const me = useQuery(getMeOptions());
  const outstanding = useQuery(listMyOutstandingHandoffsOptions());
  const all = outstanding.data ?? [];
  return {
    query: outstanding,
    toReturn: all.filter((h) => h.to.id === me.data?.id),
    waiting: all.filter((h) => h.from.id === me.data?.id),
  };
}
