import { useEffect, useState } from 'react';
import { Sidebar } from './components/Sidebar';
import { PortControlPage } from './pages/PortControlPage';
import { SystemStatusPage } from './pages/SystemStatusPage';
import { VpsDetectionPage } from './pages/VpsDetectionPage';
import type { PageKey } from './types';

const tokenKey = 'vps-inspector-token';

export function App() {
  const [activePage, setActivePage] = useState<PageKey>('status');
  const [token] = useState(() => readInitialToken());

  useEffect(() => {
    if (token !== '') {
      localStorage.setItem(tokenKey, token);
    }
  }, [token]);

  return (
    <main className="appShell">
      <Sidebar activePage={activePage} authenticated={token !== ''} onPageChange={setActivePage} />
      <section className="workspace">
        {activePage === 'status' && <SystemStatusPage token={token} />}
        {activePage === 'detection' && <VpsDetectionPage token={token} />}
        {activePage === 'ports' && <PortControlPage token={token} />}
      </section>
    </main>
  );
}

function readInitialToken() {
  const pathToken = window.location.pathname.split('/').filter(Boolean)[0] ?? '';
  if (isURLToken(pathToken)) {
    return decodeURIComponent(pathToken);
  }
  return localStorage.getItem(tokenKey) ?? '';
}

function isURLToken(value: string) {
  return /^[A-Za-z0-9_-]{16,128}$/.test(value);
}
