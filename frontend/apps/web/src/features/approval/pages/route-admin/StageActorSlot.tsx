// Extension point: approval-remediation M4 replaces the children (role+area
// selects) with the ActorSelector builder. Keep this boundary stable.

import type { ReactNode } from 'react';

import { stageActorSlotDefaultHeading } from './routeGovernance';
import styles from './RouteAdmin.module.css';

interface StageActorSlotProps {
  children: ReactNode;
  heading?: string;
}

export function StageActorSlot({
  children,
  heading = stageActorSlotDefaultHeading(),
}: StageActorSlotProps) {
  return (
    <fieldset className={styles.actorSlot}>
      <legend className={styles.actorSlotLegend}>{heading}</legend>
      {children}
    </fieldset>
  );
}
