import { craftQuery, type DbQuery } from "SharedHooks/databaseQuery";
import {
	Alert,
	AlertDescription,
	AlertTitle,
	Button,
	Spinner,
} from "@gcsim/primitives";
import type { db } from "@gcsim/types";
import axios from "axios";
import { useCallback, useEffect, useReducer, useRef, useState } from "react";
import {
	initialFilter as defaultFilter,
	FilterContext,
	FilterDispatchContext,
	type FilterState,
	filterReducer,
} from "../../SharedComponents/FilterComponents/Filter.utils";
import { DBView } from "./DBVIew";

type Props = {
	initialFilter?: FilterState;
};

export const Database = ({ initialFilter = defaultFilter }: Props) => {
	const [filter, dispatch] = useReducer(filterReducer, initialFilter);
	const [data, setData] = useState<db.Entry[]>([]);
	const [isLoading, setIsLoading] = useState<boolean>(true);
	const [hasMore, setHasMore] = useState<boolean>(true);
	const [page, setPage] = useState<number>(1);
	const [error, setError] = useState<string | null>(null);
	const abortController = useRef(new AbortController());

	// TODO(react19): drop this useCallback (inline the function) once the React
	// Compiler is enabled — it auto-memoizes. Keep the functional setData
	// updater regardless: it fixes a stale-closure bug, independent of memoization.
	const appendData = useCallback((next: db.Entry[]) => {
		// let d = [ ...data,...next.filter(e => {
		//   return false
		// })]
		setData((prev) => [...prev, ...next]);
	}, []);

	// TODO(react19): drop this useCallback (inline the function) once the React
	// Compiler is enabled — it auto-memoizes. Don't remove it before then:
	// querydb must stay referentially stable or the effect below refetch-loops,
	// and useExhaustiveDependencies is now an error.
	const querydb = useCallback(
		(query: DbQuery, nextPage: number, append: boolean) => {
			setError(null);
			const signal = abortController.current.signal;
			axios(`/api/db?q=${encodeURIComponent(JSON.stringify(query))}`, {
				signal,
			})
				.then((resp: { data: db.Entries }) => {
					if (signal.aborted) return;
					if (resp.data?.data) {
						setPage(nextPage);
						setHasMore(true);
						if (append) {
							appendData(resp.data.data);
						} else {
							setData(resp.data.data);
						}
						//check count; if we got less than limit then there's no more data...
						//TODO: this is bugged if there are exactly limit number of entries...
						//TODO: really server should tell us if there's more data
						if (resp.data.data.length < query.limit) {
							setHasMore(false);
						}
					} else {
						setHasMore(false);
						if (!append) {
							setData([]);
						}
					}
					setIsLoading(false);
				})
				.catch((err) => {
					if (signal.aborted || axios.isCancel(err)) return;
					setIsLoading(false);
					setHasMore(false);
					setError(
						err.response?.status === 400
							? "This search is not supported. Change the search or reset the filters."
							: "The database could not load. Please try again.",
					);
				});
		},
		[appendData],
	);

	useEffect(() => {
		abortController.current.abort();
		abortController.current = new AbortController();
		const query = craftQuery(filter, 1, 25);
		querydb(query, 1, false);
		return () => abortController.current.abort();
	}, [filter, querydb]);

	const fetchData = () => {
		const nextPage = page + 1;
		const query = craftQuery(filter, nextPage, 25);
		querydb(query, nextPage, true);
	};

	if (isLoading || !data)
		return (
			<div className="h-screen flex flex-col justify-center items-center">
				<Spinner className="size-12" />
			</div>
		);

	return (
		<FilterContext.Provider value={filter}>
			<FilterDispatchContext.Provider value={dispatch}>
				{error && (
					<Alert
						variant="destructive"
						className="mx-auto my-4 max-w-[1096px]"
						role="alert"
					>
						<AlertTitle>Unable to load simulations</AlertTitle>
						<AlertDescription>{error}</AlertDescription>
						<Button
							className="mt-3"
							onClick={() => querydb(craftQuery(filter, 1, 25), 1, false)}
						>
							Try again
						</Button>
					</Alert>
				)}
				<DBView data={data} fetchData={fetchData} hasMore={hasMore} />
			</FilterDispatchContext.Provider>
		</FilterContext.Provider>
	);
};
