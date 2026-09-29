import { Label, Switch } from "@gcsim/primitives";
import React from "react";
import { useTranslation } from "react-i18next";
import ServerMode from "./ServerMode";
import WasmMode from "./WasmMode";

const serverModeKey = "use-server-mode";

const App = () => {
	const { t } = useTranslation();
	const [serverMode, setServerMode] = React.useState<boolean>((): boolean => {
		return localStorage.getItem(serverModeKey) === "true";
	});
	React.useEffect(() => {
		localStorage.setItem(serverModeKey, serverMode.toString());
	}, [serverMode]);

	const children = (
		<div className="flex items-center gap-2">
			<Switch
				id="server-mode-switch"
				checked={serverMode}
				onCheckedChange={setServerMode}
			/>
			<Label htmlFor="server-mode-switch">
				{t(
					serverMode
						? "simple.server_mode_disable"
						: "simple.server_mode_enable",
				)}
			</Label>
		</div>
	);

	return (
		<>
			<header className="kqm-header">
				<a href="https://keqingmains.com">
					<img src="/kqm-logo.png" alt="KQM" width="48" height="48" />
				</a>
				<a href="/" className="kqm-title">
					KQM Sim
				</a>
				<nav aria-label="KQM sites">
					<a href="https://db.kqm.gg">Database</a>
					<a href="https://db.kqm.gg/submit">Submit a simulation</a>
				</nav>
			</header>
			{serverMode ? (
				<ServerMode>{children}</ServerMode>
			) : (
				<WasmMode>{children}</WasmMode>
			)}
			<footer className="kqm-footer">
				KQM Sim is based on gcsim.{" "}
				<a href="https://github.com/KQM-git/kqmsim">Source code (AGPL-3.0)</a>
			</footer>
		</>
	);
};

export default App;
