import type { db } from "@gcsim/types";

export type Submission = {
	id: string;
	status: "pending" | "approved" | "rejected";
	sourceUrl: string;
	submittedAt: number;
	reviewedAt: number | null;
	reviewedBy?: string;
	reason: string;
	publishedId: string | null;
	entry: db.Entry;
	duplicates?: db.Entry[];
};

export class ApiError extends Error {
	constructor(
		message: string,
		public status: number,
		public field?: string,
	) {
		super(message);
	}
}

export async function submissionApi<T>(
	path: string,
	options: {
		authorization?: string;
		body?: unknown;
		signal?: AbortSignal;
	} = {},
): Promise<T> {
	const response = await fetch(path, {
		method: options.body ? "POST" : "GET",
		credentials: "omit",
		signal: options.signal,
		headers: {
			...(options.authorization
				? { Authorization: options.authorization }
				: {}),
			...(options.body ? { "Content-Type": "application/json" } : {}),
		},
		body: options.body ? JSON.stringify(options.body) : undefined,
	});
	let data: { error?: string; field?: string };
	try {
		data = await response.json();
	} catch {
		throw new ApiError(
			"The response could not be read. Check your connection and try again.",
			response.status,
		);
	}
	if (!response.ok)
		throw new ApiError(
			data.error ?? "The request failed. Please try again.",
			response.status,
			data.field,
		);
	return data as T;
}

export const statusLabels = {
	pending: "Pending review",
	approved: "Approved",
	rejected: "Rejected",
};
export const fullDate = (value: number) =>
	new Date(value).toLocaleString(undefined, {
		dateStyle: "medium",
		timeStyle: "short",
	});
