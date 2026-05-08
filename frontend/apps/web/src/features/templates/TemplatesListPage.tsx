import { useState } from "react";
import { Icon } from "../../components/ui/Icon";
import { TabBar, type TabBarItem } from "../../components/ui/TabBar";
import { WorkspaceHeroHeader } from "../../components/ui/WorkspaceHeroHeader";
import { TemplateCard } from "./components/TemplateCard";
import { useTemplatesQuery } from "./queries/useTemplatesQuery";
import { resolveQueryError } from "../../lib/api/resolveQueryError";
import styles from "./TemplatesListPage.module.css";

export type TemplatesListPageProps = {
  onOpenTemplate: (templateId: string, versionNum: number) => void;
  onCreate: () => void;
};

type TabKey = "all" | "published" | "draft" | "archived";

function formatRelative(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime();
  const days = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  if (days === 0) return "hoje";
  if (days === 1) return "ontem";
  if (days < 30) return `${days} dias atrás`;
  const months = Math.floor(days / 30);
  if (months < 12) return `${months} ${months === 1 ? "mês" : "meses"} atrás`;
  const years = Math.floor(days / 365);
  return `${years} ${years === 1 ? "ano" : "anos"} atrás`;
}

export function TemplatesListPage(props: TemplatesListPageProps) {
  const [activeTab, setActiveTab] = useState<TabKey>("all");
  const { data, isLoading, isError, error } = useTemplatesQuery();

  const templates = (data?.templates ?? []).map((dto) => ({
    id: dto.id,
    title: dto.name,
    version: `v${dto.latest_version}`,
    status: (dto.archived_at
      ? "archived"
      : dto.published_version_id
      ? "published"
      : "draft") as "published" | "draft" | "archived",
    author: dto.created_by,
    updated: formatRelative(dto.created_at),
    latestVersion: dto.latest_version,
  }));

  const tabs: TabBarItem[] = [
    { key: "all", label: "Todos", count: templates.length },
    { key: "published", label: "Publicados", count: templates.filter((t) => t.status === "published").length },
    { key: "draft", label: "Rascunhos", count: templates.filter((t) => t.status === "draft").length },
    { key: "archived", label: "Arquivados", count: templates.filter((t) => t.status === "archived").length },
  ];

  const filtered = activeTab === "all" ? templates : templates.filter((t) => t.status === activeTab);

  return (
    <div className={styles.page}>
      <div className={styles.content}>
        <WorkspaceHeroHeader
          tone="flat"
          kicker="Templates"
          title="Layouts reutilizáveis"
          subtitle="Versionados, aprovados, publicados. Vinculados a perfis para clonagem em novos documentos."
          action={
            <button type="button" className={styles.newBtn} onClick={() => props.onCreate()}>
              <Icon name="plus" size={13} />
              Novo template
            </button>
          }
        />

        <TabBar
          tabs={tabs}
          activeKey={activeTab}
          onTabChange={(key) => setActiveTab(key as TabKey)}
          ariaLabel="Filtro de templates"
        />

        {isLoading && (
          <div className={styles.loading}>Carregando templates...</div>
        )}

        {isError && (
          <div className={styles.error} role="alert">
            {resolveQueryError(error, "Erro ao carregar templates.")}
          </div>
        )}

        {!isLoading && !isError && filtered.length === 0 && (
          <div className={styles.empty}>
            Nenhum template{activeTab !== "all" ? ` ${activeTab}` : ""} encontrado.
          </div>
        )}

        {!isLoading && !isError && filtered.length > 0 && (
          <div className={styles.cardGrid}>
            {filtered.map((tpl) => (
              <TemplateCard
                key={tpl.id}
                title={tpl.title}
                version={tpl.version}
                status={tpl.status}
                author={tpl.author}
                updated={tpl.updated}
                onClick={() => props.onOpenTemplate(tpl.id, tpl.latestVersion)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
