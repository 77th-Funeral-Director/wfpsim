import { execFileSync } from "node:child_process";
import { chmodSync, copyFileSync } from "node:fs";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("../../../../", import.meta.url));
const output = resolve(root, "ui/packages/kqmsim/public");
const goRoot = execFileSync("go", ["env", "GOROOT"], {
	encoding: "utf8",
}).trim();
copyFileSync(
	resolve(goRoot, "lib/wasm/wasm_exec.js"),
	resolve(output, "wasm_exec.js"),
);
chmodSync(resolve(output, "wasm_exec.js"), 0o644);
execFileSync(
	"go",
	["build", "-trimpath", "-o", resolve(output, "main.wasm"), "./cmd/wasm"],
	{
		cwd: root,
		env: { ...process.env, GOOS: "js", GOARCH: "wasm" },
		stdio: "inherit",
	},
);
