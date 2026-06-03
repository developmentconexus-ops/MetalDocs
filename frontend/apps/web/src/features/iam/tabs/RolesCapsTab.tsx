import { useMemo, useState } from "react";
import RoleCapabilityMatrix, {
  type SelectedCell,
} from "../components/RoleCapabilityMatrix";
import {
  CAPABILITY_DOMAINS,
  GRADE_LABEL,
  GRADE_ORDER,
  getCapabilityDomain,
  getCapabilityGrade,
  type CapabilityDomain,
  type CapabilityGrade,
} from "../presenters/capability-grade-presenter";
import { getRoleLabel } from "../presenters/role-label-presenter";
import { useRolesQuery } from "../queries/useRolesQuery";
import { useCapabilitiesQuery } from "../queries/useCapabilitiesQuery";
import { useRoleCapabilitiesQuery } from "../queries/useRoleCapabilitiesQuery";
import { useKpiQuery } from "../queries/useKpiQuery";
import styles from "./RolesCapsTab.module.css";

type Selection =
  | { kind: "none" }
  | { kind: "role"; role: string }
  | { kind: "cell"; role: string; domain: CapabilityDomain }
  | { kind: "capability"; code: string };

interface CapabilityDescriptor {
  code: string;
  description: string;
  category: string;
}

interface RoleCapabilityLink {
  role: string;
  capability: string;
}

function groupByDomain(
  codes: ReadonlyArray<string>,
  capByCode: Map<string, CapabilityDescriptor>,
): Map<CapabilityDomain, CapabilityDescriptor[]> {
  const groups = new Map<CapabilityDomain, CapabilityDescriptor[]>();
  for (const code of codes) {
    const domain = getCapabilityDomain(code);
    if (!domain) continue;
    const desc = capByCode.get(code) ?? {
      code,
      description: code,
      category: domain,
    };
    const bucket = groups.get(domain) ?? [];
    bucket.push(desc);
    groups.set(domain, bucket);
  }
  for (const bucket of groups.values()) {
    bucket.sort((a, b) => a.code.localeCompare(b.code));
  }
  return groups;
}

function CapBullet({ grade }: { grade: CapabilityGrade }) {
  const cls =
    grade === "view"
      ? styles.gradeView
      : grade === "manage"
        ? styles.gradeManage
        : styles.gradeApproval;
  return (
    <span
      className={`${styles.capBullet} ${cls}`}
      aria-label={GRADE_LABEL[grade]}
      title={GRADE_LABEL[grade]}
    />
  );
}

export default function RolesCapsTab() {
  const rolesQuery = useRolesQuery();
  const capsQuery = useCapabilitiesQuery();
  const linksQuery = useRoleCapabilitiesQuery();
  const kpiQuery = useKpiQuery();

  const [selection, setSelection] = useState<Selection>({ kind: "none" });

  const roles = rolesQuery.data?.items ?? [];
  const capabilities: ReadonlyArray<CapabilityDescriptor> =
    capsQuery.data?.items ?? [];
  const links: ReadonlyArray<RoleCapabilityLink> = linksQuery.data?.items ?? [];

  const capByCode = useMemo(() => {
    const m = new Map<string, CapabilityDescriptor>();
    for (const c of capabilities) m.set(c.code, c);
    return m;
  }, [capabilities]);

  const linksByRole = useMemo(() => {
    const m = new Map<string, string[]>();
    for (const l of links) {
      const bucket = m.get(l.role) ?? [];
      bucket.push(l.capability);
      m.set(l.role, bucket);
    }
    return m;
  }, [links]);

  const linksByCapability = useMemo(() => {
    const m = new Map<string, string[]>();
    for (const l of links) {
      const bucket = m.get(l.capability) ?? [];
      bucket.push(l.role);
      m.set(l.capability, bucket);
    }
    return m;
  }, [links]);

  const userCountByRole = useMemo(() => {
    const m = new Map<string, number>();
    for (const item of kpiQuery.data?.roleDistribution ?? []) {
      m.set(item.role, item.count);
    }
    return m;
  }, [kpiQuery.data]);

  const selectedCellForMatrix =
    selection.kind === "cell"
      ? { role: selection.role, domain: selection.domain }
      : null;

  const handleRoleSelect = (role: string) => {
    setSelection({ kind: "role", role });
  };

  const handleCellSelect = (cell: SelectedCell) => {
    setSelection((prev) =>
      prev.kind === "cell" && prev.role === cell.role && prev.domain === cell.domain
        ? { kind: "none" }
        : { kind: "cell", role: cell.role, domain: cell.domain },
    );
  };

  const handleCapability = (code: string) => {
    setSelection({ kind: "capability", code });
  };

  const isLoading =
    rolesQuery.isLoading || capsQuery.isLoading || linksQuery.isLoading;
  const error = rolesQuery.error ?? capsQuery.error ?? linksQuery.error;
  const handleRetry = () => {
    rolesQuery.refetch();
    capsQuery.refetch();
    linksQuery.refetch();
  };

  return (
    <section
      className={styles.tab}
      data-testid="admin-roles-caps-tab"
      aria-labelledby="admin-roles-caps-heading"
    >
      <div className={styles.headerBlock}>
        <h2 id="admin-roles-caps-heading" className={styles.heading}>
          Funções e capacidades
        </h2>
        <p className={styles.lede}>
          Matriz canônica de 8 funções por domínio de capacidade. Clique em uma função, célula
          ou capacidade para inspecionar os detalhes.
        </p>
        <span className={styles.readonlyTag} role="status">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <rect x="3" y="11" width="18" height="11" rx="2" />
            <path d="M7 11V7a5 5 0 0110 0v4" />
          </svg>
          Somente leitura — edição da matriz será habilitada quando o registro de capacidades expor mutação.
        </span>
      </div>

      <div className={styles.layout}>
        <RoleCapabilityMatrix
          roles={roles}
          capabilities={capabilities}
          links={links}
          isLoading={isLoading}
          error={error}
          onRetry={handleRetry}
          onRoleSelect={handleRoleSelect}
          onCellSelect={handleCellSelect}
          selectedCell={selectedCellForMatrix}
        />

        <aside
          className={styles.detail}
          aria-live="polite"
          aria-label="Detalhes da seleção"
        >
          {selection.kind === "none" ? (
            <>
              <div className={styles.detailHeader}>
                <span className={styles.detailKicker}>Detalhes</span>
                <h3 className={styles.detailTitle}>Selecione uma função ou célula</h3>
              </div>
              <p className={styles.detailEmpty}>
                Use a matriz à esquerda. Clique no nome de uma função para ver todas as
                capacidades concedidas, ou em uma célula para filtrar pelo domínio.
              </p>
              <p className={styles.deferNote}>
                <strong>Em breve:</strong> rotas REST cobertas por cada capacidade serão exibidas
                aqui assim que o endpoint <code>/iam/route-map</code> for habilitado.
              </p>
            </>
          ) : null}

          {selection.kind === "role" || selection.kind === "cell" ? (
            <RoleDetailPanel
              roleCode={selection.role}
              filterDomain={selection.kind === "cell" ? selection.domain : null}
              codes={linksByRole.get(selection.role) ?? []}
              capByCode={capByCode}
              userCount={userCountByRole.get(selection.role)}
              onClear={() => setSelection({ kind: "none" })}
              onCapability={handleCapability}
            />
          ) : null}

          {selection.kind === "capability" ? (
            <CapabilityDetailPanel
              code={selection.code}
              capByCode={capByCode}
              rolesHolding={linksByCapability.get(selection.code) ?? []}
              onClear={() => setSelection({ kind: "none" })}
            />
          ) : null}
        </aside>
      </div>
    </section>
  );
}

interface RoleDetailPanelProps {
  roleCode: string;
  filterDomain: CapabilityDomain | null;
  codes: ReadonlyArray<string>;
  capByCode: Map<string, CapabilityDescriptor>;
  userCount: number | undefined;
  onClear: () => void;
  onCapability: (code: string) => void;
}

function RoleDetailPanel({
  roleCode,
  filterDomain,
  codes,
  capByCode,
  userCount,
  onClear,
  onCapability,
}: RoleDetailPanelProps) {
  const filteredCodes = filterDomain
    ? codes.filter((c) => getCapabilityDomain(c) === filterDomain)
    : codes;
  const grouped = groupByDomain(filteredCodes, capByCode);

  const domainLabel = filterDomain
    ? CAPABILITY_DOMAINS.find((d) => d.key === filterDomain)?.label
    : null;

  return (
    <>
      <div className={styles.detailHeader}>
        <span className={styles.detailKicker}>
          {filterDomain ? "Função × Domínio" : "Função"}
        </span>
        <h3 className={styles.detailTitle}>
          {getRoleLabel(roleCode)}
          {domainLabel ? ` · ${domainLabel}` : ""}
        </h3>
        <span className={styles.detailMeta}>{roleCode}</span>
        <button type="button" className={styles.clearBtn} onClick={onClear}>
          Limpar seleção
        </button>
      </div>

      <div className={styles.statsRow}>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Capacidades</span>
          <span className={styles.statValue}>{filteredCodes.length}</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Usuários</span>
          <span className={styles.statValue}>
            {userCount === undefined ? "—" : userCount}
          </span>
        </div>
      </div>

      <h4 className={styles.sectionTitle}>Capacidades concedidas</h4>
      {grouped.size === 0 ? (
        <p className={styles.detailEmpty}>Nenhuma capacidade neste recorte.</p>
      ) : (
        <ul className={styles.groupList}>
          {CAPABILITY_DOMAINS.filter((d) => grouped.has(d.key)).map((d) => {
            const items = grouped.get(d.key) ?? [];
            return (
              <li key={d.key} className={styles.group}>
                <div className={styles.groupHeader}>
                  <span>{d.label}</span>
                  <span>{items.length}</span>
                </div>
                <ul className={styles.capList}>
                  {items.map((c) => (
                    <li key={c.code} className={styles.capItem}>
                      <CapBullet grade={getCapabilityGrade(c.code)} />
                      <button
                        type="button"
                        className={styles.clearBtn}
                        onClick={() => onCapability(c.code)}
                        title={c.description}
                      >
                        <span className={styles.capCode}>{c.code}</span>
                      </button>
                      <span className={styles.capDesc}>· {c.description}</span>
                    </li>
                  ))}
                </ul>
              </li>
            );
          })}
        </ul>
      )}

      <p className={styles.deferNote}>
        <strong>Em breve:</strong> rotas REST gatadas por estas capacidades.
      </p>
    </>
  );
}

interface CapabilityDetailPanelProps {
  code: string;
  capByCode: Map<string, CapabilityDescriptor>;
  rolesHolding: ReadonlyArray<string>;
  onClear: () => void;
}

function CapabilityDetailPanel({
  code,
  capByCode,
  rolesHolding,
  onClear,
}: CapabilityDetailPanelProps) {
  const desc = capByCode.get(code);
  const grade = getCapabilityGrade(code);
  const domain = getCapabilityDomain(code);
  const domainLabel = domain
    ? CAPABILITY_DOMAINS.find((d) => d.key === domain)?.label
    : null;

  return (
    <>
      <div className={styles.detailHeader}>
        <span className={styles.detailKicker}>Capacidade</span>
        <h3 className={styles.detailTitle}>{code}</h3>
        <span className={styles.detailMeta}>
          {domainLabel ? `${domainLabel} · ` : ""}
          {GRADE_LABEL[grade]}
        </span>
        <button type="button" className={styles.clearBtn} onClick={onClear}>
          Limpar seleção
        </button>
      </div>

      <p className={styles.detailDescription}>
        {desc?.description ?? "Capacidade sem descrição registrada."}
      </p>

      <h4 className={styles.sectionTitle}>
        Funções que concedem ({rolesHolding.length})
      </h4>
      {rolesHolding.length === 0 ? (
        <p className={styles.detailEmpty}>Nenhuma função concede esta capacidade.</p>
      ) : (
        <div className={styles.rolePillRow}>
          {[...rolesHolding]
            .sort((a, b) => getRoleLabel(a).localeCompare(getRoleLabel(b), "pt-BR"))
            .map((role) => (
              <span key={role} className={styles.rolePill}>
                {getRoleLabel(role)}
              </span>
            ))}
        </div>
      )}

      <p className={styles.deferNote}>
        <strong>Em breve:</strong> rotas REST gatadas por esta capacidade.
      </p>

      <p className={styles.detailEmpty}>
        Grades: {GRADE_ORDER.map((g) => GRADE_LABEL[g]).join(" · ")}
      </p>
    </>
  );
}
