import { useMemo } from "react";
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
import type { IamRole } from "../types";
import styles from "./RoleCapabilityMatrix.module.css";

interface RoleDescriptor {
  code: IamRole | string;
  label?: string;
  description?: string;
  category?: string;
}

interface CapabilityDescriptor {
  code: string;
  description: string;
  category: string;
}

interface RoleCapabilityLink {
  role: IamRole | string;
  capability: string;
}

export interface SelectedCell {
  role: IamRole | string;
  domain: CapabilityDomain;
}

interface RoleCapabilityMatrixProps {
  roles: ReadonlyArray<RoleDescriptor>;
  capabilities: ReadonlyArray<CapabilityDescriptor>;
  links: ReadonlyArray<RoleCapabilityLink>;
  isLoading?: boolean;
  error?: unknown;
  onRetry?: () => void;
  onRoleSelect?: (role: string) => void;
  onCellSelect?: (cell: SelectedCell) => void;
  selectedCell?: SelectedCell | null;
}

const SCOPE_LABEL: Record<string, string> = {
  tenant: "Tenant",
  area: "Área",
};

type DomainGradeMap = Map<CapabilityGrade, number>;
type RoleMatrix = Map<CapabilityDomain, DomainGradeMap>;

function buildRoleMatrix(
  roleCode: string,
  links: ReadonlyArray<RoleCapabilityLink>,
): RoleMatrix {
  const matrix: RoleMatrix = new Map();
  for (const link of links) {
    if (link.role !== roleCode) continue;
    const domain = getCapabilityDomain(link.capability);
    if (!domain) continue;
    const grade = getCapabilityGrade(link.capability);
    let domainMap = matrix.get(domain);
    if (!domainMap) {
      domainMap = new Map();
      matrix.set(domain, domainMap);
    }
    domainMap.set(grade, (domainMap.get(grade) ?? 0) + 1);
  }
  return matrix;
}

export default function RoleCapabilityMatrix({
  roles,
  capabilities: _capabilities,
  links,
  isLoading,
  error,
  onRetry,
  onRoleSelect,
  onCellSelect,
  selectedCell,
}: RoleCapabilityMatrixProps) {
  const perRoleMatrix = useMemo(() => {
    const out = new Map<string, RoleMatrix>();
    for (const r of roles) out.set(r.code, buildRoleMatrix(r.code, links));
    return out;
  }, [roles, links]);

  if (isLoading) {
    return (
      <div className={styles.wrap}>
        <div className={styles.skeleton}>Carregando matriz de funções…</div>
      </div>
    );
  }

  return (
    <div className={styles.wrap} data-testid="role-capability-matrix">
      {error ? (
        <div className={styles.error} role="alert">
          <span>Falha ao carregar a matriz.</span>
          {onRetry ? (
            <button type="button" className={styles.retryBtn} onClick={onRetry}>
              Tentar novamente
            </button>
          ) : null}
        </div>
      ) : null}

      <div className={styles.scroll}>
        <table
          className={styles.table}
          role="grid"
          aria-label="Matriz de funções e capacidades"
        >
          <caption className={styles.caption}>
            Funções canônicas (linhas) por domínio de capacidade (colunas). Pontos coloridos
            indicam capacidades concedidas por grade: leitura, gestão e aprovação.
          </caption>
          <thead>
            <tr>
              <th scope="col" className={`${styles.colHeader} ${styles.corner}`}>
                Função
              </th>
              {CAPABILITY_DOMAINS.map((d) => (
                <th key={d.key} scope="col" className={styles.colHeader}>
                  {d.label}
                  <span className={styles.colHelper}>{d.helper}</span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {roles.map((role) => {
              const matrix: RoleMatrix = perRoleMatrix.get(role.code) ?? new Map();
              const scope = role.category ? SCOPE_LABEL[role.category] : undefined;
              return (
                <tr key={role.code} className={styles.row}>
                  <th scope="row" className={styles.roleHeader}>
                    <button
                      type="button"
                      className={styles.roleHeaderBtn}
                      onClick={() => onRoleSelect?.(role.code)}
                      aria-label={`Detalhes da função ${getRoleLabel(role.code)}`}
                    >
                      <span className={styles.roleLabel}>{getRoleLabel(role.code)}</span>
                      <span className={styles.roleCode}>{role.code}</span>
                      {scope ? <span className={styles.roleScopeTag}>{scope}</span> : null}
                    </button>
                  </th>
                  {CAPABILITY_DOMAINS.map((domain) => {
                    const domainMap: DomainGradeMap | undefined = matrix.get(domain.key);
                    const totalGranted = domainMap
                      ? Array.from(domainMap.values()).reduce((a, b) => a + b, 0)
                      : 0;
                    const isSelected =
                      selectedCell?.role === role.code &&
                      selectedCell?.domain === domain.key;
                    const a11y = totalGranted
                      ? `${getRoleLabel(role.code)} · ${domain.label}: ${GRADE_ORDER.filter(
                          (g) => (domainMap?.get(g) ?? 0) > 0,
                        )
                          .map(
                            (g) =>
                              `${GRADE_LABEL[g]} ${domainMap?.get(g) ?? 0}`,
                          )
                          .join(", ")}`
                      : `${getRoleLabel(role.code)} · ${domain.label}: sem capacidades`;
                    return (
                      <td key={domain.key} className={styles.cell}>
                        <button
                          type="button"
                          className={styles.cellBtn}
                          aria-pressed={isSelected}
                          aria-label={a11y}
                          onClick={() =>
                            onCellSelect?.({ role: role.code, domain: domain.key })
                          }
                        >
                          {GRADE_ORDER.map((grade) => {
                            const count = domainMap?.get(grade) ?? 0;
                            const gradeClass =
                              grade === "view"
                                ? styles.gradeView
                                : grade === "manage"
                                  ? styles.gradeManage
                                  : styles.gradeApproval;
                            return (
                              <span
                                key={grade}
                                className={`${styles.dot} ${
                                  count > 0
                                    ? `${styles.dotGranted} ${gradeClass}`
                                    : styles.dotEmpty
                                }`}
                                aria-hidden="true"
                              >
                                {count > 1 ? (
                                  <span className={styles.count}>{count}</span>
                                ) : null}
                              </span>
                            );
                          })}
                        </button>
                      </td>
                    );
                  })}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <div className={styles.legend} aria-label="Legenda de grades">
        <span className={styles.legendItem}>
          <span
            className={`${styles.dot} ${styles.dotGranted} ${styles.gradeView}`}
            aria-hidden="true"
          />
          Leitura
        </span>
        <span className={styles.legendItem}>
          <span
            className={`${styles.dot} ${styles.dotGranted} ${styles.gradeManage}`}
            aria-hidden="true"
          />
          Gestão
        </span>
        <span className={styles.legendItem}>
          <span
            className={`${styles.dot} ${styles.dotGranted} ${styles.gradeApproval}`}
            aria-hidden="true"
          />
          Aprovação
        </span>
        <span className={styles.legendItem}>
          <span className={`${styles.dot} ${styles.dotEmpty}`} aria-hidden="true" />
          Sem capacidade
        </span>
      </div>
    </div>
  );
}
