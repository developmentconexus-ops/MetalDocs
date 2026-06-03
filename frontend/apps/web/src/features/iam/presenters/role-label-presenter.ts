import type { IamRole } from "../types";

const ROLE_LABELS: Readonly<Record<IamRole, string>> = {
  system_admin: "Administrador do sistema",
  approver: "Aprovador",
  author: "Autor",
  editor: "Editor",
  viewer: "Visualizador",
  signer: "Assinante",
  area_admin: "Administrador de área",
  qms_admin: "Administrador do SGQ",
};

export function getRoleLabel(role: string): string {
  return (ROLE_LABELS as Record<string, string>)[role] ?? role;
}
