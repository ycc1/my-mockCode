import { Check, Plus, ShieldCheck, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { PageFrame } from "../account/account";

type Role = { role_id: string; name: string; description?: string };
type Feature = { feature_id: string; code: string; name: string };
type FeatureGroup = { resource: string; name: string; actions: Feature[] };
const featureResources = [
  ["offer", "广告内容与投放配置"],
  ["ads_partner", "广告商配置"],
  ["channel_partner", "渠道商配置"],
  ["channel_number", "渠道号配置"],
  ["channel_link", "渠道链接配置"],
  ["role", "角色管理"],
  ["feature", "功能权限管理"],
  ["report", "报表"],
  ["report.upload_log", "报表 / 上报日志"],
  ["report.settlement", "报表 / 结算报表"],
  ["report.attribution", "报表 / 归因报表"],
  ["account", "帐号管理"],
] as const;
const actions = ["create", "read", "update", "delete"] as const;
const fallbackFeatures: Feature[] = featureResources.flatMap(
  ([resource, name]) =>
    (resource === "report" || resource.startsWith("report.")
      ? ["read"]
      : actions
    ).map((action) => ({
      feature_id: "",
      code: `${resource}.${action}`,
      name: `${name} / ${action}`,
    })),
);
const actionLabels: Record<string, string> = {
  create: "新增",
  read: "查询",
  update: "修改",
  delete: "删除",
};
const resourceLabels: Record<string, string> = {
  offer: "广告内容与投放配置",
  ads_partner: "广告商配置",
  channel_partner: "渠道商配置",
  channel_number: "渠道号配置",
  channel_link: "渠道链接配置",
  role: "角色管理",
  feature: "功能权限管理",
  report: "报表",
  "report.upload_log": "报表 / 上报日志",
  "report.settlement": "报表 / 结算报表",
  "report.attribution": "报表 / 归因报表",
  account: "帐号管理",
};

export default function RolesPage() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [features, setFeatures] = useState<Feature[]>(fallbackFeatures);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [selectedFeatures, setSelectedFeatures] = useState<string[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [roleName, setRoleName] = useState("");
  const [message, setMessage] = useState("");
  const featureGroups = features.reduce<FeatureGroup[]>((groups, feature) => {
    const resource = feature.code.slice(0, feature.code.lastIndexOf("."));
    const group = groups.find((item) => item.resource === resource);
    if (group) group.actions.push(feature);
    else
      groups.push({
        resource,
        name: resourceLabels[resource] ?? resource,
        actions: [feature],
      });
    return groups;
  }, []);

  async function load() {
    const [roleResponse, featureResponse] = await Promise.all([
      fetch("/api/v1/roles", { credentials: "include" }),
      fetch("/api/v1/features", { credentials: "include" }),
    ]);
    if (roleResponse.ok) {
      const payload = await roleResponse.json();
      setRoles(payload.data ?? []);
    }
    if (featureResponse.ok) {
      const payload = await featureResponse.json();
      if (payload.data?.length) {
        const received = (payload.data as Feature[]).filter(
          (feature) =>
            !feature.code.startsWith("report.") ||
            feature.code.endsWith(".read"),
        );
        setFeatures([
          ...fallbackFeatures.map(
            (feature) =>
              received.find((item) => item.code === feature.code) ?? feature,
          ),
          ...received.filter(
            (feature) =>
              !fallbackFeatures.some((item) => item.code === feature.code),
          ),
        ]);
      }
    }
  }

  useEffect(() => {
    void load();
  }, []);
  useEffect(() => {
    if (!selectedRole && roles.length) setSelectedRole(roles[0]);
  }, [roles, selectedRole]);
  useEffect(() => {
    async function loadPermissions() {
      if (!selectedRole) return;
      const response = await fetch(`/api/v1/roles/${selectedRole.role_id}`, {
        credentials: "include",
      });
      if (!response.ok) return;
      const payload = await response.json();
      setSelectedFeatures(
        (payload.data?.features ?? []).map(
          (feature: Feature) => feature.feature_id,
        ),
      );
    }
    void loadPermissions();
  }, [selectedRole]);

  async function createRole() {
    if (!roleName.trim()) return;
    const response = await fetch("/api/v1/roles", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: roleName.trim() }),
    });
    if (!response.ok) return;
    const payload = await response.json();
    setRoles((current) => [...current, payload.data]);
    setSelectedRole(payload.data);
    setRoleName("");
    setShowCreate(false);
  }

  async function savePermissions() {
    if (!selectedRole) return;
    const response = await fetch(`/api/v1/roles/${selectedRole.role_id}`, {
      method: "PUT",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ feature_ids: selectedFeatures }),
    });
    setMessage(response.ok ? "权限已保存" : "权限保存失败");
    window.setTimeout(() => setMessage(""), 2200);
  }

  async function deleteRole() {
    if (
      !selectedRole ||
      !window.confirm(`确定删除角色「${selectedRole.name}」？`)
    )
      return;
    const response = await fetch(`/api/v1/roles/${selectedRole.role_id}`, {
      method: "DELETE",
      credentials: "include",
    });
    if (response.ok) {
      setRoles((current) =>
        current.filter((role) => role.role_id !== selectedRole.role_id),
      );
      setSelectedRole(null);
    }
  }

  return (
    <PageFrame
      eyebrow="SYSTEM / ACCESS CONTROL"
      title="角色管理"
      description="建立角色并勾选可使用的功能权限。"
    >
      <div className="roles-layout">
        <section className="table-panel role-list">
          <div className="panel-toolbar">
            <div>
              <h2>角色</h2>
              <p>{roles.length} 个已建立角色</p>
            </div>
            <button
              className="icon-button light-icon"
              onClick={() => setShowCreate(true)}
              title="新增角色"
            >
              <Plus size={17} />
            </button>
          </div>
          {roles.map((role) => (
            <button
              className={`role-item ${selectedRole?.role_id === role.role_id ? "active" : ""}`}
              key={role.role_id}
              onClick={() => setSelectedRole(role)}
            >
              <span className="role-icon">
                <ShieldCheck size={16} />
              </span>
              <span>
                <strong>{role.name}</strong>
                <small>{role.description || role.role_id}</small>
              </span>
            </button>
          ))}
          {!roles.length && <p className="empty-state">尚未建立角色</p>}
        </section>
        <section className="table-panel permission-panel">
          <div className="panel-toolbar">
            <div>
              <h2>{selectedRole?.name || "选择一个角色"}</h2>
              <p>
                勾选此角色可使用的功能；报表及其子功能仅支持查询，其他功能支持新增、查询、修改、删除
              </p>
            </div>
            {selectedRole && (
              <div className="permission-actions">
                <button
                  className="icon-button danger-icon"
                  onClick={deleteRole}
                  title="删除角色"
                >
                  <Trash2 size={16} />
                </button>
                <button
                  className="primary-button compact"
                  onClick={savePermissions}
                >
                  保存权限
                </button>
              </div>
            )}
          </div>
          <div className="permission-grid">
            {featureGroups.map((group) => (
              <div className="permission-group" key={group.resource}>
                <div className="permission-group-title">
                  <strong>{group.name}</strong>
                  <small>{group.resource}</small>
                </div>
                {actions.map((action) => {
                  const feature = group.actions.find((item) =>
                    item.code.endsWith(`.${action}`),
                  );
                  if (!feature) return null;
                  return (
                    <label className="permission-item" key={feature.code}>
                      <input
                        type="checkbox"
                        disabled={!selectedRole || !feature.feature_id}
                        checked={selectedFeatures.includes(feature.feature_id)}
                        onChange={() =>
                          setSelectedFeatures((current) =>
                            current.includes(feature.feature_id)
                              ? current.filter(
                                  (id) => id !== feature.feature_id,
                                )
                              : [...current, feature.feature_id],
                          )
                        }
                      />
                      <span className="check-box">
                        <Check size={13} />
                      </span>
                      <span>
                        <strong>{actionLabels[action]}</strong>
                        <small>{feature.code}</small>
                        {!feature.feature_id && (
                          <small>后端尚未提供此权限，请更新并重启 API</small>
                        )}
                      </span>
                    </label>
                  );
                })}
              </div>
            ))}
          </div>
          {message && <p className="save-message">{message}</p>}
        </section>
      </div>
      {showCreate && (
        <div className="modal-backdrop" onClick={() => setShowCreate(false)}>
          <div
            className="modal-card"
            onClick={(event) => event.stopPropagation()}
          >
            <h2>新增角色</h2>
            <p className="muted">建立后即可配置功能权限。</p>
            <label className="modal-label">
              角色名称
              <input
                autoFocus
                value={roleName}
                onChange={(event) => setRoleName(event.target.value)}
                placeholder="例如：渠道运营"
              />
            </label>
            <div className="modal-actions">
              <button
                className="secondary-button"
                onClick={() => setShowCreate(false)}
              >
                取消
              </button>
              <button className="primary-button compact" onClick={createRole}>
                建立角色
              </button>
            </div>
          </div>
        </div>
      )}
    </PageFrame>
  );
}
