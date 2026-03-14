import { useForm, Controller, UseFormWatch, FieldErrors, UseFormRegister } from 'react-hook-form';
import type { Control } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import {
  Database,
  Server,
  Shield,
  Settings,
  TestTube,
  Save,
  Loader2,
  ChevronDown,
} from 'lucide-react';
import {
  connectionFormSchema,
  ConnectionFormData,
  databaseTypes,
  defaultPorts,
  getDefaultFormValues,
  sshAuthMethods,
  sslModes,
  safeModeLevels,
} from '../lib/connectionSchema';
import type { DatabaseConnection, DatabaseType } from '../types';

type TabId = 'general' | 'ssh' | 'ssl' | 'advanced';

interface ConnectionFormProps {
  initialData?: DatabaseConnection;
  onSave: (data: ConnectionFormData) => Promise<void>;
  onTest?: (data: ConnectionFormData) => Promise<boolean>;
  onCancel?: () => void;
}

const databaseIcons: Record<DatabaseType, string> = {
  postgres: '🐘',
  mysql: '🐬',
  sqlite: '📦',
  duckdb: '🦆',
  mssql: '🏢',
  clickhouse: '🏠',
  mongodb: '🍃',
  redis: '🔴',
};

const tabs: { id: TabId; label: string; icon: React.ReactNode }[] = [
  { id: 'general', label: 'General', icon: <Database className="w-4 h-4" /> },
  { id: 'ssh', label: 'SSH', icon: <Server className="w-4 h-4" /> },
  { id: 'ssl', label: 'SSL', icon: <Shield className="w-4 h-4" /> },
  { id: 'advanced', label: 'Advanced', icon: <Settings className="w-4 h-4" /> },
];

export function ConnectionForm({ initialData, onSave, onTest, onCancel }: ConnectionFormProps) {
  const [activeTab, setActiveTab] = useState<TabId>('general');
  const [isTesting, setIsTesting] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [testResult, setTestResult] = useState<'success' | 'error' | null>(null);

  const {
    register,
    control,
    watch,
    setValue,
    handleSubmit,
    formState: { errors, isDirty },
  } = useForm<ConnectionFormData>({
    resolver: zodResolver(connectionFormSchema) as any,
    defaultValues: initialData
      ? {
          ...initialData,
          ssh: initialData.ssh || getDefaultFormValues().ssh,
          ssl: initialData.ssl || getDefaultFormValues().ssl,
        }
      : getDefaultFormValues(),
  });

  const connectionType = watch('type');
  const sshEnabled = watch('ssh.enabled');
  const sslEnabled = watch('ssl.enabled');
  const isFileBased = connectionType === 'sqlite' || connectionType === 'duckdb';

  const handleTypeChange = (newType: DatabaseType) => {
    setValue('type', newType);
    setValue('port', defaultPorts[newType] || 5432);
  };

  const handleTest = async () => {
    if (!onTest) return;

    setIsTesting(true);
    setTestResult(null);

    try {
      const data = watch();
      const success = await onTest(data);
      setTestResult(success ? 'success' : 'error');
    } catch {
      setTestResult('error');
    } finally {
      setIsTesting(false);
    }
  };

  const handleSave = async (data: ConnectionFormData) => {
    setIsSaving(true);
    try {
      await onSave(data);
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <form onSubmit={handleSubmit(handleSave)} className="h-full flex flex-col">
      <div className="flex items-center justify-between px-6 py-4 border-b border-slate-700">
        <h2 className="text-xl font-semibold text-white">
          {initialData ? 'Edit Connection' : 'New Connection'}
        </h2>
        <div className="flex items-center gap-3">
          {testResult && (
            <span
              className={`text-sm px-3 py-1 rounded ${
                testResult === 'success'
                  ? 'bg-emerald-500/20 text-emerald-400'
                  : 'bg-red-500/20 text-red-400'
              }`}
            >
              {testResult === 'success' ? 'Connection successful' : 'Connection failed'}
            </span>
          )}
          {onTest && (
            <button
              type="button"
              onClick={handleTest}
              disabled={isTesting}
              className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-slate-300 bg-slate-700 rounded-lg hover:bg-slate-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isTesting ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <TestTube className="w-4 h-4" />
              )}
              Test Connection
            </button>
          )}
          <button
            type="submit"
            disabled={isSaving || !isDirty}
            className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-primary rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {isSaving ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Save className="w-4 h-4" />
            )}
            Save
          </button>
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              className="px-4 py-2 text-sm font-medium text-slate-400 hover:text-white transition-colors"
            >
              Cancel
            </button>
          )}
        </div>
      </div>

      <div className="flex border-b border-slate-700">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            type="button"
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 px-5 py-3 text-sm font-medium border-b-2 transition-colors ${
              activeTab === tab.id
                ? 'text-primary border-primary bg-slate-800/50'
                : 'text-slate-400 border-transparent hover:text-white hover:bg-slate-800/30'
            }`}
          >
            {tab.icon}
            {tab.label}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-y-auto p-6">
        {activeTab === 'general' && (
          <GeneralTab
            register={register}
            control={control}
            errors={errors}
            onTypeChange={handleTypeChange}
            isFileBased={isFileBased}
          />
        )}
        {activeTab === 'ssh' && (
          <SSHTab register={register} control={control} errors={errors} watch={watch} enabled={sshEnabled} />
        )}
        {activeTab === 'ssl' && (
          <SSLTab register={register} control={control} enabled={sslEnabled} />
        )}
        {activeTab === 'advanced' && <AdvancedTab register={register} control={control} />}
      </div>
    </form>
  );
}

interface TabProps {
  register: UseFormRegister<ConnectionFormData>;
  control: Control<ConnectionFormData>;
  errors: FieldErrors<ConnectionFormData>;
}

interface GeneralTabProps extends TabProps {
  onTypeChange: (type: DatabaseType) => void;
  isFileBased: boolean;
}

interface SSHTabProps extends TabProps {
  watch: UseFormWatch<ConnectionFormData>;
  enabled: boolean;
}

function GeneralTab({ register, control, errors, onTypeChange, isFileBased }: GeneralTabProps) {
  return (
    <div className="space-y-6">
      <div>
        <label className="block text-sm font-medium text-slate-300 mb-2">Connection Name</label>
        <input
          {...register('name')}
          placeholder="My Database"
          className={`w-full px-4 py-2.5 bg-slate-800 border rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors ${
            errors.name ? 'border-red-500' : 'border-slate-600'
          }`}
        />
        {errors.name && <p className="mt-1 text-sm text-red-400">{errors.name.message}</p>}
      </div>

      <div>
        <label className="block text-sm font-medium text-slate-300 mb-2">Database Type</label>
        <Controller
          name="type"
          control={control}
          render={({ field }) => (
            <div className="relative">
              <select
                {...field}
                onChange={(e) => onTypeChange(e.target.value as DatabaseType)}
                className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white appearance-none cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
              >
                {databaseTypes.map((type) => (
                  <option key={type} value={type}>
                    {databaseIcons[type as DatabaseType]} {type.charAt(0).toUpperCase() + type.slice(1)}
                  </option>
                ))}
              </select>
              <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 pointer-events-none" />
            </div>
          )}
        />
      </div>

      {!isFileBased && (
        <div className="grid grid-cols-3 gap-4">
          <div className="col-span-2">
            <label className="block text-sm font-medium text-slate-300 mb-2">Host</label>
            <input
              {...register('host')}
              placeholder="localhost"
              className={`w-full px-4 py-2.5 bg-slate-800 border rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors ${
                errors.host ? 'border-red-500' : 'border-slate-600'
              }`}
            />
            {errors.host && <p className="mt-1 text-sm text-red-400">{errors.host.message}</p>}
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Port</label>
            <input
              type="number"
              {...register('port', { valueAsNumber: true })}
              className={`w-full px-4 py-2.5 bg-slate-800 border rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors ${
                errors.port ? 'border-red-500' : 'border-slate-600'
              }`}
            />
            {errors.port && <p className="mt-1 text-sm text-red-400">{errors.port.message}</p>}
          </div>
        </div>
      )}

      {isFileBased && (
        <div>
          <label className="block text-sm font-medium text-slate-300 mb-2">Database File</label>
          <input
            {...register('localFilePath')}
            placeholder="/path/to/database.db"
            className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
          />
        </div>
      )}

      {!isFileBased && (
        <div>
          <label className="block text-sm font-medium text-slate-300 mb-2">Database Name</label>
          <input
            {...register('database')}
            placeholder="mydb"
            className={`w-full px-4 py-2.5 bg-slate-800 border rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors ${
              errors.database ? 'border-red-500' : 'border-slate-600'
            }`}
          />
          {errors.database && <p className="mt-1 text-sm text-red-400">{errors.database.message}</p>}
        </div>
      )}

      {!isFileBased && (
        <div>
          <label className="block text-sm font-medium text-slate-300 mb-2">Username</label>
          <input
            {...register('username')}
            placeholder="root"
            className={`w-full px-4 py-2.5 bg-slate-800 border rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors ${
              errors.username ? 'border-red-500' : 'border-slate-600'
            }`}
          />
          {errors.username && <p className="mt-1 text-sm text-red-400">{errors.username.message}</p>}
        </div>
      )}

      {!isFileBased && (
        <div>
          <label className="block text-sm font-medium text-slate-300 mb-2">Password</label>
          <input
            type="password"
            placeholder="••••••••"
            className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
          />
          <p className="mt-1 text-xs text-slate-500">Passwords are stored securely in your OS keychain</p>
        </div>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-slate-300 mb-2">Group</label>
          <input
            {...register('group')}
            placeholder="Production"
            className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-300 mb-2">Color Tag</label>
          <Controller
            name="colorTag"
            control={control}
            render={({ field }) => (
              <div className="flex gap-2">
                {['', 'red', 'orange', 'yellow', 'green', 'blue', 'purple'].map((color) => (
                  <button
                    key={color}
                    type="button"
                    onClick={() => field.onChange(color)}
                    className={`w-8 h-8 rounded-full border-2 transition-all ${
                      field.value === color ? 'border-white scale-110' : 'border-transparent'
                    }`}
                    style={{
                      backgroundColor: color
                        ? {
                            red: '#ef4444',
                            orange: '#f97316',
                            yellow: '#eab308',
                            green: '#22c55e',
                            blue: '#3b82f6',
                            purple: '#a855f7',
                          }[color]
                        : '#475569',
                    }}
                  />
                ))}
              </div>
            )}
          />
        </div>
      </div>
    </div>
  );
}

function SSHTab({ register, control, errors, watch, enabled }: SSHTabProps) {
  const sshAuthMethod = watch('ssh.authMethod');
  
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Controller
          name="ssh.enabled"
          control={control}
          render={({ field }) => (
            <button
              type="button"
              onClick={() => field.onChange(!field.value)}
              className={`relative w-12 h-6 rounded-full transition-colors ${
                field.value ? 'bg-primary' : 'bg-slate-600'
              }`}
            >
              <span
                className={`absolute top-1 w-4 h-4 rounded-full bg-white transition-transform ${
                  field.value ? 'left-7' : 'left-1'
                }`}
              />
            </button>
          )}
        />
        <span className="text-slate-300">Enable SSH Tunnel</span>
      </div>

      {enabled && (
        <>
          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2">
              <label className="block text-sm font-medium text-slate-300 mb-2">SSH Host</label>
              <input
                {...register('ssh.host')}
                placeholder="ssh.example.com"
                className={`w-full px-4 py-2.5 bg-slate-800 border rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors ${
                  errors.ssh?.host ? 'border-red-500' : 'border-slate-600'
                }`}
              />
              {errors.ssh?.host && <p className="mt-1 text-sm text-red-400">{errors.ssh.host.message}</p>}
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Port</label>
              <input
                type="number"
                {...register('ssh.port', { valueAsNumber: true })}
                className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Username</label>
            <input
              {...register('ssh.username')}
              placeholder="sshuser"
              className={`w-full px-4 py-2.5 bg-slate-800 border rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors ${
                errors.ssh?.username ? 'border-red-500' : 'border-slate-600'
              }`}
            />
            {errors.ssh?.username && <p className="mt-1 text-sm text-red-400">{errors.ssh.username.message}</p>}
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Authentication Method</label>
            <Controller
              name="ssh.authMethod"
              control={control}
              render={({ field }) => (
                <div className="flex gap-3">
                  {sshAuthMethods.map((method) => (
                    <button
                      key={method}
                      type="button"
                      onClick={() => field.onChange(method)}
                      className={`px-4 py-2 text-sm font-medium rounded-lg border transition-colors ${
                        field.value === method
                          ? 'bg-primary/20 border-primary text-primary'
                          : 'bg-slate-800 border-slate-600 text-slate-400 hover:border-slate-500'
                      }`}
                    >
                      {method.charAt(0).toUpperCase() + method.slice(1)}
                    </button>
                  ))}
                </div>
              )}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">
              {sshAuthMethod === 'key' ? 'Private Key' : 'Password'}
            </label>
            <input
              type="password"
              placeholder="••••••••"
              className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
            />
            <p className="mt-1 text-xs text-slate-500">Credentials are stored securely in your OS keychain</p>
          </div>
        </>
      )}
    </div>
  );
}

function SSLTab({ register, control, enabled }: { register: TabProps['register']; control: TabProps['control']; enabled: boolean }) {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Controller
          name="ssl.enabled"
          control={control}
          render={({ field }) => (
            <button
              type="button"
              onClick={() => field.onChange(!field.value)}
              className={`relative w-12 h-6 rounded-full transition-colors ${
                field.value ? 'bg-primary' : 'bg-slate-600'
              }`}
            >
              <span
                className={`absolute top-1 w-4 h-4 rounded-full bg-white transition-transform ${
                  field.value ? 'left-7' : 'left-1'
                }`}
              />
            </button>
          )}
        />
        <span className="text-slate-300">Enable SSL/TLS</span>
      </div>

      {enabled && (
        <>
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">SSL Mode</label>
            <Controller
              name="ssl.mode"
              control={control}
              render={({ field }) => (
                <div className="grid grid-cols-2 gap-3">
                  {sslModes.map((mode) => (
                    <button
                      key={mode}
                      type="button"
                      onClick={() => field.onChange(mode)}
                      className={`px-4 py-2.5 text-sm font-medium rounded-lg border transition-colors ${
                        field.value === mode
                          ? 'bg-primary/20 border-primary text-primary'
                          : 'bg-slate-800 border-slate-600 text-slate-400 hover:border-slate-500'
                      }`}
                    >
                      {mode
                        .split('-')
                        .map((s) => s.charAt(0).toUpperCase() + s.slice(1))
                        .join(' ')}
                    </button>
                  ))}
                </div>
              )}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">CA Certificate</label>
            <input
              {...register('ssl.caCert')}
              placeholder="/path/to/ca-cert.pem"
              className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Client Certificate</label>
            <input
              {...register('ssl.clientCert')}
              placeholder="/path/to/client-cert.pem"
              className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Client Key</label>
            <input
              type="password"
              placeholder="/path/to/client-key.pem"
              className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
            />
            <p className="mt-1 text-xs text-slate-500">Keys are stored securely in your OS keychain</p>
          </div>
        </>
      )}
    </div>
  );
}

function AdvancedTab({ register, control }: { register: TabProps['register']; control: TabProps['control'] }) {
  return (
    <div className="space-y-6">
      <div>
        <label className="block text-sm font-medium text-slate-300 mb-2">Safe Mode</label>
        <Controller
          name="safeMode"
          control={control}
          render={({ field }) => (
            <div className="space-y-3">
              {safeModeLevels.map((level) => (
                <button
                  key={level}
                  type="button"
                  onClick={() => field.onChange(level)}
                  className={`w-full px-4 py-3 text-left rounded-lg border transition-colors ${
                    field.value === level
                      ? 'bg-primary/20 border-primary text-white'
                      : 'bg-slate-800 border-slate-600 text-slate-400 hover:border-slate-500'
                  }`}
                >
                  <div className="font-medium">
                    {level === 'off' ? 'Off' : level === 'safe' ? 'Safe' : 'Very Safe'}
                  </div>
                  <div className="text-sm text-slate-500 mt-1">
                    {level === 'off' && 'No protection - all operations allowed'}
                    {level === 'safe' && 'Require WHERE clause for UPDATE/DELETE operations'}
                    {level === 'very_safe' && 'Require WHERE clause and LIMIT for UPDATE/DELETE operations'}
                  </div>
                </button>
              ))}
            </div>
          )}
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-slate-300 mb-2">Startup Command</label>
        <textarea
          {...register('startupCommand')}
          placeholder="SET search_path TO my_schema;"
          rows={3}
          className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors font-mono text-sm"
        />
        <p className="mt-1 text-xs text-slate-500">SQL commands to run after connecting</p>
      </div>

      <div>
        <label className="block text-sm font-medium text-slate-300 mb-2">Pre-Connect Script</label>
        <textarea
          {...register('preConnectScript')}
          placeholder="#!/bin/bash&#10;echo 'Connecting...'"
          rows={3}
          className="w-full px-4 py-2.5 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors font-mono text-sm"
        />
        <p className="mt-1 text-xs text-slate-500">Shell script to run before connecting</p>
      </div>
    </div>
  );
}

export default ConnectionForm;