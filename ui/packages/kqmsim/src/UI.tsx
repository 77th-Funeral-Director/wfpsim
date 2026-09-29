import { dynamicKey } from "@gcsim/localization";
import { useTranslation } from "react-i18next";
import {
	defaultEditorPrefs,
	DefaultSampleOptions,
	Editor,
	ExecutorProvider,
	namedSeeds,
	ResultsView,
	SampleLog,
	SeedPicker,
	useExecutor,
	useRunResult,
	useValidation,
} from "@gcsim/components";
import {
	Button,
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@gcsim/primitives";
import type { Executor, ExecutorSupplier, model, Sample } from "@gcsim/types";
import { useLocalStorage } from "@gcsim/utils";
import { type ReactNode, useEffect, useState } from "react";

export function UI({
	exec,
	children,
}: {
	exec: ExecutorSupplier<Executor>;
	children: ReactNode;
	gitCommit: string;
	mode: string;
}) {
	return (
		<ExecutorProvider exec={exec}>
			<Workspace settings={children} />
		</ExecutorProvider>
	);
}

function Workspace({ settings }: { settings: ReactNode }) {
	const { t } = useTranslation();
	const { exec, isReady, busy, run, cancel } = useExecutor();
	const { result, error } = useRunResult();
	const [config, setConfig] = useLocalStorage("cfg", "");
	const [prefs, setPrefs] = useLocalStorage(
		"kqm-editor-prefs",
		defaultEditorPrefs,
	);
	const [tab, setTab] = useState("simulator");
	const [settingsOpen, setSettingsOpen] = useState(false);
	const [loaded, setLoaded] = useState<model.SimulationResult | null>(null);
	const [loadError, setLoadError] = useState("");
	const [loading, setLoading] = useState(false);
	const { isValid, error: configError, parsedTeam } = useValidation(config);
	const [share, setShare] = useState("");
	const [sharing, setSharing] = useState(false);
	const [shareError, setShareError] = useState("");
	const [copied, setCopied] = useState(false);
	const [sample, setSample] = useState<Sample | null>(null);
	const [seed, setSeed] = useState<string | null>(null);
	const [sampleError, setSampleError] = useState("");
	const [generating, setGenerating] = useState(false);
	const [sampleSettings, setSampleSettings] = useLocalStorage<string[]>(
		"kqm-sample-settings",
		DefaultSampleOptions,
	);
	const data = loaded ?? result;

	useEffect(() => {
		const match = window.location.pathname.match(
			/^\/sh\/([A-Za-z0-9_-]{1,150})$/,
		);
		if (!match) return;
		const controller = new AbortController();
		setLoading(true);
		setTab("results");
		fetch(`/api/share/${encodeURIComponent(match[1])}`, {
			signal: controller.signal,
		})
			.then(async (response) => {
				if (!response.ok)
					throw new Error(
						response.status === 404
							? "This share link has expired or does not exist."
							: "The result could not load. Try again.",
					);
				return response.json();
			})
			.then((value) => {
				if (!controller.signal.aborted) {
					setLoaded(value);
					setShare(window.location.href);
				}
			})
			.catch((reason) => {
				if (!controller.signal.aborted) setLoadError(String(reason.message));
			})
			.finally(() => {
				if (!controller.signal.aborted) setLoading(false);
			});
		return () => controller.abort();
	}, []);

	function start() {
		setLoaded(null);
		setLoadError("");
		setShare("");
		setShareError("");
		setSample(null);
		if (window.location.pathname.startsWith("/sh/"))
			window.history.replaceState(null, "", "/");
		run(config);
		setTab("results");
	}

	async function createShare() {
		if (!data || busy) return;
		setSharing(true);
		setShareError("");
		try {
			const response = await fetch("/api/share", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify(data),
			});
			if (!response.ok)
				throw new Error("The share link could not be saved. Try again.");
			const key = await response.text();
			if (!/^[A-Za-z0-9_-]{1,150}$/.test(key))
				throw new Error("The server returned an invalid share link.");
			setShare(`${window.location.origin}/sh/${key}`);
		} catch (reason) {
			setShareError(String(reason instanceof Error ? reason.message : reason));
		} finally {
			setSharing(false);
		}
	}

	async function generateSample(next: string) {
		if (!data?.config_file || busy) return;
		setGenerating(true);
		setSampleError("");
		setSeed(next);
		try {
			setSample(await exec().sample(data.config_file, next));
		} catch (reason) {
			setSampleError(String(reason));
		} finally {
			setGenerating(false);
		}
	}

	return (
		<main className="kqm-main">
			<div className="kqm-tabs" role="tablist" aria-label="Simulation">
				{["simulator", "results", "sample"].map((value) => (
					<Button
						key={value}
						role="tab"
						aria-selected={tab === value}
						variant={tab === value ? "default" : "outline"}
						disabled={value !== "simulator" && !data && !loading}
						onClick={() => setTab(value)}
					>
						{value[0].toUpperCase() + value.slice(1)}
					</Button>
				))}
			</div>
			{(error || loadError) && (
				<div role="alert" className="kqm-error">
					{error || loadError}
					<Button
						variant="outline"
						onClick={() => {
							cancel();
							setTab("simulator");
						}}
					>
						Return to simulator
					</Button>
				</div>
			)}
			{tab === "simulator" && (
				<>
					<p className="mb-4">
						KQM Sim includes the current gcsim characters and the KQM character
						changes.{" "}
						<a href="https://gist.github.com/77th-Funeral-Director/76e601b721559e0763c0898b323e79da">
							Read the assumptions for characters under development.
						</a>
					</p>
					<Editor
						config={config}
						setConfig={setConfig}
						parsedTeam={parsedTeam}
						error={configError}
						prefs={prefs}
						onPrefsChange={setPrefs}
						showThemeSelector
						canRun={isReady && isValid && !busy}
						busy={!isReady || busy}
						onRun={start}
						settings={
							<Button variant="outline" onClick={() => setSettingsOpen(true)}>
								Settings
							</Button>
						}
					/>
				</>
			)}
			{tab === "results" && (
				<>
					{loading && <p role="status">Loading simulation results…</p>}
					{busy && (
						<div role="status" className="kqm-actions">
							Running: {data?.statistics?.iterations ?? 0} /{" "}
							{data?.simulator_settings?.iterations ?? "…"}
							<Button onClick={cancel}>Stop</Button>
						</div>
					)}
					{data && (
						<>
							<div className="kqm-actions">
								<Button
									disabled={busy || sharing || Boolean(share)}
									onClick={createShare}
								>
									{sharing ? "Saving…" : "Create share link"}
								</Button>
								<Button
									variant="outline"
									onClick={async () => {
										try {
											await navigator.clipboard.writeText(
												data.config_file ?? "",
											);
											setCopied(true);
										} catch {
											setCopied(false);
										}
									}}
								>
									{copied ? "Copied" : "Copy config"}
								</Button>
								<Button
									variant="outline"
									onClick={() => {
										setConfig(data.config_file ?? "");
										setTab("simulator");
									}}
								>
									Edit config
								</Button>
							</div>
							{shareError && <p role="alert">{shareError}</p>}
							{share && (
								<div className="kqm-actions">
									<label>
										Share link{" "}
										<input
											aria-label="Share link"
											readOnly
											value={share}
											onFocus={(event) => event.target.select()}
										/>
									</label>
									<a
										href={`https://db.kqm.gg/submit?link=${encodeURIComponent(share)}`}
									>
										Submit to the database
									</a>
								</div>
							)}
							<ResultsView
								model={data}
								names={data.character_details?.map((c) =>
									t(dynamicKey(`game:character_names.${c.name}`)),
								)}
							/>
						</>
					)}
				</>
			)}
			{tab === "sample" && data && (
				<>
					<SeedPicker
						seeds={namedSeeds(data)}
						value={seed}
						onPick={generateSample}
						running={busy || generating || !isReady}
					/>
					{generating && <p role="status">Generating sample…</p>}
					{sampleError && <p role="alert">{sampleError}</p>}
					{sample && (
						<SampleLog
							sample={sample}
							settings={sampleSettings}
							onSettingsChange={setSampleSettings}
						/>
					)}
				</>
			)}
			<Dialog open={settingsOpen} onOpenChange={setSettingsOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Settings</DialogTitle>
					</DialogHeader>
					{settings}
				</DialogContent>
			</Dialog>
		</main>
	);
}
