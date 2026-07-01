import { ArtifactDetailLayout } from "../../shared/controlled-artifact/ArtifactDetailLayout";
import type { ArtifactTab } from "../../shared/controlled-artifact/types";

/**
 * Documents-owned tab set for the controlled-artifact detail shell. The shared
 * `ArtifactDetailLayout` holds no kind-specific defaults (ADR 0053 purity); each
 * consumer supplies its own tabs. Documents get "Documento" + "Distribuição".
 */
const DOCUMENT_TABS: ArtifactTab[] = [
  { key: "documento", label: "Documento", href: "." },
  { key: "distribuicao", label: "Distribuição", href: "distribution" },
];

export function DocumentDetailLayout() {
  return <ArtifactDetailLayout tabs={DOCUMENT_TABS} />;
}
