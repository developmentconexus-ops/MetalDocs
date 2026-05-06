import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Icon } from './Icon';
import { Avatar } from './Avatar';
import { CodeChip } from './CodeChip';
import { StatusPill } from './StatusPill';

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
