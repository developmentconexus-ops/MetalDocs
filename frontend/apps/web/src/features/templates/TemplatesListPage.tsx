import { useState } from "react";
import { Icon } from "../../components/ui/Icon";
import { TabBar, type TabBarItem } from "../../components/ui/TabBar";
import { WorkspaceHeroHeader } from "../../components/ui/WorkspaceHeroHeader";
import { TemplateCard } from "./components/TemplateCard";
import styles from "./TemplatesListPage.module.css";

export type TemplatesListPageProps = {
  onOpenTemplate: (templateId: string, versionNum: number) => void;
  onCreate: () => void;
};

type TabKey = "all" | "published" | "draft" | "archived";

type MockTemplate = {
  id: string;
  title: string;
  version: string;
  status: "published" | "draft" | "archived";
  author: string;
  updated: string;
};

const MOCK_TEMPLATES: MockTemplate[] = [
  { id: "1", title: "POP — Procedimento Operacional", version: "v3.2", status: "published", author: "Juliana Prado", updated: "12 dias atrás" },
  { id: "2", title: "DC — Descrição de Cargo", version: "v2.0", status: "published", author: "Marina Silveira", updated: "2 meses atrás" },
  { id: "3", title: "POL — Política", version: "v1.4", status: "published", author: "André Lima", updated: "5 dias atrás" },
  { id: "4", title: "IT — Instrução de Trabalho", version: "v4.1", status: "published", author: "Bruno Mendes", updated: "1 mês atrás" },
  { id: "5", title: "CHK — Checklist Operacional", version: "v0.4", status: "draft", author: "Camila Tavares", updated: "ontem" },
  { id: "6", title: "FORM — Formulário Antigo", version: "v1.0", status: "archived", author: "Rafael Costa", updated: "1 ano atrás" },
];

export function TemplatesListPage(_props: TemplatesListPageProps) {
  const [activeTab, setActiveTab] = useState<TabKey>("all");

  const tabs: TabBarItem[] = [
    { key: "all", label: "Todos", count: MOCK_TEMPLATES.length },
    { key: "published", label: "Publicados", count: MOCK_TEMPLATES.filter((t) => t.status === "published").length },
    { key: "draft", label: "Rascunhos", count: MOCK_TEMPLATES.filter((t) => t.status === "draft").length },
    { key: "archived", label: "Arquivados", count: MOCK_TEMPLATES.filter((t) => t.status === "archived").length },
  ];

  return (
    <div className={styles.page}>
      <div className={styles.content}>
        <WorkspaceHeroHeader
          kicker="Templates"
          title="Layouts reutilizáveis"
          subtitle="Versionados, aprovados, publicados. Vinculados a perfis para clonagem em novos documentos."
          action={
            <button type="button" className={styles.newBtn}>
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

        <div className={styles.cardGrid}>
          {MOCK_TEMPLATES.map((tpl) => (
            <TemplateCard
              key={tpl.id}
              title={tpl.title}
              version={tpl.version}
              status={tpl.status}
              author={tpl.author}
              updated={tpl.updated}
            />
          ))}
        </div>
      </div>
    </div>
  );
}
