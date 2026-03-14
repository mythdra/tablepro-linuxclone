import React from 'react';

const PAGE_SIZE_OPTIONS = [100, 500, 1000, 5000] as const;

export interface PaginationBarProps {
  page: number;
  pageSize: number;
  totalCount: number;
  isExact: boolean;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
}

export function PaginationBar({
  page,
  pageSize,
  totalCount,
  isExact,
  onPageChange,
  onPageSizeChange,
}: PaginationBarProps): React.ReactElement {
  const totalPages = totalCount > 0 ? Math.ceil(totalCount / pageSize) : 0;
  const startRow = totalCount > 0 ? (page - 1) * pageSize + 1 : 0;
  const endRow = Math.min(page * pageSize, totalCount);
  const hasPrev = page > 1;
  const hasNext = page < totalPages;

  const countPrefix = isExact ? '' : '~';

  const handlePageSizeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const newSize = Number(e.target.value);
    onPageSizeChange(newSize);
  };

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '6px 12px',
        borderTop: '1px solid #e5e7eb',
        fontSize: '13px',
        color: '#374151',
        backgroundColor: '#f9fafb',
        flexShrink: 0,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
        <span>
          Rows {startRow.toLocaleString()}-{endRow.toLocaleString()} of{' '}
          {countPrefix}
          {totalCount.toLocaleString()}
        </span>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
        <button
          onClick={() => onPageChange(page - 1)}
          disabled={!hasPrev}
          style={{
            padding: '2px 8px',
            border: '1px solid #d1d5db',
            borderRadius: '4px',
            backgroundColor: hasPrev ? '#ffffff' : '#f3f4f6',
            color: hasPrev ? '#374151' : '#9ca3af',
            cursor: hasPrev ? 'pointer' : 'default',
            fontSize: '13px',
          }}
          aria-label="Previous page"
        >
          Previous
        </button>

        <span style={{ minWidth: '80px', textAlign: 'center' }}>
          Page {page} / {totalPages}
        </span>

        <button
          onClick={() => onPageChange(page + 1)}
          disabled={!hasNext}
          style={{
            padding: '2px 8px',
            border: '1px solid #d1d5db',
            borderRadius: '4px',
            backgroundColor: hasNext ? '#ffffff' : '#f3f4f6',
            color: hasNext ? '#374151' : '#9ca3af',
            cursor: hasNext ? 'pointer' : 'default',
            fontSize: '13px',
          }}
          aria-label="Next page"
        >
          Next
        </button>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
        <label htmlFor="page-size-select" style={{ fontSize: '13px' }}>
          Rows per page:
        </label>
        <select
          id="page-size-select"
          value={pageSize}
          onChange={handlePageSizeChange}
          style={{
            padding: '2px 6px',
            border: '1px solid #d1d5db',
            borderRadius: '4px',
            backgroundColor: '#ffffff',
            fontSize: '13px',
          }}
        >
          {PAGE_SIZE_OPTIONS.map((size) => (
            <option key={size} value={size}>
              {size.toLocaleString()}
            </option>
          ))}
        </select>
      </div>
    </div>
  );
}

export default PaginationBar;
