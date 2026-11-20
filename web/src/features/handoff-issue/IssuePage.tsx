import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { issueHandoffTokenMutation } from "../../shared/api";
import { PageHeader } from "../../shared/ui";
import { IssuedQr } from "./IssuedQr";
import { IssueForm } from "./IssueForm";

export function IssuePage() {
  const [item, setItem] = useState("");
  const [qty, setQty] = useState(1);
  const issue = useMutation(issueHandoffTokenMutation());

  const submit = () => issue.mutate({ body: { item: item.trim(), qty } });

  return (
    <section>
      <PageHeader title="渡す" description="品目と数量を入力して、QRを発行します。" />
      {issue.data ? (
        <IssuedQr
          token={issue.data.token}
          expiresAt={issue.data.expiresAt}
          item={item.trim()}
          qty={qty}
          onReissue={submit}
          onReset={() => issue.reset()}
          reissuing={issue.isPending}
        />
      ) : (
        <IssueForm
          item={item}
          qty={qty}
          onItemChange={setItem}
          onQtyChange={setQty}
          onSubmit={submit}
          pending={issue.isPending}
          failed={issue.isError}
        />
      )}
    </section>
  );
}
