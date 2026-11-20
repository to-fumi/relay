export function ErrorMessage({ error }: { error: unknown }) {
  if (!error) return null;
  const message =
    typeof error === "object" && error !== null && "title" in error ? String(error.title) : "エラーが発生しました";
  return <p role="alert">{message}</p>;
}
