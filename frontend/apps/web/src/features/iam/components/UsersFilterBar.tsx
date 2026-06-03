import { useMemo } from "react";
import type { ChangeEvent } from "react";
import type { IamRole } from "../types";
import { ROLE_OPTIONS } from "../constants";
import styles from "./UsersFilterBar.module.css";

export type StatusFacet = "active" | "pending" | "suspended" | "locked";
export type MfaFacet = "on" | "off";
export type LastLoginFacet = "24h" | "7d" | "30d" | "90d";

export interface UsersFilterValue {
  status?: StatusFacet;
  role?: IamRole;
  area?: string;
  mfa?: MfaFacet;
  lastLogin?: LastLoginFacet;
}

interface UsersFilterBarProps {
  value: UsersFilterValue;
  onChange: (next: UsersFilterValue) => void;
  areaCodes: ReadonlyArray<string>;
}

const STATUS_OPTIONS: ReadonlyArray<[StatusFacet, string]> = [
  ["active", "Ativos"],
  ["pending", "Pendentes"],
  ["suspended", "Suspensos"],
  ["locked", "Bloqueados"],
];

const MFA_OPTIONS: ReadonlyArray<[MfaFacet, string]> = [
  ["on", "MFA ativo"],
  ["off", "Sem MFA"],
];

const LAST_LOGIN_OPTIONS: ReadonlyArray<[LastLoginFacet, string]> = [
  ["24h", "Últimas 24h"],
  ["7d", "Últimos 7 dias"],
  ["30d", "Últimos 30 dias"],
  ["90d", "Últimos 90 dias"],
];

function readChange<T extends string>(
  event: ChangeEvent<HTMLSelectElement>,
): T | undefined {
  const raw = event.target.value;
  return raw === "" ? undefined : (raw as T);
}

export default function UsersFilterBar({
  value,
  onChange,
  areaCodes,
}: UsersFilterBarProps) {
  const hasAny = useMemo(
    () =>
      Boolean(
        value.status || value.role || value.area || value.mfa || value.lastLogin,
      ),
    [value],
  );

  return (
    <div className={styles.bar} role="group" aria-label="Filtros de usuários">
      <label className={`${styles.facet} ${value.status ? styles.active : ""}`}>
        <span className="visually-hidden">Status</span>
        <select
          aria-label="Filtrar por status"
          value={value.status ?? ""}
          onChange={(e) =>
            onChange({ ...value, status: readChange<StatusFacet>(e) })
          }
        >
          <option value="">Status</option>
          {STATUS_OPTIONS.map(([v, label]) => (
            <option key={v} value={v}>
              {label}
            </option>
          ))}
        </select>
      </label>

      <label className={`${styles.facet} ${value.role ? styles.active : ""}`}>
        <span className="visually-hidden">Função</span>
        <select
          aria-label="Filtrar por função"
          value={value.role ?? ""}
          onChange={(e) =>
            onChange({ ...value, role: readChange<IamRole>(e) })
          }
        >
          <option value="">Função</option>
          {ROLE_OPTIONS.map(([v, label]) => (
            <option key={v} value={v}>
              {label}
            </option>
          ))}
        </select>
      </label>

      <label className={`${styles.facet} ${value.area ? styles.active : ""}`}>
        <span className="visually-hidden">Área</span>
        <select
          aria-label="Filtrar por área"
          value={value.area ?? ""}
          onChange={(e) =>
            onChange({ ...value, area: readChange<string>(e) })
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

      <label className={`${styles.facet} ${value.mfa ? styles.active : ""}`}>
        <span className="visually-hidden">MFA</span>
        <select
          aria-label="Filtrar por MFA"
          value={value.mfa ?? ""}
          onChange={(e) =>
            onChange({ ...value, mfa: readChange<MfaFacet>(e) })
          }
        >
          <option value="">MFA</option>
          {MFA_OPTIONS.map(([v, label]) => (
            <option key={v} value={v}>
              {label}
            </option>
          ))}
        </select>
      </label>

      <label className={`${styles.facet} ${value.lastLogin ? styles.active : ""}`}>
        <span className="visually-hidden">Último acesso</span>
        <select
          aria-label="Filtrar por último acesso"
          value={value.lastLogin ?? ""}
          onChange={(e) =>
            onChange({ ...value, lastLogin: readChange<LastLoginFacet>(e) })
          }
        >
          <option value="">Último acesso</option>
          {LAST_LOGIN_OPTIONS.map(([v, label]) => (
            <option key={v} value={v}>
              {label}
            </option>
          ))}
        </select>
      </label>

      {hasAny ? (
        <button
          type="button"
          className={styles.clear}
          onClick={() => onChange({})}
        >
          Limpar filtros
        </button>
      ) : null}
    </div>
  );
}
