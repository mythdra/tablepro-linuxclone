import React from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';

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
    <div className="flex items-center justify-between px-3 py-1.5 border-t border-slate-700 text-xs text-slate-400 bg-[#1E293B] shrink-0">
      <div className="flex items-center gap-3">
        <span>
          Rows {startRow.toLocaleString()}-{endRow.toLocaleString()} of{' '}
          {countPrefix}
          {totalCount.toLocaleString()}
        </span>
      </div>

      <div className="flex items-center gap-2">
        <button
          onClick={() => onPageChange(page - 1)}
          disabled={!hasPrev}
          className="p-1 rounded hover:bg-slate-700 disabled:opacity-40 disabled:cursor-default transition-colors cursor-pointer"
          aria-label="Previous page"
        >
          <ChevronLeft className="w-4 h-4" />
        </button>

        <span className="min-w-[80px] text-center">
          Page {page} / {totalPages}
        </span>

        <button
          onClick={() => onPageChange(page + 1)}
          disabled={!hasNext}
          className="p-1 rounded hover:bg-slate-700 disabled:opacity-40 disabled:cursor-default transition-colors cursor-pointer"
          aria-label="Next page"
        >
          <ChevronRight className="w-4 h-4" />
        </button>
      </div>

      <div className="flex items-center gap-2">
        <label htmlFor="page-size-select">
          Rows per page:
        </label>
        <select
          id="page-size-select"
          value={pageSize}
          onChange={handlePageSizeChange}
          className="px-2 py-1 bg-slate-800 border border-slate-600 rounded text-xs text-slate-300 cursor-pointer focus:outline-none focus:border-slate-500"
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
