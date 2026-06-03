import { useMemo } from "react";
import type { ChangeEvent } from "react";
import type { IamRole } from "../types";
import { IAM_AREA_ROLES } from "../constants";
import styles from "./MembershipsFilterBar.module.css";

export interface MembershipsFilterValue {
  areaCode?: string;
  role?: IamRole;
}

interface MembershipsFilterBarProps {
  value: MembershipsFilterValue;
  onChange: (next: MembershipsFilterValue) => void;
  areaCodes: ReadonlyArray<string>;
}

function readChange<T extends string>(
  event: ChangeEvent<HTMLSelectElement>,
): T | undefined {
  const raw = event.target.value;
  return raw === "" ? undefined : (raw as T);
}

export default function MembershipsFilterBar({
  value,
  onChange,
  areaCodes,
}: MembershipsFilterBarProps) {
  const hasAny = useMemo(
    () => Boolean(value.areaCode || value.role),
    [value],
  );

  return (
    <div className={styles.bar} role="group" aria-label="Filtros de memberships">
      <label className={`${styles.facet} ${value.areaCode ? styles.active : ""}`}>
        <span className="visually-hidden">Área</span>
        <select
          aria-label="Filtrar por área"
          value={value.areaCode ?? ""}
          onChange={(e) =>
            onChange({ ...value, areaCode: readChange<string>(e) })
          }
        >
          <option value="">Área</option>
          {areaCodes.map((code) => (
            <option key={code} value={code}>
              {code}
            </option>
          ))}
        </select>
      </label>

      <label className={`${styles.facet} ${value.role ? styles.active : ""}`}>
        <span className="visually-hidden">Função</span>
        <select
          aria-label="Filtrar por função"
          value={value.role ?? ""}
          onChange={(e) => onChange({ ...value, role: readChange<IamRole>(e) })}
        >
          <option value="">Função</option>
          {IAM_AREA_ROLES.map(([v, label]) => (
            <option key={v} value={v}>
              {label}
            </option>
          ))}
        </select>
      </label>

      {hasAny ? (
        <button type="button" className={styles.clear} onClick={() => onChange({})}>
          Limpar filtros
        </button>
      ) : null}
    </div>
  );
}
