import { useState } from "react";
import type { FormEvent } from "react";
import { Activity, ArrowUpRight } from "lucide-react";

type LoginProps = { onAuthenticated: () => void };

export default function Login({ onAuthenticated }: LoginProps) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    try {
      const response = await fetch("/api/v1/membership/login", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
      });
      if (!response.ok) throw new Error("login failed");
      onAuthenticated();
    } catch {
      setError("登入失败，请检查帐号与密码");
    }
  }

  return (
    <main className="login-shell">
      <section className="login-aside">
        <div className="brand-mark">
          <Activity size={19} />
        </div>
        <p className="eyebrow">DSP OPERATIONS CONSOLE</p>
        <h1>
          把每一次
          <br />
          <em>增长</em>变成可见的系统。
        </h1>
        <p className="login-copy">
          统一管理广告商务、渠道网络、投放配置与归因数据，让运营团队在同一个节奏里工作。
        </p>
        <div className="aside-foot">
          <span className="pulse" />
          System ready · v1.0
        </div>
      </section>
      <section className="login-panel">
        <div className="login-box">
          <div className="mobile-brand">
            <div className="brand-mark">
              <Activity size={19} />
            </div>
            <span>DSP / OPS</span>
          </div>
          <p className="eyebrow">WELCOME BACK</p>
          <h2>登入后台</h2>
          <p className="muted">使用你的工作帐号进入操作台</p>
          <form onSubmit={submit}>
            <label>
              帐号
              <input
                value={username}
                onChange={(event) => setUsername(event.target.value)}
                placeholder="输入帐号"
                autoComplete="username"
                required
              />
            </label>
            <label>
              密码
              <input
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder="输入密码"
                type="password"
                autoComplete="current-password"
                required
              />
            </label>
            {error && <p className="form-error">{error}</p>}
            <button className="primary-button" type="submit">
              进入操作台 <ArrowUpRight size={17} />
            </button>
          </form>
          <p className="login-note">需要协助？联系系统管理员</p>
        </div>
      </section>
    </main>
  );
}
