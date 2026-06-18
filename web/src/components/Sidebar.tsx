import type { PageKey } from '../types';

type Props = {
  activePage: PageKey;
  onPageChange: (page: PageKey) => void;
  authenticated: boolean;
};

const navItems: Array<{ key: PageKey; label: string }> = [
  { key: 'status', label: '系统状态' },
  { key: 'detection', label: 'VPS检测' },
  { key: 'ports', label: '端口控制' },
];

export function Sidebar({ activePage, authenticated, onPageChange }: Props) {
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
        <span className={authenticated ? 'authState ok' : 'authState warn'}>
          {authenticated ? '已通过安装链接授权' : '请使用安装输出的访问链接'}
        </span>
      </div>
    </aside>
  );
}
