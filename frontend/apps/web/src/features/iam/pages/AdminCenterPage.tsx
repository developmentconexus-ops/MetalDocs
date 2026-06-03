import { NavLink, Outlet } from "react-router-dom";
import { WorkspaceViewFrame } from "../../../components/WorkspaceViewFrame";
import { useHasCapability } from "../hooks/useHasCapability";
import styles from "./AdminCenterPage.module.css";

type TabConfig = {
  path: string;
  label: string;
  isVisible: boolean;
};

function AdminCenterPage() {
  const canViewUser = useHasCapability("user.view");
  const canViewMembership = useHasCapability("membership.view");
  const canViewMetrics = useHasCapability("metrics.view");
  const canReadAudit = useHasCapability("audit.read");

  const tabs: readonly TabConfig[] = [
    {
      path: "overview",
      label: "Visao geral",
      isVisible: canViewUser || canViewMembership || canViewMetrics,
    },
    {
      path: "people",
      label: "Pessoas",
      isVisible: canViewUser,
    },
    {
      path: "roles",
      label: "Funcoes",
      isVisible: canViewMembership,
    },
    {
      path: "audit",
      label: "Auditoria",
      isVisible: canReadAudit,
    },
    {
      path: "sessions",
      label: "Sessoes",
      isVisible: canViewUser,
    },
    {
      path: "usage",
      label: "Consumo",
      isVisible: canViewMetrics,
    },
  ];

  const visibleTabs = tabs.filter((t) => t.isVisible);

  return (
    <WorkspaceViewFrame
      kicker="IAM"
      title="Central de administracao"
      description="Gerencie usuarios, funcoes, auditoria e consumo do workspace."
      testId="admin-center-page"
    >
      <div className={styles.frame}>
        <nav role="tablist" aria-label="Abas de administracao" className={styles.tabNav}>
          {visibleTabs.map((tab) => (
            <NavLink
              key={tab.path}
              to={tab.path}
              role="tab"
              className={({ isActive }) =>
                isActive
                  ? `${styles.tabLink} ${styles.tabLinkActive}`
                  : styles.tabLink
              }
            >
              {({ isActive }) => (
                <span aria-current={isActive ? "page" : undefined}>{tab.label}</span>
              )}
            </NavLink>
          ))}
        </nav>
        <div className={styles.outlet}>
          <Outlet />
        </div>
      </div>
    </WorkspaceViewFrame>
  );
}

export function Component() {
  return <AdminCenterPage />;
}

export default AdminCenterPage;
