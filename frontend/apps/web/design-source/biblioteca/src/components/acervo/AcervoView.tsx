import { PROFILES, STATS, STATUS_CONFIG } from "../../data/mock";
import { Chip, StatCard } from "../shared";
import { GroupSection } from "./GroupSection";
import { useAcervo } from "../../hooks/useAcervo";
import type { ActiveFilter, DocStatus, ProfileCode } from "../../types";

interface AcervoViewProps {
  activeFilter: ActiveFilter | null;
}

export function AcervoView({ activeFilter }: AcervoViewProps) {
  const {
    profileFilter, toggleProfile,
    statusFilter,  toggleStatus,
    clearFilters,
    byArea, filtered,
    urgentCount, total,
  } = useAcervo();

  const breadcrumb = [
    "Acervo completo",
    activeFilter?.type ? PROFILES.find(p => p.code === activeFilter.type)?.label : null,
  ].filter(Boolean).join(" — ");

  return (
    <div style={{ flex: 1, display: "flex", flexDirection: "column", overflow: "hidden" }}>

      {/* Toolbar */}
      <div style={{ background: "#fff", borderBottom: "1px solid #E8DEDE", padding: "0 24px", flexShrink: 0 }}>
        <div style={{ height: 44, display: "flex", alignItems: "center", gap: 12 }}>
          <Breadcrumb label={breadcrumb} />
          <div style={{ flex: 1 }} />
          <ViewToggle />
          <GroupSelect />
          <SortSelect />
        </div>

        {/* Filter chips */}
        <div style={{ display: "flex", alignItems: "center", gap: 6, padding: "8px 0", flexWrap: "wrap" }}>
          <Chip
            label="Todos"
            count={total}
            active={!profileFilter && !statusFilter}
            onClick={clearFilters}
          />
          {PROFILES.map(p => (
            <Chip
              key={p.code}
              label={p.code.toUpperCase()}
              count={p.count}
              dot={p.color}
              active={profileFilter === p.code}
              onClick={() => toggleProfile(p.code as ProfileCode)}
            />
          ))}
          <div style={{ width: 1, height: 16, background: "#E8DEDE", margin: "0 2px" }} />
          {(Object.entries(STATUS_CONFIG) as [DocStatus, { label: string }][]).map(([k, v]) => (
            <Chip
              key={k}
              label={v.label}
              active={statusFilter === k}
              onClick={() => toggleStatus(k)}
            />
          ))}
          <div style={{ width: 1, height: 16, background: "#E8DEDE", margin: "0 2px" }} />
          <Chip
            label="Mais filtros"
            icon={
              <svg width={11} height={11} viewBox="0 0 11 11" fill="none" stroke="currentColor" strokeWidth={1.4} strokeLinecap="round">
                <path d="M1 3h9M2.5 5.5h6M4 8h3" />
              </svg>
            }
          />
        </div>
      </div>

      {/* Content */}
      <div style={{ flex: 1, overflowY: "auto", padding: "18px 24px" }}>

        {/* Urgent alert */}
        {urgentCount > 0 && (
          <div style={{ background: "#FDF5E0", border: "1px solid #E8C870", borderRadius: 8, padding: "9px 13px", display: "flex", alignItems: "center", gap: 9, marginBottom: 16, fontSize: 12, color: "#6B4010" }}>
            <svg width={14} height={14} viewBox="0 0 14 14" fill="none" stroke="#C89020" strokeWidth={1.4} strokeLinecap="round" strokeLinejoin="round">
              <path d="M7 1L1 12.5h12L7 1z" /><path d="M7 5.5v3" /><circle cx="7" cy="10.5" r=".6" fill="#C89020" stroke="none" />
            </svg>
            <span><strong>{urgentCount} documentos</strong> com revisão vencendo nos próximos 30 dias.</span>
            <span style={{ marginLeft: "auto", fontWeight: 500, cursor: "pointer" }}>Ver documentos →</span>
          </div>
        )}

        {/* Stats */}
        <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 10, marginBottom: 18 }}>
          {STATS.map((s, i) => <StatCard key={i} {...s} />)}
        </div>

        {/* Document groups */}
        {byArea.map(({ area, docs }) => (
          <GroupSection key={area.code} area={area} docs={docs} />
        ))}

        {/* Empty state */}
        {filtered.length === 0 && (
          <div style={{ display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", padding: "60px 0", color: "#A89090", gap: 8 }}>
            <svg width={36} height={36} viewBox="0 0 36 36" fill="none" stroke="currentColor" strokeWidth={1} opacity={0.4}>
              <path d="M8 4h13l7 7v21H8V4z" strokeLinejoin="round" /><path d="M21 4v7h7" strokeLinejoin="round" /><path d="M12 16h12M12 20h12M12 24h8" strokeLinecap="round" />
            </svg>
            <p style={{ fontSize: 13 }}>Nenhum documento encontrado</p>
          </div>
        )}

        {/* Pagination */}
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "12px 0" }}>
          <span style={{ fontSize: 11.5, color: "#A89090" }}>
            Mostrando {filtered.length} de {total} documentos
          </span>
          <div style={{ display: "flex", gap: 3 }}>
            {["‹", "1", "2", "3", "›"].map((p, i) => (
              <button key={i} style={{ width: 27, height: 27, borderRadius: 6, border: "1px solid #E8DEDE", background: p === "1" ? "#6B1F2A" : "#fff", color: p === "1" ? "#fff" : "#483030", fontSize: 11.5, cursor: "pointer", display: "flex", alignItems: "center", justifyContent: "center", fontFamily: "DM Sans, sans-serif" }}>
                {p}
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

// ─── Sub-components ───────────────────────────────────────────────────────────

function Breadcrumb({ label }: { label: string }) {
  return (
    <div style={{ display: "flex", alignItems: "center", gap: 5, fontSize: 12.5 }}>
      <span style={{ color: "#7A6060", cursor: "pointer" }}>MetalDocs</span>
      <span style={{ color: "#A89090" }}>›</span>
      <span style={{ color: "#7A6060", cursor: "pointer" }}>Documentos</span>
      <span style={{ color: "#A89090" }}>›</span>
      <span style={{ color: "#1A0E0E", fontWeight: 500 }}>{label}</span>
    </div>
  );
}

function ViewToggle() {
  const icons = [
    <svg key="table" width={14} height={14} viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth={1.4} strokeLinecap="round"><rect x="1" y="1" width="12" height="12" rx="2" /><path d="M1 5h12M5 5v8" /></svg>,
    <svg key="grid"  width={14} height={14} viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth={1.4}><rect x="1" y="1" width="5" height="5" rx="1.5" /><rect x="8" y="1" width="5" height="5" rx="1.5" /><rect x="1" y="8" width="5" height="5" rx="1.5" /><rect x="8" y="8" width="5" height="5" rx="1.5" /></svg>,
  ];
  return (
    <div style={{ display: "flex", gap: 2, background: "#FAF6F6", border: "1px solid #E8DEDE", borderRadius: 7, padding: 2 }}>
      {icons.map((icon, i) => (
        <button key={i} style={{ width: 28, height: 26, borderRadius: 5, border: "none", cursor: "pointer", display: "flex", alignItems: "center", justifyContent: "center", background: i === 0 ? "#fff" : "none", color: i === 0 ? "#6B1F2A" : "#7A6060", boxShadow: i === 0 ? "0 1px 3px rgba(0,0,0,0.07)" : "none" }}>
          {icon}
        </button>
      ))}
    </div>
  );
}

const selectStyle: React.CSSProperties = { height: 30, padding: "0 10px", border: "1px solid #E8DEDE", borderRadius: 7, background: "#fff", fontFamily: "DM Sans, sans-serif", fontSize: 12, color: "#483030", outline: "none", cursor: "pointer" };

function GroupSelect() {
  return (
    <select style={selectStyle}>
      <option>Agrupar: por área</option>
      <option>Agrupar: por tipo</option>
      <option>Sem agrupamento</option>
    </select>
  );
}

function SortSelect() {
  return (
    <select style={selectStyle}>
      <option>Código ↑</option>
      <option>Título A–Z</option>
      <option>Mais recente</option>
      <option>Revisão próxima</option>
    </select>
  );
}
