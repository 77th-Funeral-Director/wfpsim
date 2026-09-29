// Translate the database UI's Mongo-style filters to bound D1 queries.
// Only known fields and operators can become SQL. All values remain parameters.
const fields = {
	_id: "id",
	create_date: "create_date",
	"summary.mean_dps_per_target": "dps",
	"summary.sim_duration.mean": "duration",
	description: "json_extract(document, '$.description')",
	submitter: "json_extract(document, '$.submitter')",
	// The public JSON omits mode 0 (fixed duration), its protobuf default.
	"summary.mode": "COALESCE(json_extract(document, '$.summary.mode'), 0)",
	"summary.target_count": "json_extract(document, '$.summary.target_count')",
};
const arrays = {
	"summary.char_names": "$.summary.char_names",
	accepted_tags: "$.accepted_tags",
};
export class QueryError extends Error {}

export function compileQuery(input = {}) {
	const params = [];
	let clauses = 0;
	function bind(value) {
		if (
			typeof value !== "string" &&
			(typeof value !== "number" || !Number.isFinite(value))
		) {
			throw new QueryError("Filter values must be text or numbers.");
		}
		if (params.length >= 90 || String(value).length > 2000)
			throw new QueryError("The search is too large.");
		params.push(value);
		return "?";
	}
	function comparison(column, condition) {
		if (typeof condition !== "object" || condition === null)
			return `${column} = ${bind(condition)}`;
		const comparisons = Object.entries(condition)
			.filter(([op]) => op !== "$options")
			.map(([op, value]) => {
				const operators = {
					$eq: "=",
					$ne: "!=",
					$gt: ">",
					$gte: ">=",
					$lt: "<",
					$lte: "<=",
				};
				if (Object.hasOwn(operators, op))
					return `${column} ${operators[op]} ${bind(value)}`;
				if (op === "$in" || op === "$nin") {
					if (!Array.isArray(value) || value.length > 40)
						throw new QueryError("Invalid list filter.");
					if (value.length === 0) return op === "$in" ? "0" : "1";
					return `${column} ${op === "$nin" ? "NOT " : ""}IN (${value.map(bind).join(",")})`;
				}
				if (op === "$regex") {
					// The standard search box escapes regex symbols. D1 uses literal
					// substring matching; reject actual regex rather than change its meaning.
					if (
						typeof value !== "string" ||
						/(?<!\\)[.*+?^${}()|[\]]/.test(value) ||
						/\\[^.*+?^${}()|[\]\\]/.test(value)
					) {
						throw new QueryError(
							"Regular expressions are not supported. Use a plain text search.",
						);
					}
					if (condition.$options && condition.$options !== "i")
						throw new QueryError("Unsupported search option.");
					const literal = value.replace(/\\([.*+?^${}()|[\]\\])/g, "$1");
					return condition.$options === "i"
						? `instr(lower(${column}), lower(${bind(literal)})) > 0`
						: `instr(${column}, ${bind(literal)}) > 0`;
				}
				throw new QueryError(`Unsupported filter operator: ${op}`);
			});
		if (!comparisons.length) throw new QueryError("Empty filter.");
		return `(${comparisons.join(" AND ")})`;
	}
	function filter(query, depth = 0) {
		if (
			!query ||
			typeof query !== "object" ||
			Array.isArray(query) ||
			depth > 8
		)
			throw new QueryError("Invalid filter.");
		return (
			Object.entries(query)
				.map(([key, condition]) => {
					if (++clauses > 60) throw new QueryError("Too many filters.");
					if (key === "$sampleRate") {
						if (
							typeof condition !== "number" ||
							!Number.isFinite(condition) ||
							condition < 0 ||
							condition > 1
						)
							throw new QueryError("Sample rate must be between 0 and 1.");
						// Sample each row before applying the requested sort and limit.
						// Use 31 random bits to produce a value in [0, 1).
						return `((random() & 2147483647) / 2147483648.0 < ${bind(condition)})`;
					}
					if (key === "$and" || key === "$or") {
						if (!Array.isArray(condition) || condition.length === 0)
							throw new QueryError("Invalid filter group.");
						return `(${condition.map((v) => filter(v, depth + 1)).join(key === "$and" ? " AND " : " OR ")})`;
					}
					if (Object.hasOwn(arrays, key)) {
						const negative =
							condition &&
							typeof condition === "object" &&
							(Object.hasOwn(condition, "$ne") ||
								Object.hasOwn(condition, "$nin"));
						if (negative && Object.keys(condition).length !== 1)
							throw new QueryError(
								"Use separate array filters for each condition.",
							);
						const positive = negative
							? Object.hasOwn(condition, "$ne")
								? { $eq: condition.$ne }
								: { $in: condition.$nin }
							: condition;
						return `${negative ? "NOT " : ""}EXISTS (SELECT 1 FROM json_each(document, '${arrays[key]}') WHERE ${comparison("value", positive)})`;
					}
					if (!Object.hasOwn(fields, key))
						throw new QueryError(`Unsupported filter field: ${key}`);
					return comparison(fields[key], condition);
				})
				.join(" AND ") || "1"
		);
	}
	if (!input || typeof input !== "object" || Array.isArray(input))
		throw new QueryError("Invalid search.");
	const where = filter(input.query ?? {});
	const sort = input.sort ?? { create_date: -1 };
	if (
		typeof sort !== "object" ||
		Array.isArray(sort) ||
		Object.keys(sort).length > 3
	)
		throw new QueryError("Invalid sort.");
	const order = Object.entries(sort).map(([key, direction]) => {
		if (!Object.hasOwn(fields, key) || ![1, -1].includes(direction))
			throw new QueryError("Invalid sort.");
		return `${fields[key]} ${direction === 1 ? "ASC" : "DESC"}`;
	});
	const limit = input.limit ?? 25;
	const skip = input.skip ?? 0;
	if (
		!Number.isInteger(limit) ||
		limit < 1 ||
		limit > 100 ||
		!Number.isInteger(skip) ||
		skip < 0 ||
		skip > 100000
	)
		throw new QueryError("Invalid page.");
	params.push(limit, skip);
	return {
		sql: `SELECT document FROM simulations WHERE visible = 1 AND (${where}) ORDER BY ${[...order, "id ASC"].join(", ")} LIMIT ? OFFSET ?`,
		params,
	};
}
