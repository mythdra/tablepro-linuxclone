import { Database, HardDrive, Server, Warehouse, Box, CircleDot } from 'lucide-react';
import type { DatabaseType } from '../types';

/**
 * Color and icon configuration for each database type.
 * Uses Lucide SVG icons instead of emojis per ui-ux-pro-max guidelines.
 */
const databaseConfig: Record<DatabaseType, { icon: typeof Database; color: string; label: string }> = {
  postgres: { icon: Database, color: 'text-blue-400', label: 'PostgreSQL' },
  mysql: { icon: Database, color: 'text-orange-400', label: 'MySQL' },
  sqlite: { icon: HardDrive, color: 'text-emerald-400', label: 'SQLite' },
  duckdb: { icon: Box, color: 'text-yellow-400', label: 'DuckDB' },
  mssql: { icon: Server, color: 'text-red-400', label: 'SQL Server' },
  clickhouse: { icon: Warehouse, color: 'text-amber-400', label: 'ClickHouse' },
  mongodb: { icon: CircleDot, color: 'text-green-400', label: 'MongoDB' },
  redis: { icon: CircleDot, color: 'text-red-500', label: 'Redis' },
};

interface DatabaseIconProps {
  type: DatabaseType;
  className?: string;
  size?: 'sm' | 'md' | 'lg';
}

/**
 * SVG database type icon component.
 * Replaces emoji icons with consistent Lucide SVG icons with per-database colors.
 */
export function DatabaseIcon({ type, className = '', size = 'md' }: DatabaseIconProps) {
  const config = databaseConfig[type] || databaseConfig.postgres;
  const Icon = config.icon;
  const sizeClass = size === 'sm' ? 'w-3.5 h-3.5' : size === 'lg' ? 'w-6 h-6' : 'w-4 h-4';

  return <Icon className={`${sizeClass} ${config.color} ${className}`} />;
}

/**
 * Get the display label for a database type.
 */
export function getDatabaseLabel(type: DatabaseType): string {
  return databaseConfig[type]?.label || type;
}

export default DatabaseIcon;
