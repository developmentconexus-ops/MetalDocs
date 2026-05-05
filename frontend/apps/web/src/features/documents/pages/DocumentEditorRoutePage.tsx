import { useNavigate, useParams } from "react-router-dom";
import { DocumentEditorPage } from "./DocumentEditorPage";

export function Component() {
  const navigate = useNavigate();
  const { documentID } = useParams();

  if (!documentID) {
    return null;
  }

  return (
    <DocumentEditorPage
      documentID={documentID}
      onDone={() => {
        navigate("/registry-v2");
      }}
    />
  );
}
