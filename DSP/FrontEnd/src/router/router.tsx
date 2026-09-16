import { useState } from "react";
import {
  Activity,
  BarChart3,
  ChevronDown,
  ChevronRight,
  Layers3,
  LayoutDashboard,
  LogOut,
  Megaphone,
  Plus,
  RadioTower,
  Search,
  Settings2,
} from "lucide-react";
import AccountPage from "../app/account/account";
import ChannelPartnersPage from "../app/channelPartners/channelPartners";
import RolesPage from "../app/roles/roles";
import SettingDataPage from "../app/settingData/settingData";
import NormalizedResponsePage from "../app/normalizedResponse/normalizedResponse";

type Group = { label: string; icon: typeof Megaphone; items: string[] };
const groups: Group[] = [
  {
    label: "广告商务配置",
    icon: Megaphone,
    items: ["广告商配置", "上报规范化回应"],
  },
  {
    label: "渠道管理",
    icon: RadioTower,
    items: ["渠道商配置", "渠道号配置", "渠道链接配置"],
  },
  { label: "广告配置", icon: Layers3, items: ["广告内容配置"] },
  {
    label: "报表",
    icon: BarChart3,
    items: ["上报日志", "结算报表", "归因报表"],
  },
  {
    label: "系统管理",
    icon: Settings2,
    items: ["帐号管理", "角色管理", "数据权限管理"],
  },
];
const partners = [
  "Oceanic Performance",
  "内投增长中心",
  "Northstar Creators",
  "Search Lab",
];

export default function Router({ onLogout }: { onLogout: () => void }) {
  const [active, setActive] = useState("渠道商配置");
  const [expanded, setExpanded] = useState("渠道管理");
  const [search, setSearch] = useState("");
  const visible = partners.filter((name) =>
    name.toLowerCase().includes(search.toLowerCase()),
  );
  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="side-head">
          <div className="brand-mark">
            <Activity size={18} />
          </div>
          <div>
            <strong>DSP / OPS</strong>
            <span>运营管理平台</span>
          </div>
        </div>
        <div className="workspace-switch">
          <span className="workspace-dot" />
          Growth workspace
        </div>
        <nav className="nav-list">
          <button
            className="nav-dashboard"
            onClick={() => setActive("运营总览")}
          >
            <LayoutDashboard size={18} />
            <span>运营总览</span>
          </button>
          {groups.map((group) => {
            const Icon = group.icon;
            const open = expanded === group.label;
            return (
              <div className="nav-group" key={group.label}>
                <button
                  className="group-title"
                  onClick={() => setExpanded(open ? "" : group.label)}
                >
                  <Icon size={17} />
                  <span>{group.label}</span>
                  {open ? (
                    <ChevronDown size={14} />
                  ) : (
                    <ChevronRight size={14} />
                  )}
                </button>
                {open &&
                  group.items.map((item) => (
                    <button
                      className={`nav-item ${active === item ? "active" : ""}`}
                      key={item}
                      onClick={() => setActive(item)}
                    >
                      {item}
                    </button>
                  ))}
              </div>
            );
          })}
        </nav>
        <div className="sidebar-bottom">
          <div className="profile-avatar">A</div>
          <div className="profile-text">
            <strong>Admin</strong>
            <span>超级管理员</span>
          </div>
          <button className="icon-button" onClick={onLogout} title="登出">
            <LogOut size={16} />
          </button>
        </div>
      </aside>
      <main className="main-area">
        <header className="topbar">
          <div className="breadcrumb">
            <span>运营中心</span>
            <ChevronRight size={14} />
            <strong>{active}</strong>
          </div>
          <div className="top-user">
            Admin <ChevronDown size={14} />
          </div>
        </header>
        <div className="content-wrap">
          {active === "渠道商配置" ? (
            <ChannelPartnersPage />
          ) : active === "帐号管理" ? (
            <AccountPage />
          ) : active === "角色管理" ? (
            <RolesPage />
          ) : active === "数据权限管理" ? (
            <SettingDataPage />
          ) : active === "上报规范化回应" ? (
            <NormalizedResponsePage />
          ) : (
            <>
              <section className="page-intro">
                <div>
                  <p className="eyebrow">CHANNEL MANAGEMENT / 01</p>
                  <h1>{active}</h1>
                  <p className="muted">
                    统一维护渠道伙伴、流量模式与回传系统设置。
                  </p>
                </div>
                <button className="primary-button compact">
                  <Plus size={17} />
                  新增渠道商
                </button>
              </section>
              <section className="metric-grid">
                <Metric label="渠道商总数" value="24" note="+3 本月新增" />
                <Metric label="启用中" value="18" note="占比 75%" />
                <Metric label="已配置回传" value="16" note="覆盖 66.7%" />
              </section>
              <section className="table-panel">
                <div className="panel-toolbar">
                  <div>
                    <h2>渠道商列表</h2>
                    <p>共 24 个渠道商配置</p>
                  </div>
                  <div className="search-box">
                    <Search size={16} />
                    <input
                      placeholder="搜索渠道商"
                      value={search}
                      onChange={(event) => setSearch(event.target.value)}
                    />
                  </div>
                </div>
                <div className="table-scroll">
                  <table>
                    <thead>
                      <tr>
                        <th>渠道商</th>
                        <th>类型</th>
                        <th>服务类别</th>
                        <th>状态</th>
                        <th>修改时间</th>
                      </tr>
                    </thead>
                    <tbody>
                      {visible.map((name, index) => (
                        <tr key={name}>
                          <td>
                            <strong>{name}</strong>
                            <small>CP-{index + 21}</small>
                          </td>
                          <td>
                            <span className="tag">
                              {index === 1 ? "内部" : "外部"}
                            </span>
                          </td>
                          <td>{index % 2 ? "大媒体采买" : "DSP"}</td>
                          <td>
                            <span
                              className={`status-dot ${index === 2 ? "off" : ""}`}
                            />
                            {index === 2 ? "停用" : "启用"}
                          </td>
                          <td className="date-cell">
                            今天 {String(9 + index).padStart(2, "0")}:42
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </section>
            </>
          )}
        </div>
      </main>
    </div>
  );
}

function Metric({
  label,
  value,
  note,
}: {
  label: string;
  value: string;
  note: string;
}) {
  return (
    <div className="metric-card">
      <span>{label}</span>
      <strong>{value}</strong>
      <small>{note}</small>
    </div>
  );
}
