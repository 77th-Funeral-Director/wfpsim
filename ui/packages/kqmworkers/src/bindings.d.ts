export interface Env {
	ASSETS: Fetcher;
	kqmsim_r2: R2Bucket;
	kqmsim_kv: KVNamespace;
	KQMSIM_ASSETS_ENDPOINT: string;
	BROWSER: {
		quickAction(
			action: "screenshot",
			options: {
				url: string;
				viewport: { width: number; height: number; deviceScaleFactor: number };
				screenshotOptions: { type: "png" };
				gotoOptions: { waitUntil: "networkidle0"; timeout: number };
				waitForSelector: { selector: string; timeout: number };
			},
		): Promise<Response>;
	};
}
