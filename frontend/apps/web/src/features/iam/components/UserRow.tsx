import { type KeyboardEvent, type MouseEvent } from "react";
import { Avatar } from "../../../components/ui/Avatar";
import { getRelativeTime } from "../presenters/relative-time-presenter";
import {
  deriveStatus,
  type ManagedUser,
} from "../presenters/user-status-presenter";
import { SP_DATE_TIME_FORMATTER } from "../constants";
import UserStatusPill from "./UserStatusPill";
import styles from "./UserRow.module.css";

interface UserRowProps {
  user: ManagedUser;
  selected: boolean;
  onToggleSelect: (userId: string) => void;
  onOpen: (userId: string) => void;
}

function formatLastLogin(iso?: string): string {
  if (!iso) return "Nunca";
  return getRelativeTime(iso);
}

export default function UserRow({
  user,
  selected,
  onToggleSelect,
  onOpen,
}: UserRowProps) {
  const status = deriveStatus(user);
  const mfaOn = user.mfa_enabled === true;

  const handleRowClick = (e: MouseEvent<HTMLDivElement>) => {
    const target = e.target as HTMLElement;
    if (target.closest("input,button,a,label")) return;
    onOpen(user.user_id);
  };

  const handleKey = (e: KeyboardEvent<HTMLDivElement>) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      onOpen(user.user_id);
    }
  };

  return (
    <div
      role="row"
      tabIndex={0}
      className={`${styles.row} ${selected ? styles.selected : ""}`}
      onClick={handleRowClick}
      onKeyDown={handleKey}
      data-testid={`user-row-${user.user_id}`}
      data-user-id={user.user_id}
    >
      <div role="cell" className={styles.cell}>
        <input
          type="checkbox"
          className={styles.checkbox}
          checked={selected}
          onChange={() => onToggleSelect(user.user_id)}
          aria-label={`Selecionar ${user.display_name}`}
        />
      </div>
      <div role="cell" className={styles.cell}>
        <div className={styles.identity}>
          <Avatar name={user.display_name || user.username} size="sm" />
          <div className={styles.who}>
            <span className={styles.name}>{user.display_name || user.username}</span>
            <span className={styles.username}>@{user.username}</span>
          </div>
        </div>
      </div>
      <div role="cell" className={styles.cell}>
        <span className={styles.email}>{user.email ?? "—"}</span>
      </div>
      <div role="cell" className={styles.cell}>
        <span
          className={`${styles.roleChip} ${styles[user.tenant_role] ?? ""}`}
          title={user.tenant_role}
        >
          {user.tenant_role}
        </span>
      </div>
      <div role="cell" className={styles.cell}>
        <span className={styles.areaCount}>
          {user.area_memberships.length}{" "}
          {user.area_memberships.length === 1 ? "área" : "áreas"}
        </span>
      </div>
      <div role="cell" className={styles.cell}>
        <UserStatusPill status={status} withDot />
      </div>
      <div role="cell" className={styles.cell}>
        <span
          className={styles.lastLogin}
          title={user.last_login_at ? SP_DATE_TIME_FORMATTER.format(new Date(user.last_login_at)) : "Sem registro"}
        >
          {formatLastLogin(user.last_login_at)}
        </span>
      </div>
      <div role="cell" className={styles.cell}>
        <span className={`${styles.mfa} ${mfaOn ? styles.mfaOn : styles.mfaOff}`}>
          {mfaOn ? "Sim" : "Não"}
        </span>
      </div>
      <div role="cell" className={styles.cell}>
        <div className={styles.actions}>
          <button
            type="button"
            className={styles.iconBtn}
            aria-label={`Abrir detalhes de ${user.display_name}`}
            onClick={() => onOpen(user.user_id)}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M9 6l6 6-6 6" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  );
}
