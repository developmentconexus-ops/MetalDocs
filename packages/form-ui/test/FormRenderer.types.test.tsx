import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import type { RJSFSchema } from '@rjsf/utils';
import { FormRenderer } from '../src/FormRenderer';

const schema = {
  title: 'Test',
  type: 'object',
  properties: { name: { type: 'string', title: 'Name' } },
  required: ['name'],
} satisfies RJSFSchema;

describe('FormRenderer', () => {
  it('renders input for string property', () => {
    render(<FormRenderer schema={schema} formData={{}} onChange={() => {}} />);
    expect(screen.getByLabelText(/Name/)).toBeTruthy();
  });
});
