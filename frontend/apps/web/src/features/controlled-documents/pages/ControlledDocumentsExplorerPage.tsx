import { useNavigate } from "react-router-dom";
import { ControlledDocumentListPage } from "../ControlledDocumentListPage";

export function Component() {
  const navigate = useNavigate();

  return (
    <ControlledDocumentListPage
      basePath="/controlled-documents"
      onOpenDocumentEditor={(docId) => navigate(`/documents/${docId}/edit`)}
    />
  );
}
