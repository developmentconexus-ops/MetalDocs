import { describe, it, expect } from 'vitest';
import { QK } from './queryKeys';

describe('QK', () => {
  it('documents.list returns stable key', () => {
    expect(QK.documents.list()).toEqual(['documents', 'list']);
  });

  it('documents.detail includes id', () => {
    expect(QK.documents.detail('abc')).toEqual(['documents', 'detail', 'abc']);
  });

  it('inbox includes params', () => {
    expect(QK.inbox({ page: 2, areaFilter: 'RH' }))
      .toEqual(['approval', 'inbox', { page: 2, areaFilter: 'RH' }]);
  });

  it('inbox with no params uses empty object', () => {
    expect(QK.inbox()).toEqual(['approval', 'inbox', {}]);
  });

  it('approval.instance includes documentId', () => {
    expect(QK.approval.instance('doc-1')).toEqual(['approval', 'instance', 'doc-1']);
  });
});
