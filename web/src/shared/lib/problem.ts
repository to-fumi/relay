/** API が返す Problem Details から HTTP ステータスを取り出す。 */
export function problemStatus(error: unknown): number | undefined {
  if (typeof error === "object" && error !== null && "status" in error && typeof error.status === "number") {
    return error.status;
  }
  return undefined;
}

/** API が返す Problem Details から type を取り出す。 */
export function problemType(error: unknown): string | undefined {
  if (typeof error === "object" && error !== null && "type" in error && typeof error.type === "string") {
    return error.type;
  }
  return undefined;
}
