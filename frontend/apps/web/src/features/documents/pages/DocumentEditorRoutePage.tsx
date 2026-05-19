import { useNavigate, useParams } from "react-router-dom";
import { DocumentEditorPage } from "./DocumentEditorPage";

export function Component() {
  const navigate = useNavigate();
  const { documentId } = useParams();

  if (!documentId) {
    return null;
  }

  return (
    <DocumentEditorPage
      documentID={documentId}
      onDone={() => {
        navigate("/controlled-documents");
      }}
    />
  );
}
