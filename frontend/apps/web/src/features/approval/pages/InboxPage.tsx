import { InboxStack } from '../components/InboxStack';
import { InboxTimeline } from '../components/InboxTimeline';
import { InboxToolbar } from '../components/InboxToolbar';
import styles from './InboxPage.module.css';

export function InboxPage() {
  const view: string = 'stack'; // placeholder — Phase 3c replaces with useState
  return (
    <div className={styles.page}>
      <InboxToolbar view={view} onViewChange={() => {}} />
      {view === 'timeline' ? <InboxTimeline /> : <InboxStack />}
    </div>
  );
}
