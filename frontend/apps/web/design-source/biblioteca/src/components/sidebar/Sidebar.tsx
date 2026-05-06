import type { NavKey, ActiveFilter } from "../../types";
import { PROFILES } from "../../data/mock";
import { NavItem } from "./NavItem";
import { AccordionSection } from "./AccordionSection";
import { SbDivider, SbLabel } from "../shared";

interface SidebarProps {
  activeNav:    NavKey;
  setActiveNav: (k: NavKey) => void;
  activeFilter: ActiveFilter | null;
  setActiveFilter: (f: ActiveFilter | null) => void;
}

export function Sidebar({ activeNav, setActiveNav, activeFilter, setActiveFilter }: SidebarProps) {
  function goAcervo() {
    setActiveNav("acervo");
    setActiveFilter(null);
  }

  return (
    <aside style={{ width: 240, flexShrink: 0, background: "#fff", borderRight: "1px solid #E8DEDE", display: "flex", flexDirection: "column", overflow: "hidden" }}>

      {/* Scroll area */}
      <div style={{ flex: 1, overflowY: "auto", padding: "10px 8px 16px" }}>

        {/* Workspace pill */}
        <div style={{ display: "flex", alignItems: "center", gap: 8, padding: "8px 8px 4px", marginBottom: 4 }}>
          <div style={{ width: 26, height: 26, borderRadius: 7, background: "#F9F3F3", display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
            <svg width={13} height={13} viewBox="0 0 13 13" fill="none" stroke="#6B1F2A" strokeWidth={1.4} strokeLinecap="round">
              <path d="M2 2h9v9H2z" strokeLinejoin="round" /><path d="M2 5h9M5 5v6" />
            </svg>
          </div>
          <div style={{ lineHeight: 1.2 }}>
            <div style={{ fontSize: 12.5, fontWeight: 500, color: "#1A0E0E" }}>Metal Nobre</div>
            <div style={{ fontSize: 10, color: "#A89090" }}>Administrador</div>
          </div>
          <svg width={12} height={12} viewBox="0 0 12 12" fill="none" stroke="#A89090" strokeWidth={1.5} strokeLinecap="round" style={{ marginLeft: "auto" }}>
            <path d="M3 5l3 3 3-3" />
          </svg>
        </div>

        <SbDivider />

        {/* Visão geral */}
        <SbLabel>Visão geral</SbLabel>

        <NavItem
          icon={<svg viewBox="0 0 15 15" fill="none" stroke="currentColor" strokeWidth={1.4}><rect x="1.5" y="1.5" width="5" height="5" rx="1.5" /><rect x="8.5" y="1.5" width="5" height="5" rx="1.5" /><rect x="1.5" y="8.5" width="5" height="5" rx="1.5" /><rect x="8.5" y="8.5" width="5" height="5" rx="1.5" /></svg>}
          label="Dashboard"
          active={activeNav === "dashboard"}
          onClick={() => setActiveNav("dashboard")}
        />
        <NavItem
          icon={<svg viewBox="0 0 15 15" fill="none" stroke="currentColor" strokeWidth={1.4}><circle cx="7.5" cy="7.5" r="5.5" /><path d="M7.5 5v3l2 1.5" /></svg>}
          label="Aprovações"
          badge={5}
          badgeVariant="red"
          active={activeNav === "approvals"}
          onClick={() => setActiveNav("approvals")}
        />
        <NavItem
          icon={<svg viewBox="0 0 15 15" fill="none" stroke="currentColor" strokeWidth={1.4}><path d="M1.5 3.5h12M1.5 7.5h8M1.5 11.5h5" /></svg>}
          label="Audit trail"
          active={activeNav === "audit"}
          onClick={() => setActiveNav("audit")}
        />

        <SbDivider />

        {/* Documentos */}
        <SbLabel onAdd={() => {}}>Documentos</SbLabel>

        <NavItem
          icon={<svg viewBox="0 0 15 15" fill="none" stroke="currentColor" strokeWidth={1.4}><path d="M2 2h11v11H2z" strokeLinejoin="round" /><path d="M5 5.5h5M5 8h5M5 10.5h3" /></svg>}
          label="Acervo completo"
          badge={78}
          active={activeNav === "acervo" && !activeFilter}
          onClick={goAcervo}
        />
        <NavItem
          icon={<svg viewBox="0 0 15 15" fill="none" stroke="currentColor" strokeWidth={1.4}><circle cx="7.5" cy="5" r="2.5" /><path d="M2.5 13c0-2.8 2.2-5 5-5s5 2.2 5 5" /></svg>}
          label="Meus documentos"
          badge={12}
          badgeVariant="amber"
          active={activeNav === "mine"}
          onClick={() => setActiveNav("mine")}
        />
        <NavItem
          icon={<svg viewBox="0 0 15 15" fill="none" stroke="currentColor" strokeWidth={1.4}><circle cx="7.5" cy="7.5" r="5.5" /><path d="M7.5 4.5v3.5l2.5 1.5" /></svg>}
          label="Recentes"
          active={activeNav === "recent"}
          onClick={() => setActiveNav("recent")}
        />

        <SbDivider />

        {/* Por tipo */}
        <SbLabel>Por tipo</SbLabel>

        {PROFILES.map(p => (
          <AccordionSection
            key={p.code}
            profile={p}
            activeFilter={activeFilter}
            onFilter={f => { setActiveNav("acervo"); setActiveFilter(f); }}
          />
        ))}

        <SbDivider />

        {/* Admin */}
        <SbLabel>Admin</SbLabel>

        <NavItem
          icon={<svg viewBox="0 0 15 15" fill="none" stroke="currentColor" strokeWidth={1.4}><circle cx="7.5" cy="5" r="2.5" /><path d="M2.5 13c0-2.8 2.2-5 5-5s5 2.2 5 5" /></svg>}
          label="Usuários e acessos"
          active={activeNav === "users"}
          onClick={() => setActiveNav("users")}
        />
        <NavItem
          icon={<svg viewBox="0 0 15 15" fill="none" stroke="currentColor" strokeWidth={1.4}><circle cx="7.5" cy="7.5" r="2" /><path d="M7.5 1.5v1.5M7.5 12v1.5M1.5 7.5H3M12 7.5h1.5" /></svg>}
          label="Profiles e schemas"
          active={activeNav === "profiles"}
          onClick={() => setActiveNav("profiles")}
        />
        <NavItem
          icon={<svg viewBox="0 0 15 15" fill="none" stroke="currentColor" strokeWidth={1.4}><circle cx="5" cy="7.5" r="1.5" /><circle cx="10" cy="4" r="1.5" /><circle cx="10" cy="11" r="1.5" /><path d="M6.5 7.5l2-2.5M6.5 7.5l2 2.5" /></svg>}
          label="Operations center"
          dot
          active={activeNav === "ops"}
          onClick={() => setActiveNav("ops")}
        />
      </div>

      {/* Footer */}
      <div style={{ padding: 8, borderTop: "1px solid #E8DEDE", display: "flex", flexDirection: "column", gap: 1 }}>
        <NavItem
          icon={<svg viewBox="0 0 15 15" fill="none" stroke="currentColor" strokeWidth={1.4}><circle cx="7.5" cy="7.5" r="5.5" /><path d="M7.5 5.5a2 2 0 0 1 2 2c0 1.1-.9 2-2 2M7.5 11v.5" /></svg>}
          label="Ajuda e suporte"
          active={activeNav === "help"}
          onClick={() => setActiveNav("help")}
        />
        <NavItem
          icon={<svg viewBox="0 0 15 15" fill="none" stroke="currentColor" strokeWidth={1.4}><rect x="2" y="2" width="11" height="11" rx="2" /><path d="M5 7.5h5M8 5.5l2 2-2 2" /></svg>}
          label="Configurações"
          active={activeNav === "settings"}
          onClick={() => setActiveNav("settings")}
        />
      </div>
    </aside>
  );
}
