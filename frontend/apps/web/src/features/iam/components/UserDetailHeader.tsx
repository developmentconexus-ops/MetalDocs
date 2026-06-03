import { Avatar } from "../../../components/ui/Avatar";
import { getRelativeTime } from "../presenters/relative-time-presenter";
import {
  deriveStatus,
  type ManagedUser,
} from "../presenters/user-status-presenter";
import UserStatusPill from "./UserStatusPill";
import styles from "./UserDetailHeader.module.css";

interface UserDetailHeaderProps {
  user: ManagedUser;
}

export default function UserDetailHeader({ user }: UserDetailHeaderProps) {
  const status = deriveStatus(user);
  const mfaOn = user.mfaEnabled === true;
  const areaCount = user.areaMemberships.length;

  return (
    <header className={styles.header}>
      <Avatar name={user.displayName || user.username} size="lg" />
      <div className={styles.identity}>
        <h2 className={styles.name}>{user.displayName || user.username}</h2>
        <span className={styles.username}>@{user.username}</span>
        {user.email ? <span className={styles.email}>{user.email}</span> : null}
        <div className={styles.metaRow}>
          <UserStatusPill status={status} />
          <span className={styles.roleChip}>{user.tenantRole}</span>
          <span className={styles.metaItem}>
            <strong>{areaCount}</strong>
            {areaCount === 1 ? "área" : "áreas"}
          </span>
          <span className={styles.metaItem}>
            Último acesso:{" "}
            <strong>{user.lastLoginAt ? getRelativeTime(user.lastLoginAt) : "nunca"}</strong>
          </span>
          <span
            className={`${styles.metaItem} ${mfaOn ? styles.mfaOn : styles.mfaOff}`}
          >
            MFA: <strong>{mfaOn ? "Sim" : "Não"}</strong>
          </span>
        </div>
      </div>
    </header>
  );
}
