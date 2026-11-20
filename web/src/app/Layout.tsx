import { NavLink, Outlet } from "react-router";

export function Layout() {
  return (
    <div>
      <header>
        <nav style={{ display: "flex", gap: 12 }}>
          <NavLink to="/">渡す</NavLink>
          <NavLink to="/receive">受け取る</NavLink>
          <NavLink to="/history">履歴</NavLink>
          <NavLink to="/analytics">分析</NavLink>
          <NavLink to="/admin">管理</NavLink>
        </nav>
      </header>
      <main>
        <Outlet />
      </main>
    </div>
  );
}
