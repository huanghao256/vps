import type { PageKey } from '../types';

type Props = {
  activePage: PageKey;
  token: string;
  onPageChange: (page: PageKey) => void;
  onTokenChange: (token: string) => void;
};

const navItems: Array<{ key: PageKey; label: string }> = [
  { key: 'status', label: '系统状态' },
  { key: 'detection', label: 'VPS检测' },
  { key: 'ports', label: '端口控制' },
];

export function Sidebar({ activePage, token, onPageChange, onTokenChange }: Props) {
  return (
    <aside className="sidebar">
      <div className="brand">
        <span className="brandMark">VI</span>
        <div>
          <strong>VPS Inspector</strong>
          <small>Linux Node Panel</small>
        </div>
      </div>

      <nav className="navList">
        {navItems.map((item) => (
          <button
            className={activePage === item.key ? 'active' : ''}
            key={item.key}
            onClick={() => onPageChange(item.key)}
          >
            {item.label}
          </button>
        ))}
      </nav>

      <div className="tokenPanel">
        <label htmlFor="token">访问令牌</label>
        <input
          id="token"
          type="password"
          value={token}
          onChange={(event) => onTokenChange(event.target.value)}
          placeholder="本地留空"
        />
      </div>
    </aside>
  );
}

