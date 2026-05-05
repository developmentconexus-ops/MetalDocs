import { useNavigate } from "react-router-dom";
import { RegistryListPage } from "../RegistryListPage";

export function Component() {
  const navigate = useNavigate();

  return (
    <RegistryListPage
      onOpenDocumentEditor={(docId) => navigate(`/documents-v2/${docId}`)}
    />
  );
}
