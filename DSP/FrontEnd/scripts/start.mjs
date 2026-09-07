import { existsSync, readFileSync, unlinkSync, writeFileSync } from "node:fs";
import { spawn } from "node:child_process";
import { resolve } from "node:path";

const pidFile = resolve(".vite.pid");
const inputArgs = process.argv.slice(2);
const portIndex = inputArgs.findIndex(
  (argument) => argument === "--port" || argument === "-p",
);
const positionalPort = inputArgs.find((argument) => /^\d+$/.test(argument));
const port =
  portIndex >= 0
    ? inputArgs[portIndex + 1]
    : positionalPort || process.env.PORT || "5173";

if (!/^\d+$/.test(port) || Number(port) < 1 || Number(port) > 65535) {
  console.error(`Invalid port: ${port}`);
  process.exit(1);
}

if (existsSync(pidFile)) {
  const oldPid = Number(readFileSync(pidFile, "utf8"));
  if (oldPid > 0) {
    console.error(
      `Vite may already be running with PID ${oldPid}. Run npm run stop first.`,
    );
    process.exit(1);
  }
  unlinkSync(pidFile);
}

const viteCli = resolve("node_modules/vite/bin/vite.js");
const viteArgs = [viteCli, "--host", "127.0.0.1", "--port", port];
const child = spawn(process.execPath, viteArgs, {
  detached: true,
  stdio: "ignore",
  shell: false,
});

writeFileSync(pidFile, String(child.pid));
child.unref();
console.log(`Frontend started at http://127.0.0.1:${port} (PID ${child.pid})`);
