import { useMemo, useRef, type KeyboardEvent } from "react";
import {
  AUDIT_CATEGORY_LABELS,
  KNOWN_AUDIT_ACTIONS,
  getAuditEventCategory,
  getAuditEventLabel,
  type AuditEventCategory,
} from "../presenters/audit-event-presenter";
import styles from "./AuditFilterBar.module.css";

export type AuditDatePreset = "24h" | "7d" | "30d" | "90d" | "custom";

export interface AuditFilterValue {
  actorId?: string;
  action?: string;
  resourceType?: string;
  resourceId?: string;
  datePreset?: AuditDatePreset;
  occurredAfter?: string;
  occurredBefore?: string;
  q?: string;
}

interface AuditFilterBarProps {
  value: AuditFilterValue;
  onChange: (next: AuditFilterValue) => void;
}

const PRESET_OPTIONS: ReadonlyArray<readonly [AuditDatePreset, string]> = [
  ["24h", "24 h"],
  ["7d", "7 dias"],
  ["30d", "30 dias"],
  ["90d", "90 dias"],
  ["custom", "Personalizado"],
];

// Stable resource type options. Mirrors values the backend currently emits
// across the canonical modules:
//   - internal/modules/iam/delivery/http/*.go
//   - internal/modules/audit/...
//   - internal/modules/auth/...
//   - internal/modules/documents/application/service.go
// To enumerate live resource strings before extending:
//   rg -n 'ResourceType:\s*"' internal/modules
// Extend if a new resource type starts appearing.
const RESOURCE_TYPES: ReadonlyArray<readonly [string, string]> = [
  ["user", "Usuário"],
  ["role", "Função"],
  ["auth_session", "Sessão"],
  ["document", "Documento"],
  ["controlled_document", "Documento controlado"],
  ["template", "Template"],
  ["workflow", "Workflow"],
];

function groupActionsByCategory(): ReadonlyArray<
  readonly [AuditEventCategory, ReadonlyArray<string>]
> {
  const grouped = new Map<AuditEventCategory, string[]>();
  for (const action of KNOWN_AUDIT_ACTIONS) {
    const cat = getAuditEventCategory(action);
    const existing = grouped.get(cat);
    if (existing) existing.push(action);
    else grouped.set(cat, [action]);
  }
  return Array.from(grouped.entries()).map(
    ([cat, actions]) => [cat, actions.sort()] as const,
  );
}

const ACTION_GROUPS = groupActionsByCategory();

// Brazil abolished DST in 2019 (Decreto 9.772/2019), so São Paulo is fixed at
// UTC-3 year-round. Used by datetime-local conversion below — avoid pulling
// in a date library for this single offset.
const SP_UTC_OFFSET_HOURS = 3;

export default function AuditFilterBar({ value, onChange }: AuditFilterBarProps) {
  const hasAny = useMemo(
    () =>
      Boolean(
        value.actorId ||
          value.action ||
          value.resourceType ||
          value.resourceId ||
          value.datePreset ||
          value.occurredAfter ||
          value.occurredBefore ||
          value.q,
      ),
    [value],
  );

  const handlePreset = (preset: AuditDatePreset) => {
    if (preset === "custom") {
      onChange({ ...value, datePreset: "custom" });
      return;
    }
    const next = preset === value.datePreset ? undefined : preset;
    onChange({
      ...value,
      datePreset: next,
      occurredAfter: undefined,
      occurredBefore: undefined,
    });
  };

  // ARIA APG radiogroup: only the active (or first) chip is tab-focusable;
  // Arrow/Home/End move focus and select. Wrap-around.
  const presetRefs = useRef<Array<HTMLButtonElement | null>>([]);
  const activeIdx = PRESET_OPTIONS.findIndex(([v]) => v === value.datePreset);
  const focusedPresetIdx = activeIdx >= 0 ? activeIdx : 0;

  const handlePresetKeyDown = (e: KeyboardEvent<HTMLDivElement>) => {
    const last = PRESET_OPTIONS.length - 1;
    let nextIdx: number | null = null;
    switch (e.key) {
      case "ArrowLeft":
      case "ArrowUp":
        nextIdx = focusedPresetIdx <= 0 ? last : focusedPresetIdx - 1;
        break;
      case "ArrowRight":
      case "ArrowDown":
        nextIdx = focusedPresetIdx >= last ? 0 : focusedPresetIdx + 1;
        break;
      case "Home":
        nextIdx = 0;
        break;
      case "End":
        nextIdx = last;
        break;
      default:
        return;
    }
    if (nextIdx === null) return;
    e.preventDefault();
    const [nextPreset] = PRESET_OPTIONS[nextIdx];
    handlePreset(nextPreset);
    presetRefs.current[nextIdx]?.focus();
  };

  return (
    <div className={styles.bar} role="group" aria-label="Filtros de auditoria">
      <div className={styles.row}>
        <span className={styles.label}>Período</span>
        <div
          className={styles.presetGroup}
          role="radiogroup"
          aria-label="Janela de tempo"
          onKeyDown={handlePresetKeyDown}
        >
          {PRESET_OPTIONS.map(([v, label], idx) => {
            const active = value.datePreset === v;
            const focusable = focusedPresetIdx === idx;
            return (
              <button
                key={v}
                type="button"
                role="radio"
                aria-checked={active}
                tabIndex={focusable ? 0 : -1}
                ref={(el) => {
                  presetRefs.current[idx] = el;
                }}
                className={`${styles.presetBtn} ${active ? styles.presetActive : ""}`}
                onClick={() => handlePreset(v)}
              >
                {label}
              </button>
            );
          })}
        </div>

        {value.datePreset === "custom" ? (
          <div className={styles.customRange}>
            <label htmlFor="audit-after">De</label>
            <input
              id="audit-after"
              type="datetime-local"
              value={toLocalInput(value.occurredAfter)}
              onChange={(e) =>
                onChange({ ...value, occurredAfter: fromLocalInput(e.target.value) })
              }
            />
            <label htmlFor="audit-before">até</label>
            <input
              id="audit-before"
              type="datetime-local"
              value={toLocalInput(value.occurredBefore)}
              onChange={(e) =>
                onChange({ ...value, occurredBefore: fromLocalInput(e.target.value) })
              }
            />
          </div>
        ) : null}
      </div>

      <div className={styles.row}>
        <label className={`${styles.facet} ${styles.select} ${value.action ? styles.active : ""}`}>
          <span className="visually-hidden">Ação</span>
          <select
            aria-label="Filtrar por ação"
            value={value.action ?? ""}
            onChange={(e) =>
              onChange({ ...value, action: e.target.value || undefined })
            }
          >
            <option value="">Ação</option>
            {ACTION_GROUPS.map(([cat, actions]) => (
              <optgroup key={cat} label={AUDIT_CATEGORY_LABELS[cat]}>
                {actions.map((a) => (
                  <option key={a} value={a}>
                    {getAuditEventLabel(a)}
                  </option>
                ))}
              </optgroup>
            ))}
          </select>
        </label>

        <label className={`${styles.facet} ${styles.select} ${value.resourceType ? styles.active : ""}`}>
          <span className="visually-hidden">Tipo de recurso</span>
          <select
            aria-label="Filtrar por tipo de recurso"
            value={value.resourceType ?? ""}
            onChange={(e) =>
              onChange({ ...value, resourceType: e.target.value || undefined })
            }
          >
            <option value="">Recurso</option>
            {RESOURCE_TYPES.map(([v, label]) => (
              <option key={v} value={v}>
                {label}
              </option>
            ))}
          </select>
        </label>

        <label className={`${styles.facet} ${value.resourceId ? styles.active : ""}`}>
          <span className="visually-hidden">ID do recurso</span>
          <input
            type="text"
            aria-label="ID do recurso"
            placeholder="ID do recurso"
            value={value.resourceId ?? ""}
            onChange={(e) =>
              onChange({ ...value, resourceId: e.target.value || undefined })
            }
          />
        </label>

        <label className={`${styles.facet} ${value.actorId ? styles.active : ""}`}>
          <span className="visually-hidden">Ator</span>
          <input
            type="text"
            aria-label="Filtrar por ator (ID ou login)"
            placeholder="Ator (ID/login)"
            value={value.actorId ?? ""}
            onChange={(e) =>
              onChange({ ...value, actorId: e.target.value || undefined })
            }
          />
        </label>

        <div className={styles.searchBox}>
          <input
            type="search"
            aria-label="Buscar em ação ou payload"
            placeholder="Buscar em ação/payload"
            value={value.q ?? ""}
            onChange={(e) =>
              onChange({ ...value, q: e.target.value || undefined })
            }
          />
        </div>

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
    </div>
  );
}

// datetime-local inputs require `YYYY-MM-DDTHH:mm`. Convert ISO ↔ local string,
// rendering as America/Sao_Paulo wall time (the operator's frame).
function toLocalInput(iso: string | undefined): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  // Compute SP-local components via Intl, then format YYYY-MM-DDTHH:mm.
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: "America/Sao_Paulo",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).formatToParts(d);
  const get = (t: string) => parts.find((p) => p.type === t)?.value ?? "00";
  return `${get("year")}-${get("month")}-${get("day")}T${get("hour")}:${get("minute")}`;
}

function fromLocalInput(local: string): string | undefined {
  if (!local) return undefined;
  // datetime-local has no timezone. Treat as SP wall time and emit RFC3339 UTC.
  const [datePart, timePart] = local.split("T");
  if (!datePart || !timePart) return undefined;
  const [y, mo, d] = datePart.split("-").map(Number);
  const [h, mi] = timePart.split(":").map(Number);
  if ([y, mo, d, h, mi].some((n) => Number.isNaN(n))) return undefined;
  const utcMs = Date.UTC(y, mo - 1, d, h + SP_UTC_OFFSET_HOURS, mi);
  return new Date(utcMs).toISOString();
}
