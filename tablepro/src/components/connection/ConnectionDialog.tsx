import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useAppStore } from '@/stores/appStore';
import { tauriApi } from '@/lib/tauri';
import type { ConnectionConfig } from '@/types';

interface ConnectionDialogProps {
  open: boolean;
  onClose: () => void;
  editMode?: boolean;
  connection?: ConnectionConfig;
}

export function ConnectionDialog({
  open,
  onClose,
  editMode = false,
  connection,
}: ConnectionDialogProps) {
  const { addConnection, updateConnection, setActiveConnection, setConnectionInfo } =
    useAppStore();

  const [formData, setFormData] = useState({
    name: '',
    host: 'localhost',
    port: '5432',
    database: '',
    username: '',
    password: '',
    type: 'postgresql' as 'postgresql' | 'mysql',
  });

  const [isConnecting, setIsConnecting] = useState(false);

  // Reset form when opening/closing or when connection changes
  useEffect(() => {
    if (open) {
      if (editMode && connection) {
        setFormData({
          name: connection.name || '',
          host: connection.host,
          port: String(connection.port),
          database: connection.database,
          username: connection.username,
          password: '', // Don't show password
          type: 'postgresql',
        });
      } else {
        // Reset to defaults for new connection
        setFormData({
          name: '',
          host: 'localhost',
          port: '5432',
          database: '',
          username: '',
          password: '',
          type: 'postgresql',
        });
      }
    }
  }, [open, editMode, connection]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsConnecting(true);

    try {
      if (editMode && connection) {
        // Update existing connection (just save config, don't connect)
        updateConnection(connection.id, {
          name: formData.name || `${formData.host}:${formData.port}`,
          host: formData.host,
          port: parseInt(formData.port, 10),
          database: formData.database,
          username: formData.username,
        });
        onClose();
        return;
      }

      // Create new connection
      const connectionId = crypto.randomUUID();

      // Connect directly
      const info = await tauriApi.connect(
        connectionId,
        formData.type,
        formData.host,
        parseInt(formData.port, 10),
        formData.database,
        formData.username,
        formData.password
      );

      const config: ConnectionConfig = {
        id: connectionId,
        name: formData.name || `${formData.type}://${formData.host}`,
        host: formData.host,
        port: parseInt(formData.port, 10),
        database: formData.database,
        username: formData.username,
        sslMode: 'disable',
      };

      addConnection(config);
      setActiveConnection(connectionId);
      setConnectionInfo(connectionId, info);

      // Reset form
      setFormData({
        name: '',
        host: 'localhost',
        port: '5432',
        database: '',
        username: '',
        password: '',
        type: 'postgresql',
      });

      onClose();
    } catch (error) {
      console.error('Connection failed:', error);
      alert(`Connection failed: ${error}`);
    } finally {
      setIsConnecting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[425px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>
              {editMode ? 'Edit Connection' : 'New Connection'}
            </DialogTitle>
            <DialogDescription>
              Connect to a PostgreSQL or MySQL database.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            {!editMode && (
              <div className="grid gap-2">
                <Label htmlFor="type">Database Type</Label>
                <Select
                  value={formData.type}
                  onValueChange={(value: 'postgresql' | 'mysql') =>
                    setFormData({
                      ...formData,
                      type: value,
                      port: value === 'postgresql' ? '5432' : '3306',
                    })
                  }
                >
                  <SelectTrigger id="type">
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="postgresql">PostgreSQL</SelectItem>
                    <SelectItem value="mysql">MySQL</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}
            <div className="grid gap-2">
              <Label htmlFor="name">Connection Name (optional)</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                placeholder="My Database"
              />
            </div>
            <div className="grid grid-cols-3 gap-2">
              <div className="grid gap-2 col-span-2">
                <Label htmlFor="host">Host</Label>
                <Input
                  id="host"
                  value={formData.host}
                  onChange={(e) =>
                    setFormData({ ...formData, host: e.target.value })
                  }
                  placeholder="localhost"
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="port">Port</Label>
                <Input
                  id="port"
                  value={formData.port}
                  onChange={(e) =>
                    setFormData({ ...formData, port: e.target.value })
                  }
                  placeholder="5432"
                />
              </div>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="database">Database</Label>
              <Input
                id="database"
                value={formData.database}
                onChange={(e) =>
                  setFormData({ ...formData, database: e.target.value })
                }
                placeholder="mydb"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                value={formData.username}
                onChange={(e) =>
                  setFormData({ ...formData, username: e.target.value })
                }
                placeholder="postgres"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="password">
                Password {editMode && '(leave empty to keep current)'}
              </Label>
              <Input
                id="password"
                type="password"
                value={formData.password}
                onChange={(e) =>
                  setFormData({ ...formData, password: e.target.value })
                }
                placeholder={editMode ? '••••••••' : '••••••••'}
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={isConnecting}>
              {isConnecting
                ? 'Connecting...'
                : editMode
                ? 'Save'
                : 'Connect'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
