import {
  statusLabel,
  type DerivedStatus,
} from "../presenters/user-status-presenter";
import styles from "./UserStatusPill.module.css";

const STATUS_CLASS: Record<DerivedStatus, string> = {
  active: styles.statusActive,
  pending: styles.statusPending,
  suspended: styles.statusSuspended,
  locked: styles.statusLocked,
};

interface UserStatusPillProps {
  status: DerivedStatus;
  withDot?: boolean;
}

export default function UserStatusPill({ status, withDot }: UserStatusPillProps) {
  return (
    <span className={`${styles.statusPill} ${STATUS_CLASS[status]}`}>
      {withDot ? <span className={styles.statusDot} aria-hidden="true" /> : null}
      {statusLabel(status)}
    </span>
  );
}
