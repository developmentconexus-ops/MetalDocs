import { useNavigate, useParams } from "react-router-dom";
import { TemplateAuthorPage } from "../v2/TemplateAuthorPage";

export function Component() {
  const navigate = useNavigate();
  const { templateId, versionNum } = useParams();
  const parsedVersion = Number(versionNum);

  if (!templateId || !Number.isFinite(parsedVersion)) {
    return null;
  }

  return (
    <TemplateAuthorPage
      templateId={templateId}
      versionNum={parsedVersion}
      onNavigateToVersion={(nextTemplateId, nextVersionNum) => navigate(`/templates-v2/${nextTemplateId}/versions/${nextVersionNum}`)}
      onBack={() => navigate("/templates-v2")}
    />
  );
}
