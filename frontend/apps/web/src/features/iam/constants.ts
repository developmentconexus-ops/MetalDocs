import type { IamRole } from "./types";

export const ROLE_OPTIONS: ReadonlyArray<readonly [IamRole, string]> = [
  ["viewer", "Visualizador"],
  ["author", "Autor"],
  ["editor", "Editor"],
  ["approver", "Aprovador"],
  ["signer", "Assinante"],
  ["area_admin", "Area admin"],
  ["qms_admin", "QMS admin"],
  ["system_admin", "System admin"],
];

export const SP_DATE_TIME_FORMATTER = new Intl.DateTimeFormat("pt-BR", {
  timeZone: "America/Sao_Paulo",
  dateStyle: "short",
  timeStyle: "short",
});

export const SP_DATE_FORMATTER = new Intl.DateTimeFormat("pt-BR", {
  timeZone: "America/Sao_Paulo",
  dateStyle: "short",
});
