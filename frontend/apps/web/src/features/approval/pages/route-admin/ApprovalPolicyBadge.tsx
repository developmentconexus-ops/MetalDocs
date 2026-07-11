import type { RoutePolicy } from './routeGovernance';
import styles from './RouteAdmin.module.css';

interface ApprovalPolicyBadgeProps {
  policy: RoutePolicy;
}

const TONE_CLASS: Record<RoutePolicy['badgeTone'], string> = {
  required: styles.policyBadgeRequired,
  optional: styles.policyBadgeOptional,
  blocked: styles.policyBadgeBlocked,
};

export function ApprovalPolicyBadge({ policy }: ApprovalPolicyBadgeProps) {
  return (
    <span className={`${styles.policyBadge} ${TONE_CLASS[policy.badgeTone]}`}>
      {policy.badgeLabel}
    </span>
  );
}
