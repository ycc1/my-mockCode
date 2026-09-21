import { useEffect, useState } from "react";
import { PageFrame } from "../account/account";
import "../channelPartners/channelPartners.css";

type AdsPartner = {
  id: string;
  name: string;
  ads_merchant_id: string;
  api_key: string;
  security_type: string;
  create_time: string;
  update_time: string;
  create_by: string;
  update_by: string;
};
type Form = Pick<AdsPartner, "name" | "ads_merchant_id" | "api_key" | "security_type">;

const securityTypes = ["MD5", "SHA312", "AES", "DES"] as const;
const keyLengths: Record<string, number> = { MD5: 32, SHA312: 78, AES: 32, DES: 8 };
const blank = (): Form => ({ name: "", ads_merchant_id: "", api_key: "", security_type: "" });
const randomString = (length: number) => {
  const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  const values = new Uint32Array(length);
  crypto.getRandomValues(values);
  return Array.from(values, (value) => alphabet[value % alphabet.length]).join("");
};
const formatTime = (value: string) => new Date(value).toLocaleString("zh-CN", { hour12: false });

async function api(path = "", options?: RequestInit) {
  const response = await fetch(`/api/v1/ads-partners${path}`, {
    credentials: "include", ...options,
    headers: { "Content-Type": "application/json", ...options?.headers },
  });
  const payload = await response.json();
  if (!response.ok) throw new Error(payload.message || "操作失败");
  return payload.data;
}

export default function AdsPartnersPage() {
  const [rows, setRows] = useState<AdsPartner[]>([]);
  const [query, setQuery] = useState("");
  const [form, setForm] = useState<Form | null>(null);
  const [editing, setEditing] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  useEffect(() => {
    let active = true;
    api().then((data) => { if (active) setRows(data); }).catch((err) => { if (active) setError(err.message); }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, []);

  const filtered = rows.filter((row) => row.name.toLowerCase().includes(query.toLowerCase()) || row.ads_merchant_id.toLowerCase().includes(query.toLowerCase()));
  function open(row?: AdsPartner) {
    setError(""); setEditing(row?.id ?? "");
    setForm(row ? { name: row.name, ads_merchant_id: row.ads_merchant_id, api_key: row.api_key, security_type: row.security_type } : blank());
  }
  async function save(event: React.FormEvent) {
    event.preventDefault();
    if (!form || busy) return;
    const merchantID = form.ads_merchant_id.trim();
    if (!form.name.trim() || !merchantID || merchantID.length > 20 || form.api_key.length !== keyLengths[form.security_type]) { setError("请正确填写所有必填项"); return; }
    setBusy(true); setError(""); setNotice("");
    try {
      const saved: AdsPartner = await api(editing ? `/${editing}` : "", { method: editing ? "PATCH" : "POST", body: JSON.stringify({ ...form, name: form.name.trim(), ads_merchant_id: merchantID }) });
      setRows((old) => editing ? old.map((item) => item.id === editing ? saved : item) : [saved, ...old]);
      setForm(null); setNotice(editing ? "广告商已更新" : "广告商已新增");
    } catch (err) { setError(err instanceof Error ? err.message : "保存失败"); }
    finally { setBusy(false); }
  }
  async function remove(row: AdsPartner) {
    if (busy || !window.confirm(`确定删除广告商「${row.name}」？`)) return;
    setBusy(true); setError(""); setNotice("");
    try { await api(`/${row.id}`, { method: "DELETE" }); setRows((old) => old.filter((item) => item.id !== row.id)); setNotice("广告商已删除"); }
    catch (err) { setError(err instanceof Error ? err.message : "删除失败"); }
    finally { setBusy(false); }
  }

  return <PageFrame eyebrow="ADVERTISER MANAGEMENT" title="广告商配置" description="维护广告商的商户识别码与 API 加密配置。">
    <div className="cp-page">
      <section className="table-panel cp-filters">
        <label>名称或商户 ID<input value={query} placeholder="输入关键字筛选" onChange={(event) => setQuery(event.target.value)} /></label>
        <button className="secondary-button" onClick={() => setQuery("")}>清除</button>
      </section>
      {error && <p role="alert" className="cp-error">{error}</p>}
      {notice && <p role="status">{notice}</p>}
      <section className="table-panel">
        <div className="panel-toolbar"><h2>广告商列表</h2><button className="primary-button compact" disabled={busy} onClick={() => open()}>新增广告商</button></div>
        <div className="table-scroll"><table className="ap-table"><thead><tr>
          {["Name", "ads_merchant_id", "api_key", "security_type", "create_time", "update_time", "create_by", "update_by", "操作"].map((label) => <th key={label}>{label}</th>)}
        </tr></thead><tbody>
          {filtered.map((row) => <tr key={row.id}>
            <td>{row.name}</td><td>{row.ads_merchant_id}</td><td>{row.api_key}</td><td>{row.security_type}</td>
            <td>{formatTime(row.create_time)}</td><td>{formatTime(row.update_time)}</td><td>{row.create_by}</td><td>{row.update_by}</td>
            <td><div className="cp-row-actions"><button disabled={busy} onClick={() => open(row)}>编辑</button><button disabled={busy} onClick={() => void remove(row)}>删除</button></div></td>
          </tr>)}
          {!filtered.length && <tr><td colSpan={9} className="empty-state">{loading ? "加载中…" : "暂无广告商资料"}</td></tr>}
        </tbody></table></div>
      </section>
      {form && <div className="modal-backdrop"><section className="modal-card cp-modal" role="dialog" aria-modal="true" aria-labelledby="ap-title">
        <div className="cp-modal-header"><h2 id="ap-title">{editing ? "编辑" : "新增"}广告商</h2><button aria-label="关闭" disabled={busy} onClick={() => setForm(null)}>×</button></div>
        <form onSubmit={save}><fieldset disabled={busy}><div className="cp-form-grid">
          <label><span>Name *</span><input autoFocus required maxLength={150} value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} /></label>
          <label><span>ads_merchant_id *</span><div className="cp-code"><input required maxLength={20} placeholder="最多 20 位" value={form.ads_merchant_id} onChange={(event) => setForm({ ...form, ads_merchant_id: event.target.value })} /><button type="button" onClick={() => setForm({ ...form, ads_merchant_id: randomString(16) })}>随机生成</button></div></label>
          <label><span>security_type *</span><select required value={form.security_type} onChange={(event) => { const security_type = event.target.value; setForm({ ...form, security_type, api_key: security_type ? randomString(keyLengths[security_type]) : "" }); }}><option value="">请选择</option>{securityTypes.map((item) => <option key={item}>{item}</option>)}</select></label>
          <label><span>api_key *</span><div className="cp-code"><input required readOnly value={form.api_key} placeholder="选择加密方式后自动生成" /><button type="button" disabled={!form.security_type} onClick={() => setForm({ ...form, api_key: randomString(keyLengths[form.security_type]) })}>重新生成</button></div></label>
        </div></fieldset><div className="modal-actions"><button type="button" className="secondary-button" disabled={busy} onClick={() => setForm(null)}>取消</button><button className="primary-button compact" disabled={busy}>{busy ? "保存中…" : "保存"}</button></div></form>
      </section></div>}
    </div>
  </PageFrame>;
}
