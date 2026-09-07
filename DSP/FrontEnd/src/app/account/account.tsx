import { Plus, Search, UserRound } from "lucide-react";
import { useState } from "react";
import type { ReactNode } from "react";

const accounts = [
  {
    name: "Admin",
    email: "admin@dsp.local",
    role: "超级管理员",
    status: "启用",
    lastLogin: "刚刚",
  },
  {
    name: "Lin Wei",
    email: "lin.wei@dsp.local",
    role: "运营经理",
    status: "启用",
    lastLogin: "今天 09:42",
  },
  {
    name: "Mia Chen",
    email: "mia.chen@dsp.local",
    role: "报表分析员",
    status: "停用",
    lastLogin: "昨天 16:08",
  },
];

export default function AccountPage() {
  const [search, setSearch] = useState("");
  const visible = accounts.filter((account) =>
    `${account.name} ${account.email}`
      .toLowerCase()
      .includes(search.toLowerCase()),
  );
  return (
    <PageFrame
      eyebrow="SYSTEM / ACCOUNTS"
      title="帐号管理"
      description="管理后台使用者、角色与登入状态。"
      action="新增帐号"
    >
      <div className="page-stats">
        <span>
          <strong>12</strong> 个帐号
        </span>
        <span>
          <strong>10</strong> 个启用中
        </span>
        <span>
          <strong>2</strong> 个待处理
        </span>
      </div>
      <section className="table-panel">
        <div className="panel-toolbar">
          <div>
            <h2>帐号列表</h2>
            <p>帐号权限由角色统一管理</p>
          </div>
          <div className="search-box">
            <Search size={16} />
            <input
              placeholder="搜索姓名或邮箱"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>
        </div>
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>使用者</th>
                <th>角色</th>
                <th>状态</th>
                <th>最近登入</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {visible.map((account) => (
                <tr key={account.email}>
                  <td>
                    <div className="person-cell">
                      <span className="person-avatar">
                        <UserRound size={15} />
                      </span>
                      <div>
                        <strong>{account.name}</strong>
                        <small>{account.email}</small>
                      </div>
                    </div>
                  </td>
                  <td>
                    <span className="tag">{account.role}</span>
                  </td>
                  <td>
                    <span
                      className={`status-dot ${account.status === "停用" ? "off" : ""}`}
                    />
                    {account.status}
                  </td>
                  <td className="date-cell">{account.lastLogin}</td>
                  <td>
                    <button className="more-button">•••</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </PageFrame>
  );
}

export function PageFrame({
  eyebrow,
  title,
  description,
  action,
  children,
}: {
  eyebrow: string;
  title: string;
  description: string;
  action?: string;
  children: ReactNode;
}) {
  return (
    <>
      <section className="page-intro">
        <div>
          <p className="eyebrow">{eyebrow}</p>
          <h1>{title}</h1>
          <p className="muted">{description}</p>
        </div>
        {action && (
          <button className="primary-button compact">
            <Plus size={17} />
            {action}
          </button>
        )}
      </section>
      {children}
    </>
  );
}
