import { useNavigate } from "react-router-dom";
import { TemplatesListPage } from "../TemplatesListPage";
import { useHasCapability } from "../../iam/hooks/useHasCapability";

export function Component() {
  const navigate = useNavigate();
  const canViewTokens = useHasCapability("token.view");

  return (
    <TemplatesListPage
      onOpenTemplate={(templateId, versionNum) =>
        navigate(`/templates/${templateId}/versions/${versionNum}`)
      }
      onCreate={() => navigate("/templates/new")}
      onOpenTokenDictionary={canViewTokens ? () => navigate("/templates/tokens") : undefined}
    />
  );
}
