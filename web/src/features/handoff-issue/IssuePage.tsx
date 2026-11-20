import { useMutation } from "@tanstack/react-query";
import { QRCodeSVG } from "qrcode.react";
import { type FormEvent, useState } from "react";
import { issueHandoffTokenMutation } from "../../shared/api";
import { ErrorMessage } from "../../shared/ui";

export function IssuePage() {
  const [item, setItem] = useState("");
  const [qty, setQty] = useState(1);
  const issue = useMutation(issueHandoffTokenMutation());

  const onSubmit = (e: FormEvent) => {
    e.preventDefault();
    issue.mutate({ body: { item: item.trim(), qty } });
  };

  return (
    <section>
      <h2>渡す</h2>
      <form onSubmit={onSubmit}>
        <label>
          品目
          <input value={item} onChange={(e) => setItem(e.target.value)} required maxLength={200} />
        </label>
        <label>
          数量
          <input type="number" min={1} value={qty} onChange={(e) => setQty(Number(e.target.value))} required />
        </label>
        <button type="submit" disabled={issue.isPending}>
          QRを発行
        </button>
      </form>
      <ErrorMessage error={issue.error} />
      {issue.data && (
        <div>
          <QRCodeSVG value={issue.data.token} size={256} />
          <p>有効期限：{new Date(issue.data.expiresAt).toLocaleTimeString()}</p>
        </div>
      )}
    </section>
  );
}
