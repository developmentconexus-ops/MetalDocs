import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { TemplateCreateDialog } from "../v2/TemplateCreateDialog";
import { TemplatesListPage } from "../v2/TemplatesListPage";

export function Component() {
  const navigate = useNavigate();
  const [showCreate, setShowCreate] = useState(false);

  return (
    <>
      <TemplatesListPage
        onOpenTemplate={(templateId, versionNum) => navigate(`/templates-v2/${templateId}/versions/${versionNum}`)}
        onCreate={() => setShowCreate(true)}
      />
      {showCreate && (
        <TemplateCreateDialog
          onClose={() => setShowCreate(false)}
          onCreated={(templateId, versionNum) => {
            setShowCreate(false);
            navigate(`/templates-v2/${templateId}/versions/${versionNum}`);
          }}
        />
      )}
    </>
  );
}
