import {
  describeStageKind,
  labelForStageKind,
  STAGE_KIND_OPTIONS,
  type StageKind,
} from './routeGovernance';
import styles from './RouteAdmin.module.css';

interface StageKindControlProps {
  id: string;
  value: StageKind;
  onChange: (kind: StageKind) => void;
  disabled?: boolean;
}

export function StageKindControl({ id, value, onChange, disabled }: StageKindControlProps) {
  return (
    <div>
      <div className={styles.pillGroup} role="radiogroup">
        {STAGE_KIND_OPTIONS.map((kind) => {
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
                {labelForStageKind(kind)}
              </label>
            </span>
          );
        })}
      </div>
      <small className={styles.helpText}>{describeStageKind(value)}</small>
    </div>
  );
}
