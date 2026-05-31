import { describe, expect, it } from 'vitest';
import {
  initialTemplateWizardState,
  selectMaxReachableStep,
  slugifyTemplateName,
  templateWizardReducer,
  type TemplateWizardState,
} from './templateWizard.reducer';

function readyState(overrides: Partial<TemplateWizardState> = {}): TemplateWizardState {
  return {
    ...initialTemplateWizardState,
    scopeType: 'generic',
    name: 'Procedimento Operacional',
    description: 'Template padrão para procedimentos.',
    startingPoint: 'blank',
    ...overrides,
  };
}

describe('template wizard identity validation', () => {
  it('keeps punctuation-only names at Step 2 because no API key can be generated', () => {
    const state = readyState({ step: 4, name: '!!!' });

    expect(slugifyTemplateName(state.name)).toBe('');
    expect(selectMaxReachableStep(state)).toBe(2);
    expect(templateWizardReducer(state, { type: 'GO_TO_STEP', step: 4 }).step).toBe(2);
  });
});

describe('template wizard flow', () => {
  it('reaches confirmation at step 4 without template-use permissions', () => {
    const state = readyState({ step: 4 });

    expect(selectMaxReachableStep(state)).toBe(4);
    expect(templateWizardReducer(state, { type: 'GO_TO_STEP', step: 4 }).step).toBe(4);
  });

  it('requires a selected docx file before confirming docx import', () => {
    const docxState = readyState({
      startingPoint: 'docx',
      selectedDocxFile: null,
      selectedDocxName: null,
    });

    expect(selectMaxReachableStep(docxState)).toBe(3);

    const selected = templateWizardReducer(docxState, {
      type: 'SET_SELECTED_DOCX',
      file: new File(['docx-bytes'], 'modelo.docx', {
        type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      }),
    });

    expect(selected.selectedDocxName).toBe('modelo.docx');
    expect(selected.selectedDocxSize).toBe(10);
    expect(selected.selectedDocxFile).toBeInstanceOf(File);
    expect(selectMaxReachableStep(selected)).toBe(4);
  });
});
