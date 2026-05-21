import { useNavigate } from "react-router-dom";
import { ControlledDocumentListPage } from "../ControlledDocumentListPage";

export function Component() {
  const navigate = useNavigate();

  return (
    <ControlledDocumentListPage
      onOpenDocumentEditor={(docId) => navigate(`/documents/${docId}/edit`)}
    />
  );
}
