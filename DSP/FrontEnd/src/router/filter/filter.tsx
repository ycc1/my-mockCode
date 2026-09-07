import type { ReactNode } from "react";
import Login from "../../app/login/login";

type LoginFilterProps = {
  isAuthenticated: boolean;
  onAuthenticated: () => void;
  children: ReactNode;
};

export function LoginFilter({
  isAuthenticated,
  onAuthenticated,
  children,
}: LoginFilterProps) {
  return isAuthenticated ? (
    <>{children}</>
  ) : (
    <Login onAuthenticated={onAuthenticated} />
  );
}

export function FeatureFilter({
  enabled,
  children,
}: {
  enabled: boolean;
  children: ReactNode;
}) {
  return enabled ? <>{children}</> : null;
}
