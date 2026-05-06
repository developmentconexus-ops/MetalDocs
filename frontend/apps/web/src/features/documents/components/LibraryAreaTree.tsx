import { useQuery } from '@tanstack/react-query';
import { listTaxonomyAreas } from '../../taxonomy/api/catalog';
import { QK } from '../../../lib/queryKeys';
import styles from './LibraryAreaTree.module.css';

type Props = {
  selected: string | null;
  onSelect: (areaCode: string | null) => void;
  counts: Record<string, number>;
};

export default function LibraryAreaTree(props: Props): JSX.Element {
  const { data } = useQuery({
    queryKey: QK.taxonomy.areas(),
    queryFn: listTaxonomyAreas,
  });

  const areas = data?.items ?? [];

  return (
    <nav className={styles.root} aria-label="Filtro por area">
      <button
        type="button"
        className={`${styles.item} ${props.selected === null ? styles.active : ''}`.trim()}
        onClick={() => props.onSelect(null)}
      >
        <span className={styles.label}>Todas as areas</span>
      </button>

      {areas.map((area) => {
        const count = props.counts[area.code] ?? 0;
        const isActive = props.selected === area.code;

        return (
          <button
            key={area.code}
            type="button"
            className={`${styles.item} ${isActive ? styles.active : ''}`.trim()}
            onClick={() => props.onSelect(area.code)}
          >
            <span className={styles.label}>{area.name}</span>
            {count > 0 ? <span className={styles.badge}>{count}</span> : null}
          </button>
        );
      })}
    </nav>
  );
}
