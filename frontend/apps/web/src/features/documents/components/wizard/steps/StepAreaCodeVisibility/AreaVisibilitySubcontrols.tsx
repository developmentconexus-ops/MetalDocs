import type { ProcessArea } from '../../../../../taxonomy/types';
import styles from './StepAreaCodeVisibility.module.css';

export type AreaVisibilitySubcontrolsProps = {
  areas: ProcessArea[];
  documentAreaCode: string;
  selectedCodes: string[];
  onSetAreaCodes: (codes: string[]) => void;
};

export function AreaVisibilitySubcontrols({
  areas,
  documentAreaCode,
  selectedCodes,
  onSetAreaCodes,
}: AreaVisibilitySubcontrolsProps): JSX.Element {
  function toggle(code: string) {
    if (code === documentAreaCode) return;
    const next = selectedCodes.includes(code)
      ? selectedCodes.filter((current) => current !== code)
      : [...selectedCodes, code];
    onSetAreaCodes(next);
  }

  return (
    <div className={`card ${styles.areaSubcontrolsCard}`} role="group" aria-label="Areas visiveis">
      <div className="kicker">Areas visiveis</div>
      <div className={styles.subcontrolsRow}>
        {areas.map((area) => {
          const checked = selectedCodes.includes(area.code);
          const locked = area.code === documentAreaCode;
          return (
            <label key={area.code} className={styles.areaChoice}>
              <input
                type="checkbox"
                checked={checked}
                disabled={locked}
                onChange={() => toggle(area.code)}
              />
              <span>
                {area.code} - {area.name}
                {locked ? ' (area do documento)' : ''}
              </span>
            </label>
          );
        })}
      </div>
    </div>
  );
}

