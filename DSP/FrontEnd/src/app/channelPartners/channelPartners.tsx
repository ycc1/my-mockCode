import { useEffect, useState } from "react";
import { PageFrame } from "../account/account";
import "./channelPartners.css";

const fields = [
  ["partner_type", "渠道商类型"],
  ["service_category", "服务类别"],
  ["traffic_model", "流量模式"],
  ["billing_model", "计费模式"],
  ["primary_channel", "主要渠道"],
  ["secondary_channel", "次要渠道"],
  ["market_contract", "市场覆盖&合同"],
  ["delivery_package", "投放包体"],
  ["data_system", "数据回传系统"],
] as const;
type Form = {
  name: string;
  code: string;
  merchant_id: string;
  api_key: string;
  security_type: string;
  status: boolean;
  values: Record<string, string | string[]>;
};
type Partner = {
  channel_partner_id: string;
  name: string;
  code: string;
  status: boolean;
  created_at: string;
} & Record<string, unknown>;
const blank = (): Form => ({
  name: "",
  code: "",
  merchant_id: "",
  api_key: "",
  security_type: "",
  status: true,
  values: Object.fromEntries(
    fields.flatMap(([key]) => [
      [key, isMultiple(key) ? [] : ""],
      [`${key}_other`, ""],
    ]),
  ),
});
const isMultiple = (key: string) =>
  key === "delivery_package" || key === "data_system";
const asArray = (value: unknown): string[] =>
  Array.isArray(value) ? value.map(String) : value ? [String(value)] : [];
const hasOther = (value: string | string[]) =>
  Array.isArray(value) ? value.includes("其他") : value === "其他";
const display = (row: Partner, key: string) =>
  asArray(row[key])
    .map((value) =>
      value === "其他" ? `其他：${row[`${key}_other`] ?? ""}` : value,
    )
    .join("、") || "—";
const apiKeyLengths: Record<string, number> = {
  MD5: 32,
  SHA312: 78,
  AES: 32,
  DES: 8,
};
const generateRandomString = (length: number) => {
  const alphabet =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  const random = new Uint32Array(length);
  crypto.getRandomValues(random);
  return Array.from(random, (value) => alphabet[value % alphabet.length]).join(
    "",
  );
};
const generateApiKey = (securityType: string) =>
  generateRandomString(apiKeyLengths[securityType] ?? 0);
async function api(path = "", options?: RequestInit) {
  const response = await fetch(`/api/v1/channel-partners${path}`, {
    credentials: "include",
    ...options,
    headers: { "Content-Type": "application/json", ...options?.headers },
  });
  const payload = await response.json();
  if (!response.ok)
    throw new Error(
      response.status === 401
        ? "登录已过期，请重新登录"
        : response.status === 403
          ? "没有此操作的权限"
          : payload.message || "操作失败",
    );
  return payload.data;
}

export default function ChannelPartnersPage() {
  const [rows, setRows] = useState<Partner[]>([]);
  const [options, setOptions] = useState<Record<string, string[]>>({});
  const [ready, setReady] = useState(false);
  const [reload, setReload] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [name, setName] = useState("");
  const [status, setStatus] = useState("");
  const [filter, setFilter] = useState({ name: "", status: "" });
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(10);
  const [form, setForm] = useState<Form | null>(null);
  const [editing, setEditing] = useState("");
  const [busy, setBusy] = useState(false);
  const [formError, setFormError] = useState("");
  useEffect(() => {
    let active = true;
    Promise.all([api(), api("/options")])
      .then(([data, choices]) => {
        if (
          !choices ||
          fields.some(
            ([key]) =>
              !Array.isArray(choices[key]) ||
              choices[key].length === 0 ||
              choices[key].some((value: unknown) => typeof value !== "string"),
          ) ||
          !Array.isArray(choices.security_type) ||
          choices.security_type.length === 0
        ) {
          throw new Error("下拉配置不完整，请重试或联系管理员");
        }
        if (active) {
          setRows(data);
          setOptions(choices);
          setReady(true);
        }
      })
      .catch((err) => {
        if (active) setError(err.message);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [reload]);
  const filtered = rows.filter(
    (row) =>
      row.name.toLowerCase().includes(filter.name.toLowerCase()) &&
      (!filter.status || String(row.status) === filter.status),
  );
  const pages = Math.max(1, Math.ceil(filtered.length / size));
  const current = Math.min(page, pages);
  function open(row?: Partner) {
    if (!ready) return;
    setEditing(row?.channel_partner_id ?? "");
    setFormError("");
    setForm(
      row
        ? {
            name: row.name,
            code: row.code,
            merchant_id: String(row.merchant_id ?? ""),
            api_key: String(row.api_key ?? ""),
            security_type: String(row.security_type ?? ""),
            status: row.status,
            values: Object.fromEntries(
              fields.flatMap(([key]) => [
                [
                  key,
                  isMultiple(key) ? asArray(row[key]) : String(row[key] ?? ""),
                ],
                [`${key}_other`, String(row[`${key}_other`] ?? "")],
              ]),
            ),
          }
        : blank(),
    );
  }
  async function save(event: React.FormEvent) {
    event.preventDefault();
    if (!form || busy) return;
    if (
      !form.name.trim() ||
      !form.merchant_id.trim() ||
      form.merchant_id.trim().length > 20 ||
      !form.api_key ||
      form.api_key.length !== apiKeyLengths[form.security_type] ||
      fields.some(
        ([key]) =>
          !form.values[key].length ||
          (hasOther(form.values[key]) &&
            !String(form.values[`${key}_other`]).trim()),
      )
    ) {
      setFormError("请填写所有必填项及其他选项的补充内容");
      return;
    }
    setBusy(true);
    setFormError("");
    try {
      const saved: Partner = await api(editing ? `/${editing}` : "", {
        method: editing ? "PATCH" : "POST",
        body: JSON.stringify({
          name: form.name.trim(),
          code: form.code,
          merchant_id: form.merchant_id.trim(),
          api_key: form.api_key,
          security_type: form.security_type,
          status: form.status,
          ...form.values,
        }),
      });
      setRows((old) =>
        editing
          ? old.map((row) => (row.channel_partner_id === editing ? saved : row))
          : [saved, ...old],
      );
      setForm(null);
      setNotice(editing ? "渠道商已更新" : "渠道商已新增");
    } catch (err) {
      setFormError(err instanceof Error ? err.message : "保存失败，请重试");
    } finally {
      setBusy(false);
    }
  }
  async function mutate(row: Partner, remove = false) {
    if (busy || (remove && !window.confirm(`确定删除渠道商「${row.name}」？`)))
      return;
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const saved = await api(`/${row.channel_partner_id}`, {
        method: remove ? "DELETE" : "PATCH",
        ...(remove ? {} : { body: JSON.stringify({ status: !row.status }) }),
      });
      setRows((old) =>
        remove
          ? old.filter(
              (item) => item.channel_partner_id !== row.channel_partner_id,
            )
          : old.map((item) =>
              item.channel_partner_id === row.channel_partner_id ? saved : item,
            ),
      );
      setNotice(remove ? "渠道商已删除" : "状态已更新");
    } catch (err) {
      setError(err instanceof Error ? err.message : "操作失败");
    } finally {
      setBusy(false);
    }
  }
  return (
    <PageFrame
      eyebrow="CHANNEL MANAGEMENT"
      title="渠道商配置"
      description="维护渠道商资料、投放模式与数据回传配置。"
    >
      <div className="cp-page">
        <form
          className="table-panel cp-filters"
          onSubmit={(event) => {
            event.preventDefault();
            setFilter({ name: name.trim(), status });
            setPage(1);
          }}
        >
          <label>
            渠道商名称
            <input
              placeholder="请输入关键词"
              value={name}
              onChange={(event) => setName(event.target.value)}
            />
          </label>
          <label>
            是否有效
            <select
              value={status}
              onChange={(event) => setStatus(event.target.value)}
            >
              <option value="">全部</option>
              <option value="true">启用</option>
              <option value="false">停用</option>
            </select>
          </label>
          <button className="primary-button compact">查询</button>
          <button
            type="button"
            className="secondary-button"
            onClick={() => {
              setName("");
              setStatus("");
              setFilter({ name: "", status: "" });
              setPage(1);
            }}
          >
            重置
          </button>
        </form>
        {error && (
          <p role="alert" className="cp-error">
            {error}{" "}
            {!ready && (
              <button
                disabled={loading}
                onClick={() => {
                  setError("");
                  setLoading(true);
                  setReload((value) => value + 1);
                }}
              >
                重新加载
              </button>
            )}
          </p>
        )}
        {notice && <p role="status">{notice}</p>}
        <section className="table-panel">
          <div className="panel-toolbar">
            <h2>渠道商列表</h2>
            <button
              className="primary-button compact"
              disabled={loading || busy || !ready}
              onClick={() => open()}
            >
              新增渠道商
            </button>
          </div>
          <div className="table-scroll">
            <table className="cp-table">
              <thead>
                <tr>
                  {[
                    "序号",
                    "渠道商名称",
                    "渠道商代号",
                    "商户 ID",
                    "API Key",
                    "加密方式",
                    "是否有效",
                    "渠道商类型",
                    "服务类别",
                    "流量模式",
                    "主要/次要渠道",
                    "市场覆盖&合同",
                    "投放包体",
                    "数据回传系统",
                    "计费模式",
                    "创建时间",
                    "操作",
                  ].map((label) => (
                    <th key={label}>{label}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {filtered
                  .slice((current - 1) * size, current * size)
                  .map((row, index) => (
                    <tr key={row.channel_partner_id}>
                      <td>{(current - 1) * size + index + 1}</td>
                      <td>{row.name}</td>
                      <td>{row.code}</td>
                      <td>{String(row.merchant_id ?? "")}</td>
                      <td>{String(row.api_key ?? "")}</td>
                      <td>{String(row.security_type ?? "")}</td>
                      <td>
                        <button
                          className={`cp-switch ${row.status ? "on" : ""}`}
                          role="switch"
                          aria-checked={row.status}
                          aria-label={`${row.name}是否有效`}
                          disabled={busy}
                          onClick={() => void mutate(row)}
                        >
                          {row.status ? "启用" : "停用"}
                        </button>
                      </td>
                      <td>{display(row, "partner_type")}</td>
                      <td>{display(row, "service_category")}</td>
                      <td>{display(row, "traffic_model")}</td>
                      <td>
                        {display(row, "primary_channel")} /{" "}
                        {display(row, "secondary_channel")}
                      </td>
                      <td>{display(row, "market_contract")}</td>
                      <td>{display(row, "delivery_package")}</td>
                      <td>{display(row, "data_system")}</td>
                      <td>{display(row, "billing_model")}</td>
                      <td>
                        {new Date(row.created_at).toLocaleString("zh-CN", {
                          hour12: false,
                        })}
                      </td>
                      <td>
                        <div className="cp-row-actions">
                          <button disabled={busy} onClick={() => open(row)}>
                            编辑
                          </button>
                          <button
                            disabled={busy}
                            onClick={() => void mutate(row, true)}
                          >
                            删除
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                {!filtered.length && (
                  <tr>
                    <td colSpan={17} className="empty-state">
                      {loading
                        ? "加载中…"
                        : error
                          ? "数据加载失败"
                          : "暂无符合条件的渠道商"}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
          <div className="cp-pagination">
            <span>共 {filtered.length} 条记录</span>
            <button
              disabled={current <= 1}
              onClick={() => setPage(current - 1)}
            >
              上一页
            </button>
            <span>
              {current} / {pages}
            </span>
            <button
              disabled={current >= pages}
              onClick={() => setPage(current + 1)}
            >
              下一页
            </button>
            <select
              aria-label="每页条数"
              value={size}
              onChange={(event) => {
                setSize(Number(event.target.value));
                setPage(1);
              }}
            >
              {[10, 20, 50].map((value) => (
                <option key={value} value={value}>
                  {value}条/页
                </option>
              ))}
            </select>
          </div>
        </section>
        {form && (
          <div className="modal-backdrop">
            <section
              className="modal-card cp-modal"
              role="dialog"
              aria-modal="true"
              aria-labelledby="cp-title"
            >
              <div className="cp-modal-header">
                <h2 id="cp-title">{editing ? "编辑" : "新增"}渠道商信息</h2>
                <button
                  aria-label="关闭"
                  disabled={busy}
                  onClick={() => setForm(null)}
                >
                  ×
                </button>
              </div>
              <form onSubmit={save}>
                <fieldset disabled={busy}>
                  <div className="cp-form-grid">
                    <label>
                      <span>渠道商名称 *</span>
                      <input
                        autoFocus
                        required
                        maxLength={150}
                        value={form.name}
                        placeholder="请输入渠道商名称"
                        onChange={(event) =>
                          setForm({ ...form, name: event.target.value })
                        }
                      />
                    </label>
                    <label>
                      <span>渠道商代号 *</span>
                      <div className="cp-code">
                        <input
                          required
                          inputMode="numeric"
                          pattern="[0-9]{3}"
                          maxLength={3}
                          title="请输入三位数字代号"
                          placeholder="三位数字代号"
                          value={form.code}
                          onChange={(event) =>
                            setForm({ ...form, code: event.target.value })
                          }
                        />
                        <button
                          type="button"
                          onClick={() => {
                            const used = new Set(
                              rows
                                .filter(
                                  (row) => row.channel_partner_id !== editing,
                                )
                                .map((row) => row.code),
                            );
                            const available = Array.from(
                              { length: 1000 },
                              (_, i) => String(i).padStart(3, "0"),
                            ).filter((code) => !used.has(code));
                            if (!available.length) {
                              setFormError("三位数字代号已用完");
                              return;
                            }
                            setForm({
                              ...form,
                              code: available[
                                Math.floor(Math.random() * available.length)
                              ],
                            });
                          }}
                        >
                          随机生成
                        </button>
                      </div>
                    </label>
                    <label>
                      <span>商户 ID *</span>
                      <div className="cp-code">
                        <input
                          required
                          maxLength={20}
                          value={form.merchant_id}
                          placeholder="最多 20 位"
                          onChange={(event) =>
                            setForm({ ...form, merchant_id: event.target.value })
                          }
                        />
                        <button
                          type="button"
                          onClick={() =>
                            setForm({
                              ...form,
                              merchant_id: generateRandomString(16),
                            })
                          }
                        >
                          自动生成
                        </button>
                      </div>
                    </label>
                    <label>
                      <span>加密方式 *</span>
                      <select
                        required
                        value={form.security_type}
                        onChange={(event) => {
                          const security_type = event.target.value;
                          setForm({
                            ...form,
                            security_type,
                            api_key: security_type
                              ? generateApiKey(security_type)
                              : "",
                          });
                        }}
                      >
                        <option value="">请选择加密方式</option>
                        {(options.security_type ?? []).map((option) => (
                          <option key={option} value={option}>
                            {option}
                          </option>
                        ))}
                      </select>
                    </label>
                    <label>
                      <span>API Key *</span>
                      <div className="cp-code">
                        <input
                          required
                          readOnly
                          value={form.api_key}
                          placeholder="选择加密方式后自动生成"
                        />
                        <button
                          type="button"
                          disabled={!form.security_type}
                          onClick={() =>
                            setForm({
                              ...form,
                              api_key: generateApiKey(form.security_type),
                            })
                          }
                        >
                          重新生成
                        </button>
                      </div>
                    </label>
                    {fields.map(([key, label]) => (
                      <div className="cp-field" key={key}>
                        <label htmlFor={`cp-${key}`}>{label} *</label>
                        {isMultiple(key) ? (
                          <details className="cp-multiselect">
                            <summary id={`cp-${key}`} aria-label={label}>
                              {asArray(form.values[key]).join("、") ||
                                `请选择${label}（可多选）`}
                            </summary>
                            <div className="cp-options">
                              {(options[key] ?? []).map((option) => (
                                <label key={option}>
                                  <input
                                    type="checkbox"
                                    checked={asArray(form.values[key]).includes(
                                      option,
                                    )}
                                    onChange={(event) => {
                                      const selected = asArray(
                                        form.values[key],
                                      );
                                      const next = event.target.checked
                                        ? [...selected, option]
                                        : selected.filter(
                                            (value) => value !== option,
                                          );
                                      setForm({
                                        ...form,
                                        values: {
                                          ...form.values,
                                          [key]: next,
                                          [`${key}_other`]: next.includes(
                                            "其他",
                                          )
                                            ? form.values[`${key}_other`]
                                            : "",
                                        },
                                      });
                                    }}
                                  />
                                  {option}
                                </label>
                              ))}
                            </div>
                          </details>
                        ) : (
                          <select
                            id={`cp-${key}`}
                            required
                            value={form.values[key]}
                            onChange={(event) =>
                              setForm({
                                ...form,
                                values: {
                                  ...form.values,
                                  [key]: event.target.value,
                                  [`${key}_other`]: "",
                                },
                              })
                            }
                          >
                            <option value="">请选择{label}</option>
                            {(options[key] ?? []).map((option) => (
                              <option key={option} value={option}>
                                {option}
                              </option>
                            ))}
                          </select>
                        )}
                        {hasOther(form.values[key]) && (
                          <input
                            required
                            maxLength={150}
                            aria-label={`${label}补充内容`}
                            placeholder="请填写补充内容"
                            value={form.values[`${key}_other`]}
                            onChange={(event) =>
                              setForm({
                                ...form,
                                values: {
                                  ...form.values,
                                  [`${key}_other`]: event.target.value,
                                },
                              })
                            }
                          />
                        )}
                      </div>
                    ))}
                    <label>
                      <span>状态</span>
                      <button
                        type="button"
                        className={`cp-switch ${form.status ? "on" : ""}`}
                        role="switch"
                        aria-checked={form.status}
                        onClick={() =>
                          setForm({ ...form, status: !form.status })
                        }
                      >
                        {form.status ? "启用" : "停用"}
                      </button>
                    </label>
                  </div>
                </fieldset>
                <p className="muted">
                  标注 * 为必填项；选择“其他”时须填写补充内容。
                </p>
                {formError && (
                  <p role="alert" className="cp-error">
                    {formError}
                  </p>
                )}
                <div className="modal-actions">
                  <button
                    type="button"
                    className="secondary-button"
                    disabled={busy}
                    onClick={() => setForm(null)}
                  >
                    取消
                  </button>
                  <button className="primary-button compact" disabled={busy}>
                    {busy ? "保存中…" : "确定"}
                  </button>
                </div>
              </form>
            </section>
          </div>
        )}
      </div>
    </PageFrame>
  );
}
