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
  const mfaOn = user.mfa_enabled === true;
  const areaCount = user.area_memberships.length;

  return (
    <header className={styles.header}>
      <Avatar name={user.display_name || user.username} size="lg" />
      <div className={styles.identity}>
        <h2 className={styles.name}>{user.display_name || user.username}</h2>
        <span className={styles.username}>@{user.username}</span>
        {user.email ? <span className={styles.email}>{user.email}</span> : null}
        <div className={styles.metaRow}>
          <UserStatusPill status={status} />
          <span className={styles.roleChip}>{user.tenant_role}</span>
          <span className={styles.metaItem}>
            <strong>{areaCount}</strong>
            {areaCount === 1 ? "área" : "áreas"}
          </span>
          <span className={styles.metaItem}>
            Último acesso:{" "}
            <strong>{user.last_login_at ? getRelativeTime(user.last_login_at) : "nunca"}</strong>
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
