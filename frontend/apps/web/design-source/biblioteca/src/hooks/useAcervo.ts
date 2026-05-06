import { useState, useMemo } from "react";
import type { ProfileCode, DocStatus, ActiveFilter } from "../types";
import { DOCUMENTS, AREAS } from "../data/mock";

export function useAcervo() {
  const [search,        setSearch]        = useState("");
  const [profileFilter, setProfileFilter] = useState<ProfileCode | null>(null);
  const [statusFilter,  setStatusFilter]  = useState<DocStatus | null>(null);
  const [activeFilter,  setActiveFilter]  = useState<ActiveFilter | null>(null);

  const filtered = useMemo(() => {
    let docs = DOCUMENTS;
    if (activeFilter?.type) docs = docs.filter(d => d.profile === activeFilter.type);
    if (activeFilter?.area) docs = docs.filter(d => d.area    === activeFilter.area);
    if (profileFilter)      docs = docs.filter(d => d.profile === profileFilter);
    if (statusFilter)       docs = docs.filter(d => d.status  === statusFilter);
    if (search) {
      const q = search.toLowerCase();
      docs = docs.filter(d =>
        d.title.toLowerCase().includes(q) ||
        d.code.toLowerCase().includes(q)
      );
    }
    return docs;
  }, [activeFilter, profileFilter, statusFilter, search]);

  const byArea = useMemo(() =>
    AREAS.map(a => ({ area: a, docs: filtered.filter(d => d.area === a.code) })),
    [filtered]
  );

  const urgentCount = useMemo(() => DOCUMENTS.filter(d => d.urgent).length, []);

  function toggleProfile(code: ProfileCode) {
    setProfileFilter(prev => prev === code ? null : code);
  }

  function toggleStatus(code: DocStatus) {
    setStatusFilter(prev => prev === code ? null : code);
  }

  function clearFilters() {
    setProfileFilter(null);
    setStatusFilter(null);
  }

  return {
    search, setSearch,
    profileFilter, toggleProfile,
    statusFilter,  toggleStatus,
    activeFilter,  setActiveFilter,
    clearFilters,
    filtered, byArea,
    urgentCount,
    total: DOCUMENTS.length,
  };
}
