import { useNavigate } from "react-router-dom";
import { TemplatesListPage } from "../TemplatesListPage";
import { useHasCapability } from "../../../lib/iam/useHasCapability";

export function Component() {
  const navigate = useNavigate();
  const canViewTokens = useHasCapability("token.view");

  return (
    <TemplatesListPage
      onOpenTemplate={(templateId) => navigate(`/templates/${templateId}`)}
      onCreate={() => navigate("/templates/new")}
      onOpenTokenDictionary={canViewTokens ? () => navigate("/templates/tokens") : undefined}
    />
  );
}
