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
  const mfaOn = user.mfaEnabled === true;

  const handleRowClick = (e: MouseEvent<HTMLDivElement>) => {
    const target = e.target as HTMLElement;
    if (target.closest("input,button,a,label")) return;
    onOpen(user.userId);
  };

  const handleKey = (e: KeyboardEvent<HTMLDivElement>) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      onOpen(user.userId);
    }
  };

  return (
    <div
      role="row"
      tabIndex={0}
      className={`${styles.row} ${selected ? styles.selected : ""}`}
      onClick={handleRowClick}
      onKeyDown={handleKey}
      data-testid={`user-row-${user.userId}`}
      data-user-id={user.userId}
    >
      <div role="cell" className={styles.cell}>
        <input
          type="checkbox"
          className={styles.checkbox}
          checked={selected}
          onChange={() => onToggleSelect(user.userId)}
          aria-label={`Selecionar ${user.displayName}`}
        />
      </div>
      <div role="cell" className={styles.cell}>
        <div className={styles.identity}>
          <Avatar name={user.displayName || user.username} size="sm" />
          <div className={styles.who}>
            <span className={styles.name}>{user.displayName || user.username}</span>
            <span className={styles.username}>@{user.username}</span>
          </div>
        </div>
      </div>
      <div role="cell" className={styles.cell}>
        <span className={styles.email}>{user.email ?? "—"}</span>
      </div>
      <div role="cell" className={styles.cell}>
        <span
          className={`${styles.roleChip} ${styles[user.tenantRole] ?? ""}`}
          title={user.tenantRole}
        >
          {user.tenantRole}
        </span>
      </div>
      <div role="cell" className={styles.cell}>
        <span className={styles.areaCount}>
          {user.areaMemberships.length}{" "}
          {user.areaMemberships.length === 1 ? "área" : "áreas"}
        </span>
      </div>
      <div role="cell" className={styles.cell}>
        <UserStatusPill status={status} withDot />
      </div>
      <div role="cell" className={styles.cell}>
        <span
          className={styles.lastLogin}
          title={user.lastLoginAt ? SP_DATE_TIME_FORMATTER.format(new Date(user.lastLoginAt)) : "Sem registro"}
        >
          {formatLastLogin(user.lastLoginAt)}
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
            aria-label={`Abrir detalhes de ${user.displayName}`}
            onClick={() => onOpen(user.userId)}
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
