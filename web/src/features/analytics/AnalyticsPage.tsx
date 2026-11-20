import { ChartNetwork } from "lucide-react";
import { Card, EmptyState, PageHeader } from "../../shared/ui";

export function AnalyticsPage() {
  return (
    <section>
      <PageHeader title="分析" description="全体とグループの受け渡しの流れを表示します。" />
      <Card>
        <EmptyState icon={ChartNetwork} title="準備中">
          集計の内容を決めたら、関係図や集計をここに表示します。
        </EmptyState>
      </Card>
    </section>
  );
}
