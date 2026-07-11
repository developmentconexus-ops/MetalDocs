import {
  QUORUM_KIND_OPTIONS,
  describeQuorumKind,
  labelForQuorumKind,
  type QuorumKind,
} from './routeAdminLabels';
import styles from './RouteAdmin.module.css';

interface QuorumPillsProps {
  id: string;
  value: QuorumKind;
  onChange: (kind: QuorumKind) => void;
  disabled?: boolean;
}

export function QuorumPills({ id, value, onChange, disabled }: QuorumPillsProps) {
  return (
    <div>
      <div className={styles.pillGroup} role="radiogroup">
        {QUORUM_KIND_OPTIONS.map((kind) => {
          const optionId = `${id}-${kind}`;
          const isSelected = kind === value;
          return (
            <span key={kind} className={styles.pillOption}>
              <input
                type="radio"
                id={optionId}
                name={id}
                value={kind}
                checked={isSelected}
                onChange={() => onChange(kind)}
                disabled={disabled}
              />
              <label
                htmlFor={optionId}
                className={`${styles.pillLabel} ${isSelected ? styles.pillLabelSelected : ''}`}
              >
                {labelForQuorumKind(kind)}
              </label>
            </span>
          );
        })}
      </div>
      <small className={styles.helpText}>{describeQuorumKind(value)}</small>
    </div>
  );
}
