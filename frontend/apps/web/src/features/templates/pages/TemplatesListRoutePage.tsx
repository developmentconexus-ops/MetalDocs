import { useNavigate } from "react-router-dom";
import { TemplatesListPage } from "../TemplatesListPage";

export function Component() {
  const navigate = useNavigate();

  return (
    <TemplatesListPage
      onOpenTemplate={(templateId, versionNum) =>
        navigate(`/templates-v2/${templateId}/versions/${versionNum}`)
      }
      onCreate={() => navigate("/templates-v2/new")}
    />
  );
}
