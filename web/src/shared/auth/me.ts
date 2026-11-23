import type { Query } from "@tanstack/react-query";

type AnyQuery = Query<unknown, unknown, unknown, readonly unknown[]>;

import { getMeQueryKey } from "../api";

/** ログインしているかを決める /me のクエリか（401 のときに、ほかのクエリと分けて扱う） */
export function isMeQuery(query: AnyQuery): boolean {
  return JSON.stringify(query.queryKey) === JSON.stringify(getMeQueryKey());
}
