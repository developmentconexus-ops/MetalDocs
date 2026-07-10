import { useState } from "react";
import { TabBar, type TabBarItem } from "../../../components/ui/TabBar";
import { WorkspaceHeroHeader } from "../../../components/ui/WorkspaceHeroHeader";
import { useAdminAreasQuery } from "../queries/useAdminAreasQuery";
import { useAdminProfilesQuery } from "../queries/useAdminProfilesQuery";
import { useFamiliesQuery } from "../queries/useFamiliesQuery";
import { FamilyList } from "../components/FamilyList";
import { ProfileList } from "../components/ProfileList";
import { AreaList } from "../components/AreaList";
import { resolveErrorMessage } from "../../../lib/api/problem";
import styles from "./TaxonomyAdminPage.module.css";

type Tab = "families" | "profiles" | "areas";

const TABS: TabBarItem[] = [
  { key: "families", label: "Famílias" },
  { key: "profiles", label: "Perfis" },
  { key: "areas", label: "Áreas" },
];

export function TaxonomyAdminPage() {
  const [tab, setTab] = useState<Tab>("families");
  const [includeInactiveFamilies, setIncludeInactiveFamilies] = useState(false);
  const [includeArchivedProfiles, setIncludeArchivedProfiles] = useState(false);
  const [includeArchivedAreas, setIncludeArchivedAreas] = useState(false);

  const familiesQuery = useFamiliesQuery(includeInactiveFamilies);
  const profilesQuery = useAdminProfilesQuery(includeArchivedProfiles);
  const areasQuery = useAdminAreasQuery(includeArchivedAreas);

  const activeQuery = tab === "families" ? familiesQuery : tab === "profiles" ? profilesQuery : areasQuery;
  const isLoading = activeQuery.isPending;
  const errorMessage = activeQuery.isError ? resolveErrorMessage(activeQuery.error) : null;

  return (
    <div className={styles.root}>
      <WorkspaceHeroHeader
        tone="flat"
        kicker="Taxonomia"
        title="Tipos Documentais"
        subtitle="Gerencie famílias documentais, perfis e áreas de processo."
      />

      <div className={styles.tabsRow}>
        <TabBar
          tabs={TABS}
          activeKey={tab}
          onTabChange={(key) => setTab(key as Tab)}
          ariaLabel="Seções de taxonomia"
        />
      </div>

      {isLoading && <p role="status">Carregando...</p>}
      {errorMessage && <p role="alert">{errorMessage}</p>}

      {!isLoading && !errorMessage && tab === "families" && (
        <FamilyList
          families={familiesQuery.data ?? []}
          includeInactive={includeInactiveFamilies}
          onToggleInactive={setIncludeInactiveFamilies}
        />
      )}

      {!isLoading && !errorMessage && tab === "profiles" && (
        <ProfileList
          profiles={profilesQuery.data ?? []}
          includeArchived={includeArchivedProfiles}
          onToggleArchived={setIncludeArchivedProfiles}
        />
      )}

      {!isLoading && !errorMessage && tab === "areas" && (
        <AreaList
          areas={areasQuery.data ?? []}
          includeArchived={includeArchivedAreas}
          onToggleArchived={setIncludeArchivedAreas}
        />
      )}
    </div>
  );
}
