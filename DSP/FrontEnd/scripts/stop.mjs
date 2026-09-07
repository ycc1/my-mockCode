import { existsSync, readFileSync, unlinkSync } from "node:fs";
import { resolve } from "node:path";
import { execFileSync } from "node:child_process";

const pidFile = resolve(".vite.pid");
if (!existsSync(pidFile)) {
  console.log("Frontend is not running.");
  process.exit(0);
}

const pid = Number(readFileSync(pidFile, "utf8"));
try {
  if (process.platform === "win32") {
    execFileSync("taskkill", ["/PID", String(pid), "/T", "/F"], {
      stdio: "ignore",
    });
  } else {
    process.kill(-pid, "SIGTERM");
  }
  console.log(`Frontend stopped (PID ${pid}).`);
} catch {
  console.log(`Frontend process ${pid} was already stopped.`);
} finally {
  unlinkSync(pidFile);
}
