import { Fragment } from 'react';

import { flowPreviewEmptyLabel, labelForStageKind, type StageKind } from './routeGovernance';
import { labelForQuorumKind, type QuorumKind } from './routeAdminLabels';
import styles from './RouteAdmin.module.css';

interface FlowStage {
  uid: string;
  label: string;
  kind: StageKind;
  quorum: QuorumKind;
}

interface RouteFlowPreviewProps {
  stages: FlowStage[];
}

export function RouteFlowPreview({ stages }: RouteFlowPreviewProps) {
  if (stages.length === 0) {
    return <p className={styles.flowEmpty}>{flowPreviewEmptyLabel()}</p>;
  }

  return (
    <div className={styles.flowPreview}>
      {stages.map((stage, index) => {
        const previous = stages[index - 1];
        const startsNewRound = previous !== undefined && previous.kind !== stage.kind;
        return (
          <Fragment key={stage.uid}>
            {startsNewRound ? <hr className={styles.flowRoundBreak} /> : null}
            {index > 0 && !startsNewRound ? <span className={styles.flowArrow}>→</span> : null}
            <div className={styles.flowChip}>
              <span className={styles.flowChipLabel}>{stage.label}</span>
              <span className={styles.flowChipMeta}>{labelForStageKind(stage.kind)}</span>
              <span className={styles.flowChipMeta}>{labelForQuorumKind(stage.quorum)}</span>
            </div>
          </Fragment>
        );
      })}
    </div>
  );
}
