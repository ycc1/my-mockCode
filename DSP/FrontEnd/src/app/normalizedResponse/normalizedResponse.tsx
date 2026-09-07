import { useState } from "react";
import { FileJson, Send, TriangleAlert } from "lucide-react";
import { PageFrame } from "../account/account";

type Tab = "request" | "response" | "error";
const tabs: { id: Tab; label: string; icon: typeof Send }[] = [
  { id: "request", label: "Request Model", icon: Send },
  { id: "response", label: "Response Model", icon: FileJson },
  { id: "error", label: "Error Model", icon: TriangleAlert },
];

export default function NormalizedResponsePage() {
  const [activeTab, setActiveTab] = useState<Tab>("request");
  return (
    <PageFrame
      eyebrow="ADVERTISER / EVENT NORMALIZATION"
      title="上报规范化回应"
      description="统一广告事件上报格式、标准字段与 API 回应规范。"
    >
      <section className="api-meta-panel table-panel">
        <div className="api-meta-item">
          <small>Method</small>
          <strong>POST</strong>
        </div>
        <div className="api-meta-item">
          <small>Endpoint</small>
          <strong>/api/v1/events</strong>
        </div>
        <div className="api-meta-item">
          <small>Content-Type</small>
          <strong>application/json</strong>
        </div>
        <div className="api-meta-item">
          <small>Response</small>
          <strong>JSON</strong>
        </div>
      </section>
      <section className="table-panel api-doc-panel">
        <div className="api-tabs">
          {tabs.map(({ id, label, icon: Icon }) => (
            <button
              className={`api-tab ${activeTab === id ? "active" : ""}`}
              key={id}
              onClick={() => setActiveTab(id)}
            >
              <Icon size={15} />
              {label}
            </button>
          ))}
        </div>
        {activeTab === "request" && <RequestModel />}
        {activeTab === "response" && <ResponseModel />}
        {activeTab === "error" && <ErrorModel />}
      </section>
      <section className="table-panel api-note-panel">
        <strong>处理说明</strong>
        <p>
          API
          回传成功代表事件已被平台接收，后续仍会进行去重、归因、状态判断与渠道回传处理；可使用
          event_id / request_id 追踪每笔事件。
        </p>
      </section>
    </PageFrame>
  );
}

function RequestModel() {
  return (
    <div className="api-doc-body">
      <h2>上报资料字段</h2>
      <table className="api-field-table">
        <thead>
          <tr>
            <th>Field</th>
            <th>Type</th>
            <th>Required</th>
            <th>说明</th>
            <th>Example</th>
          </tr>
        </thead>
        <tbody>
          {[
            [
              "event_id",
              "string",
              "Yes",
              "事件唯一 ID，用于去重",
              "evt_202608310001",
            ],
            ["event_name", "string", "Yes", "标准事件名称", "first_deposit"],
            [
              "event_time",
              "datetime",
              "Yes",
              "事件发生时间",
              "2026-08-31T18:10:25+08:00",
            ],
            ["channel_id", "string", "Yes", "渠道号", "4780097"],
            ["campaign_id", "string", "No", "投放 Campaign ID", "CMP_BR_001"],
            ["click_id", "string", "No", "广告点击归因 ID", "clk_8fd92ab123"],
            [
              "uid",
              "string",
              "Yes",
              "甲方使用者唯一 ID",
              "4168830563102226944",
            ],
            [
              "order_id",
              "string",
              "Conditional",
              "交易事件必填",
              "DEP202608310001",
            ],
            ["amount", "decimal", "Conditional", "交易金额", "100.00"],
            ["currency", "string", "Conditional", "币别", "BRL"],
            ["country", "string", "No", "国家代码", "BR"],
            ["properties", "object", "No", "事件扩充参数", '{"is_first":true}'],
          ].map((row) => (
            <tr key={row[0]}>
              {row.map((cell, index) => (
                <td
                  className={
                    index === 2 && cell === "Yes" ? "required-cell" : ""
                  }
                  key={`${row[0]}-${index}`}
                >
                  {cell}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      <h2>Request Example</h2>
      <pre className="api-code">{`{
  "event_id": "evt_202608310001",
  "event_name": "first_deposit",
  "event_time": "2026-08-31T18:10:25+08:00",
  "channel_id": "4780097",
  "uid": "4168830563102226944",
  "amount": 100.00,
  "currency": "BRL",
  "properties": { "is_first": true }
}`}</pre>
    </div>
  );
}
function ResponseModel() {
  return (
    <div className="api-doc-body">
      <h2>成功回应字段</h2>
      <table className="api-field-table">
        <thead>
          <tr>
            <th>Field</th>
            <th>Type</th>
            <th>说明</th>
            <th>Example</th>
          </tr>
        </thead>
        <tbody>
          {[
            ["success", "boolean", "是否成功接收", "true"],
            ["code", "string", "标准结果码", "SUCCESS"],
            ["message", "string", "回应讯息", "Event received successfully"],
            ["data.event_id", "string", "原始事件 ID", "evt_202608310001"],
            [
              "data.request_id",
              "string",
              "平台追踪 Request ID",
              "req_c57e1289",
            ],
            ["data.status", "string", "事件接收状态", "received"],
            [
              "data.received_at",
              "datetime",
              "平台接收时间",
              "2026-08-31T18:10:26+08:00",
            ],
          ].map((row) => (
            <tr key={row[0]}>
              {row.map((cell) => (
                <td key={`${row[0]}-${cell}`}>{cell}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      <h2>Response Example</h2>
      <pre className="api-code">{`{
  "success": true,
  "code": "SUCCESS",
  "message": "Event received successfully",
  "data": { "event_id": "evt_202608310001", "request_id": "req_c57e1289", "status": "received" }
}`}</pre>
    </div>
  );
}
function ErrorModel() {
  return (
    <div className="api-doc-body">
      <h2>Error Code</h2>
      <table className="api-field-table">
        <thead>
          <tr>
            <th>HTTP</th>
            <th>Code</th>
            <th>说明</th>
          </tr>
        </thead>
        <tbody>
          {[
            ["400", "INVALID_REQUEST", "JSON 或 Request 格式错误"],
            ["400", "INVALID_PARAMETER", "参数错误或必填字段缺失"],
            ["400", "UNSUPPORTED_EVENT", "不支持的事件名称"],
            ["401", "UNAUTHORIZED", "API Key / Signature 错误"],
            ["403", "CHANNEL_DISABLED", "渠道已停用"],
            ["409", "DUPLICATE_EVENT", "event_id 重复"],
            ["429", "RATE_LIMITED", "超过允许的上报频率"],
            ["500", "INTERNAL_ERROR", "平台内部异常"],
          ].map((row) => (
            <tr key={row[1]}>
              {row.map((cell) => (
                <td key={`${row[1]}-${cell}`}>{cell}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      <h2>Error Response Example</h2>
      <pre className="api-code">{`{
  "success": false,
  "code": "INVALID_PARAMETER",
  "message": "Invalid request parameter",
  "errors": [{ "field": "currency", "reason": "currency is required" }]
}`}</pre>
    </div>
  );
}
