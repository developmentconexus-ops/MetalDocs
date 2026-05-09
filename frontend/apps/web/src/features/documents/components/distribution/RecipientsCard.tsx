import { useState } from 'react';
import { Avatar } from '../../../../components/ui/Avatar';
import { SearchBar } from '../../../../components/ui/SearchBar';
import { TabBar } from '../../../../components/ui/TabBar';
import { Icon } from '../../../../components/ui/Icon';
import { useDebouncedValue } from '../../../../lib/hooks/useDebouncedValue';
import { MOCK_DISTRIBUTION, RECIPIENT_TABS } from '../../lib/distributionMeta';
import styles from './RecipientsCard.module.css';

const STATUS_CLASS: Record<string, string> = {
  ack:     styles.statusAck,
  read:    styles.statusRead,
  pending: styles.statusPending,
  overdue: styles.statusOverdue,
};

const STATUS_LABEL: Record<string, string> = {
  ack:     'Reconheceu',
  read:    'Apenas leu',
  pending: 'Pendente',
  overdue: 'Em atraso',
};

export function RecipientsCard() {
  const [activeTab, setActiveTab] = useState<string>('pending');
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebouncedValue(search, 300);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  const filtered = MOCK_DISTRIBUTION.people.filter(p => {
    const statusKey = p.overdue ? 'overdue' : p.status;
    const tabMatch = activeTab === 'all' || statusKey === activeTab;
    const q = debouncedSearch.toLowerCase();
    const searchMatch = !q || p.who.toLowerCase().includes(q) || p.area.toLowerCase().includes(q) || p.role.toLowerCase().includes(q);
    return tabMatch && searchMatch;
  });

  const allSelected = filtered.length > 0 && filtered.every(p => selectedIds.has(p.who));

  return (
    <div className={styles.card}>
      {/* Filter bar */}
      <div className={styles.filterBar}>
        <TabBar
          tabs={RECIPIENT_TABS}
          activeKey={activeTab}
          onTabChange={setActiveTab}
          ariaLabel="Filtrar destinatários"
        />
        <SearchBar
          value={search}
          onChange={setSearch}
          placeholder="Filtrar por nome, área, cargo…"
          ariaLabel="Filtrar destinatários"
        />
        <button type="button" aria-disabled="true" title="Em breve" className={styles.ctaDisabled}>
          <Icon name="filter" size={11} />
          Filtros avançados
        </button>
      </div>

      {/* Column header row */}
      <div className={styles.headerRow}>
        <span className={styles.colCheck}>
          <input
            type="checkbox"
            className={styles.checkbox}
            checked={allSelected}
            onChange={() => {
              if (allSelected) setSelectedIds(new Set());
              else setSelectedIds(new Set(filtered.map(p => p.who)));
            }}
          />
        </span>
        <span className={styles.colRecipient}>Destinatário</span>
        <span className={styles.colArea}>Área</span>
        <span className={styles.colStatus}>Status</span>
        <span className={styles.colLastActivity}>Última atividade</span>
        <span className={styles.colWhen}>Quando</span>
        <span className={styles.colActions} />
      </div>

      {/* Filtered rows */}
      {filtered.map((p, i) => {
        const statusKey = p.overdue ? 'overdue' : p.status;
        return (
          <div
            key={p.who}
            className={`${styles.recipientRow} ${i < filtered.length - 1 ? styles.recipientRowBorder : ''}`}
          >
            <span className={styles.colCheck}>
              <input
                type="checkbox"
                className={styles.checkbox}
                checked={selectedIds.has(p.who)}
                onChange={() => setSelectedIds(prev => {
                  const next = new Set(prev);
                  next.has(p.who) ? next.delete(p.who) : next.add(p.who);
                  return next;
                })}
              />
            </span>
            <div className={styles.colRecipient}>
              <Avatar name={p.who} size="sm" />
              <div className={styles.recipientInfo}>
                <div className={styles.recipientName}>{p.who}</div>
                <div className={styles.recipientRole}>{p.role}</div>
              </div>
            </div>
            <span className={styles.colArea}>{p.area}</span>
            <span className={styles.colStatus}>
              <span className={`${styles.statusPill} ${STATUS_CLASS[statusKey] ?? ''}`}>
                <span className={styles.statusDot} />
                {STATUS_LABEL[statusKey] ?? 'Pendente'}
              </span>
            </span>
            <span className={`${styles.colLastActivity} ${p.last.includes('nunca') ? styles.lastActivityDanger : ''}`}>
              {p.last}
            </span>
            <span className={styles.colWhen}>{p.when}</span>
            <div className={styles.colActions}>
              <button type="button" className={styles.actionBtn} aria-disabled="true" title="Em breve">
                <Icon name="mail" size={11} />
              </button>
              <button type="button" className={styles.actionBtn} aria-disabled="true" title="Em breve">
                <Icon name="more" size={11} />
              </button>
            </div>
          </div>
        );
      })}

      {/* Pagination footer */}
      <div className={styles.paginationFooter}>
        <span className={styles.paginationCount}>
          Mostrando <strong>{filtered.length}</strong> de <strong>{MOCK_DISTRIBUTION.people.length}</strong> destinatários
        </span>
        <span className={styles.paginationNav}>
          <button type="button" className={styles.paginationBtn}>
            <Icon name="chevron-left" size={12} />
          </button>
          <span className={styles.paginationPages}>1 / 23</span>
          <button type="button" className={styles.paginationBtn}>
            <Icon name="chevron-right" size={12} />
          </button>
        </span>
      </div>
    </div>
  );
}
