import { useState } from "react";
import { LoginFilter } from "../router/filter/filter";
import Router from "../router/router";
import "../App.css";

export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  async function logout() {
    await fetch("/api/v1/membership/logout", {
      method: "POST",
      credentials: "include",
    }).catch(() => undefined);
    setIsAuthenticated(false);
  }

  return (
    <LoginFilter
      isAuthenticated={isAuthenticated}
      onAuthenticated={() => setIsAuthenticated(true)}
    >
      <Router onLogout={logout} />
    </LoginFilter>
  );
}
