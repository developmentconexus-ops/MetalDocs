import { NotificationsPanel } from "../../../components/NotificationsPanel";

const fmtDate = (value?: string) => (value ? new Date(value).toLocaleDateString("pt-BR") : "");
const noop = () => {};

export function Component() {
  return (
    <NotificationsPanel
      loadState="ready"
      notifications={[]}
      formatDate={fmtDate}
      onRefreshWorkspace={noop}
      onMarkRead={noop}
    />
  );
}
