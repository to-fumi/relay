import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { issueHandoffTokenMutation } from "../../shared/api";
import { IssuedQr } from "./IssuedQr";
import { IssueForm } from "./IssueForm";

export function IssuePage() {
  const [item, setItem] = useState("");
  const [qty, setQty] = useState(1);
  const [dueDate, setDueDate] = useState<string | null>(null);
  const issue = useMutation(issueHandoffTokenMutation());

  const submit = () => issue.mutate({ body: { item: item.trim(), qty, dueDate: dueDate ?? undefined } });

  return (
    <section>
      {issue.data ? (
        <IssuedQr
          token={issue.data.token}
          expiresAt={issue.data.expiresAt}
          jti={issue.data.jti}
          item={item.trim()}
          qty={qty}
          dueDate={dueDate}
          onReissue={submit}
          onReset={() => issue.reset()}
          reissuing={issue.isPending}
        />
      ) : (
        <IssueForm
          item={item}
          qty={qty}
          dueDate={dueDate}
          onItemChange={setItem}
          onQtyChange={setQty}
          onDueDateChange={setDueDate}
          onSubmit={submit}
          pending={issue.isPending}
          failed={issue.isError}
        />
      )}
    </section>
  );
}
