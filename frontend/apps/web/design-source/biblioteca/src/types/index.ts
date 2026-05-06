export type ProfileCode = "po" | "it" | "rg" | "fm";
export type AreaCode    = "ven" | "est" | "fin" | "rh" | "com";
export type DocStatus   = "vigente" | "draft" | "review";
export type NavKey      = "dashboard" | "acervo" | "approvals" | "audit" | "mine" | "recent" | "users" | "profiles" | "ops" | "help" | "settings";

export interface Area {
  code:  AreaCode;
  label: string;
  color: string;
  count: number;
}

export interface Profile {
  code:   ProfileCode;
  label:  string;
  family: string;
  bg:     string;
  color:  string;
  count:  number;
}

export interface Document {
  id:         number;
  code:       string;
  title:      string;
  profile:    ProfileCode;
  area:       AreaCode;
  status:     DocStatus;
  owner:      string;
  version:    string;
  nextReview: string;
  urgent:     boolean;
}

export interface Stat {
  label: string;
  value: number;
  sub:   string;
  color: string;
  pct:   number;
}

export interface StatusConfig {
  label: string;
  bg:    string;
  color: string;
}

export interface ActiveFilter {
  type?: ProfileCode;
  area?: AreaCode;
}
