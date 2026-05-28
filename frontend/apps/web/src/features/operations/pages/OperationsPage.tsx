import { OperationsCenter } from "../../../components/OperationsCenter";

const fmtDate = (value?: string) => (value ? new Date(value).toLocaleDateString("pt-BR") : "");
const noop = () => {};

export function Component() {
  return (
    <OperationsCenter
      loadState="ready"
      documents={[]}
      notifications={[]}
      documentProfiles={[]}
      processAreas={[]}
      formatDate={fmtDate}
      onRefreshWorkspace={noop}
      onOpenDocument={noop}
    />
  );
}
