import { useCallback, useEffect, useState } from "react";
import type { DocumentFamily, DocumentProfile, ProcessArea } from "./types";
import { fetchFamilies, fetchProfiles, fetchAreas } from "./api";
import { FamilyList } from "./FamilyList";
import { ProfileList } from "./ProfileList";
import { AreaList } from "./AreaList";

type Tab = "families" | "profiles" | "areas";

export function TaxonomyAdminPage() {
  const [tab, setTab] = useState<Tab>("families");
  const [families, setFamilies] = useState<DocumentFamily[]>([]);
  const [profiles, setProfiles] = useState<DocumentProfile[]>([]);
  const [areas, setAreas] = useState<ProcessArea[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [includeInactiveFamilies, setIncludeInactiveFamilies] = useState(false);
  const [includeArchivedProfiles, setIncludeArchivedProfiles] = useState(false);
  const [includeArchivedAreas, setIncludeArchivedAreas] = useState(false);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [familiesData, profilesData, areasData] = await Promise.all([
        fetchFamilies(includeInactiveFamilies),
        fetchProfiles(includeArchivedProfiles),
        fetchAreas(includeArchivedAreas),
      ]);
      setFamilies(familiesData);
      setProfiles(profilesData);
      setAreas(areasData);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Falha ao carregar dados.");
    } finally {
      setLoading(false);
    }
  }, [includeInactiveFamilies, includeArchivedProfiles, includeArchivedAreas]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const tabStyle = (t: Tab) => ({
    padding: "8px 18px",
    fontSize: 13,
    fontWeight: tab === t ? 600 : 400,
    background: "none" as const,
    border: "none" as const,
    borderBottom: tab === t ? "2px solid #333" : "2px solid transparent",
    marginBottom: -2,
    cursor: "pointer" as const,
  });

  return (
    <div style={{ padding: 24, maxWidth: 1100 }}>
      <div style={{ marginBottom: 4 }}>
        <p style={{ margin: 0, fontSize: 11, color: "#888", textTransform: "uppercase", letterSpacing: "0.08em" }}>
          Taxonomia
        </p>
        <h1 style={{ margin: "4px 0 8px", fontSize: 20, fontWeight: 600 }}>Tipos Documentais</h1>
        <p style={{ margin: "0 0 20px", fontSize: 13, color: "#555" }}>
          Gerencie famílias documentais, perfis e áreas de processo.
        </p>
      </div>

      <div style={{ display: "flex", gap: 0, marginBottom: 20, borderBottom: "2px solid #e0e0e0" }}>
        <button type="button" onClick={() => setTab("families")} style={tabStyle("families")}>Famílias</button>
        <button type="button" onClick={() => setTab("profiles")} style={tabStyle("profiles")}>Perfis</button>
        <button type="button" onClick={() => setTab("areas")} style={tabStyle("areas")}>Áreas</button>
      </div>

      {loading && <p style={{ color: "#888", fontSize: 13 }}>Carregando...</p>}
      {error && <p style={{ color: "#c00", fontSize: 13 }}>{error}</p>}

      {!loading && !error && tab === "families" && (
        <FamilyList
          families={families}
          includeInactive={includeInactiveFamilies}
          onToggleInactive={setIncludeInactiveFamilies}
          onRefresh={() => void loadData()}
        />
      )}

      {!loading && !error && tab === "profiles" && (
        <ProfileList
          profiles={profiles}
          includeArchived={includeArchivedProfiles}
          onToggleArchived={setIncludeArchivedProfiles}
          onRefresh={() => void loadData()}
        />
      )}

      {!loading && !error && tab === "areas" && (
        <AreaList
          areas={areas}
          includeArchived={includeArchivedAreas}
          onToggleArchived={setIncludeArchivedAreas}
          onRefresh={() => void loadData()}
        />
      )}
    </div>
  );
}
