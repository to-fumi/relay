import { Settings } from "lucide-react";
import { Card, EmptyState, PageHeader } from "../../shared/ui";

export function AdminPage() {
  return (
    <section>
      <PageHeader title="管理" description="ユーザーとグループを管理します。" />
      <Card>
        <EmptyState icon={Settings} title="準備中">
          ユーザー・グループの管理とログの削除をここで行います。
        </EmptyState>
      </Card>
    </section>
  );
}
