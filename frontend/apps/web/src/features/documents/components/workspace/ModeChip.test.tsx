import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ModeChip, type ModeChipViewer } from './ModeChip';
import type { WorkspaceMode } from '../../../approval/lib/workspaceMode';

const EMPTY_VIEWER: ModeChipViewer = {
  is_author: false,
  eligible_for_active_stage: false,
  has_signed_active_stage: false,
  via_delegation_from: null,
};

const CASES: Array<{ mode: WorkspaceMode; label: string }> = [
  { mode: 'author-editing', label: 'Editando' },
  { mode: 'author-changes-requested', label: 'Editando' },
  { mode: 'author-waiting', label: 'Aguardando revisão' },
  { mode: 'reviewing', label: 'Revisando' },
  { mode: 'approving', label: 'Aprovando' },
  { mode: 'observing', label: 'Visualizando' },
  { mode: 'lifecycle', label: 'Visualizando' },
];

describe('ModeChip', () => {
  it.each(CASES)('renders the PT-BR label "$label" for mode $mode', ({ mode, label }) => {
    render(<ModeChip mode={mode} viewer={EMPTY_VIEWER} />);
    expect(screen.getByText(label)).toBeInTheDocument();
  });

  it('exposes a non-empty tooltip explaining the mode', () => {
    render(<ModeChip mode="reviewing" viewer={EMPTY_VIEWER} />);
    const chip = screen.getByText('Revisando');
    expect(chip.getAttribute('title')).toBeTruthy();
  });

  it('mentions the delegation source in the tooltip when acting via delegation', () => {
    render(
      <ModeChip
        mode="reviewing"
        viewer={{ ...EMPTY_VIEWER, via_delegation_from: { user_id: 'u2', display_name: 'Bruno Said' } }}
      />,
    );
    const chip = screen.getByText('Revisando');
    expect(chip.getAttribute('title')).toContain('Bruno Said');
  });

  it('gives distinct tooltip copy for reviewing vs observing (no fabricated eligibility)', () => {
    const { rerender } = render(<ModeChip mode="reviewing" viewer={EMPTY_VIEWER} />);
    const reviewingTitle = screen.getByText('Revisando').getAttribute('title');
    rerender(<ModeChip mode="observing" viewer={EMPTY_VIEWER} />);
    const observingTitle = screen.getByText('Visualizando').getAttribute('title');
    expect(reviewingTitle).not.toEqual(observingTitle);
  });
});
