import React from 'react';

import type { ICellRendererParams } from '@ag-grid-community/core';

export function NullCellRenderer(props: ICellRendererParams): React.ReactElement {
  const { value } = props;

  if (value === null || value === undefined) {
    return (
      <span
        style={{
          color: '#9ca3af',
          fontStyle: 'italic',
          userSelect: 'none',
        }}
      >
        NULL
      </span>
    );
  }

  if (value === '') {
    return <span />;
  }

  return <span>{String(value)}</span>;
}

export default NullCellRenderer;
