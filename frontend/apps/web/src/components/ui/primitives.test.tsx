import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Icon } from './Icon';
import { Avatar } from './Avatar';
import { CodeChip } from './CodeChip';
import { StatusPill } from './StatusPill';
import { InlineAlert } from './InlineAlert';

describe('Icon', () => {
  it('renders an svg', () => {
    const { container } = render(<Icon name="home" size={16} />);
    expect(container.querySelector('svg')).toBeTruthy();
  });

  it('applies size as width/height', () => {
    const { container } = render(<Icon name="search" size={20} />);
    const svg = container.querySelector('svg')!;
    expect(svg.getAttribute('width')).toBe('20');
    expect(svg.getAttribute('height')).toBe('20');
  });
});

describe('Avatar', () => {
  it('renders two-letter initials from full name', () => {
    render(<Avatar name="Marina Silveira" />);
    expect(screen.getByText('MS')).toBeTruthy();
  });

  it('renders initials for single name', () => {
    render(<Avatar name="Admin" />);
    expect(screen.getByText('AD')).toBeTruthy();
  });
});

describe('CodeChip', () => {
  it('renders children', () => {
    render(<CodeChip>POP-RH-001</CodeChip>);
    expect(screen.getByText('POP-RH-001')).toBeTruthy();
  });
});

describe('StatusPill', () => {
  it('renders draft pill with correct class', () => {
    const { container } = render(<StatusPill status="draft" />);
    expect(container.firstChild).toHaveClass('pill-draft');
  });

  it('renders frozen pill with correct class', () => {
    const { container } = render(<StatusPill status="frozen" />);
    expect(container.firstChild).toHaveClass('pill-frozen');
  });
});

describe('InlineAlert', () => {
  it('defaults to role="alert" + aria-live="assertive" for tone=error', () => {
    render(<InlineAlert message="Falha ao carregar." />);
    const el = screen.getByRole('alert');
    expect(el.getAttribute('aria-live')).toBe('assertive');
    expect(el.getAttribute('aria-atomic')).toBe('true');
    expect(el.textContent).toContain('Falha ao carregar.');
  });

  it('uses role="status" + aria-live="polite" for tone=info', () => {
    render(<InlineAlert tone="info" message="Aguardando confirmação." />);
    const el = screen.getByRole('status');
    expect(el.getAttribute('aria-live')).toBe('polite');
  });

  it('uses role="status" + aria-live="polite" for tone=warning', () => {
    render(<InlineAlert tone="warning" message="Atenção." />);
    const el = screen.getByRole('status');
    expect(el.getAttribute('aria-live')).toBe('polite');
  });

  it('renders an optional action', () => {
    render(
      <InlineAlert
        message="Falha ao carregar perfis."
        action={<button type="button">Tentar novamente</button>}
      />
    );
    expect(screen.getByRole('button', { name: 'Tentar novamente' })).toBeTruthy();
  });

  it('omits action wrapper when no action is passed', () => {
    const { container } = render(<InlineAlert message="Erro simples" />);
    expect(container.querySelectorAll('div').length).toBe(1);
  });

  it('applies the card utility class plus an optional extra className', () => {
    const { container } = render(<InlineAlert message="x" className="extra" />);
    const el = container.firstElementChild!;
    expect(el.className).toContain('card');
    expect(el.className).toContain('extra');
  });
});
