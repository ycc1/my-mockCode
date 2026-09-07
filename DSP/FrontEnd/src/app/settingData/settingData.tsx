import { Check, Database, Save } from "lucide-react";
import { useState } from "react";
import { PageFrame } from "../account/account";

const accounts = [
  { id: "admin", name: "Admin", role: "超级管理员" },
  { id: "lin-wei", name: "Lin Wei", role: "运营经理" },
  { id: "mia-chen", name: "Mia Chen", role: "报表分析员" },
];
const partners = [
  {
    id: "oceanic",
    name: "Oceanic Performance",
    numbers: ["OCE-001", "OCE-002", "OCE-003"],
  },
  { id: "internal", name: "内投增长中心", numbers: ["INT-001", "INT-002"] },
  { id: "northstar", name: "Northstar Creators", numbers: ["NOR-001"] },
  { id: "search", name: "Search Lab", numbers: ["SEA-001", "SEA-002"] },
];

type AccessState = Record<string, { partners: string[]; numbers: string[] }>;
const initialAccess: AccessState = {
  admin: {
    partners: partners.map((partner) => partner.id),
    numbers: partners.flatMap((partner) => partner.numbers),
  },
  "lin-wei": {
    partners: ["oceanic", "internal"],
    numbers: ["OCE-001", "OCE-002", "INT-001"],
  },
  "mia-chen": { partners: ["search"], numbers: ["SEA-001"] },
};

export default function SettingDataPage() {
  const [accountId, setAccountId] = useState("lin-wei");
  const [access, setAccess] = useState<AccessState>(initialAccess);
  const [saved, setSaved] = useState(false);
  const current = access[accountId] ?? { partners: [], numbers: [] };
  const selectedAccount = accounts.find((account) => account.id === accountId);
  const toggle = (type: "partners" | "numbers", id: string) =>
    setAccess((state) => ({
      ...state,
      [accountId]: {
        ...current,
        [type]: current[type].includes(id)
          ? current[type].filter((item) => item !== id)
          : [...current[type], id],
      },
    }));
  const togglePartner = (partnerId: string) => {
    toggle("partners", partnerId);
    const partner = partners.find((item) => item.id === partnerId);
    if (partner && current.partners.includes(partnerId))
      setAccess((state) => ({
        ...state,
        [accountId]: {
          ...state[accountId],
          numbers: state[accountId].numbers.filter(
            (number) => !partner.numbers.includes(number),
          ),
        },
      }));
  };
  return (
    <PageFrame
      eyebrow="SYSTEM / DATA ACCESS"
      title="数据权限管理"
      description="将帐号绑定到指定渠道商与渠道号，登入后只能查看已配置的数据。"
    >
      <section className="data-access-toolbar table-panel">
        <div>
          <h2>选择帐号</h2>
          <p>配置完成后，该帐号只能访问下方勾选的数据范围。</p>
        </div>
        <select
          value={accountId}
          onChange={(event) => setAccountId(event.target.value)}
        >
          {accounts.map((account) => (
            <option value={account.id} key={account.id}>
              {account.name} · {account.role}
            </option>
          ))}
        </select>
        <button
          className="primary-button compact"
          onClick={() => {
            setSaved(true);
            window.setTimeout(() => setSaved(false), 2200);
          }}
        >
          <Save size={15} />
          {saved ? "已保存" : "保存权限"}
        </button>
      </section>
      <div className="data-access-grid">
        <section className="table-panel setting-card">
          <div className="setting-heading">
            <div className="metric-icon">
              <Database size={18} />
            </div>
            <div>
              <h2>渠道商</h2>
              <p>{selectedAccount?.name} 可查看的渠道商</p>
            </div>
          </div>
          <div className="selection-list">
            {partners.map((partner) => (
              <label className="selection-row" key={partner.id}>
                <input
                  type="checkbox"
                  checked={current.partners.includes(partner.id)}
                  onChange={() => togglePartner(partner.id)}
                />
                <span className="check-box">
                  <Check size={13} />
                </span>
                <span>
                  <strong>{partner.name}</strong>
                  <small>{partner.numbers.length} 个渠道号</small>
                </span>
              </label>
            ))}
          </div>
        </section>
        <section className="table-panel setting-card">
          <div className="setting-heading">
            <div className="metric-icon blue">
              <Database size={18} />
            </div>
            <div>
              <h2>渠道号</h2>
              <p>只可配置已授权渠道商下的渠道号</p>
            </div>
          </div>
          <div className="selection-list">
            {partners
              .flatMap((partner) =>
                partner.numbers.map((number) => ({ number, partner })),
              )
              .map(({ number, partner }) => (
                <label
                  className={`selection-row ${current.partners.includes(partner.id) ? "" : "disabled-row"}`}
                  key={number}
                >
                  <input
                    type="checkbox"
                    disabled={!current.partners.includes(partner.id)}
                    checked={current.numbers.includes(number)}
                    onChange={() => toggle("numbers", number)}
                  />
                  <span className="check-box">
                    <Check size={13} />
                  </span>
                  <span>
                    <strong>{number}</strong>
                    <small>{partner.name}</small>
                  </span>
                </label>
              ))}
          </div>
        </section>
      </div>
    </PageFrame>
  );
}
