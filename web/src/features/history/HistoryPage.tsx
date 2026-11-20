import { useQuery } from "@tanstack/react-query";
import { getMeOptions, listMyHandoffsOptions } from "../../shared/api";
import { ErrorMessage } from "../../shared/ui";

export function HistoryPage() {
  const me = useQuery(getMeOptions());
  const handoffs = useQuery(listMyHandoffsOptions());

  return (
    <section>
      <h2>履歴</h2>
      <ErrorMessage error={handoffs.error} />
      <table>
        <thead>
          <tr>
            <th>日時</th>
            <th>渡した人</th>
            <th>受け取った人</th>
            <th>品目</th>
            <th>数量</th>
          </tr>
        </thead>
        <tbody>
          {handoffs.data?.map((h) => (
            <tr key={h.id} data-direction={h.from.id === me.data?.id ? "out" : "in"}>
              <td>{new Date(h.createdAt).toLocaleString()}</td>
              <td>{h.from.displayName}</td>
              <td>{h.to.displayName}</td>
              <td>{h.item}</td>
              <td>{h.qty}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
