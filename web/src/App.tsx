import { useEffect, useState } from 'react';
import { Sidebar } from './components/Sidebar';
import { PortControlPage } from './pages/PortControlPage';
import { SystemStatusPage } from './pages/SystemStatusPage';
import { VpsDetectionPage } from './pages/VpsDetectionPage';
import type { PageKey } from './types';

const tokenKey = 'vps-inspector-token';

export function App() {
  const [activePage, setActivePage] = useState<PageKey>('status');
  const [token, setToken] = useState(() => localStorage.getItem(tokenKey) ?? '');

  useEffect(() => {
    localStorage.setItem(tokenKey, token);
  }, [token]);

  return (
    <main className="appShell">
      <Sidebar activePage={activePage} token={token} onPageChange={setActivePage} onTokenChange={setToken} />
      <section className="workspace">
        {activePage === 'status' && <SystemStatusPage token={token} />}
        {activePage === 'detection' && <VpsDetectionPage token={token} />}
        {activePage === 'ports' && <PortControlPage token={token} />}
      </section>
    </main>
  );
}

