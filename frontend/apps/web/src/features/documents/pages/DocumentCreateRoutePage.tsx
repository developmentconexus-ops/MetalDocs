import { useNavigate } from "react-router-dom";
import { DocumentCreatePage } from "../v2/DocumentCreatePage";

export function Component() {
  const navigate = useNavigate();

  return (
    <DocumentCreatePage
      onCreated={(documentID) => {
        navigate(`/documents-v2/${documentID}`);
      }}
    />
  );
}
