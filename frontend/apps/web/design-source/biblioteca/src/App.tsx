import { useState } from "react";
import type { NavKey, ActiveFilter } from "./types";
import { Topbar } from "./components/shared/Topbar";
import { Sidebar } from "./components/sidebar/Sidebar";
import { AcervoView } from "./components/acervo/AcervoView";

export default function App() {
  const [activeNav,    setActiveNav]    = useState<NavKey>("acervo");
  const [activeFilter, setActiveFilter] = useState<ActiveFilter | null>(null);
  const [search,       setSearch]       = useState("");

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100vh", overflow: "hidden", fontFamily: "DM Sans, system-ui, sans-serif", fontSize: 13, color: "#1A0E0E", background: "#F0EBEB" }}>
      <Topbar search={search} setSearch={setSearch} />
      <div style={{ display: "flex", flex: 1, overflow: "hidden" }}>
        <Sidebar
          activeNav={activeNav}
          setActiveNav={setActiveNav}
          activeFilter={activeFilter}
          setActiveFilter={setActiveFilter}
        />
        {activeNav === "acervo" && <AcervoView activeFilter={activeFilter} />}
        {activeNav !== "acervo" && (
          <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: "#A89090", fontSize: 13 }}>
            Tela em construção — selecione "Acervo" na sidebar
          </div>
        )}
      </div>
    </div>
  );
}
