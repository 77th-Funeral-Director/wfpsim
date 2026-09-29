import { Home } from "Pages/Home";
import { lazy, Suspense, useEffect } from "react";
import { Route, Switch, useLocation } from "wouter";
import { Database } from "./Pages/Database";
import Layout from "./Sectioning/layout";

const Results = lazy(() => import("./Pages/Results"));
const Submit = lazy(() => import("./Pages/Submit"));
const Submission = lazy(() => import("./Pages/Submission"));
const Review = lazy(() => import("./Pages/Review"));
export default function App() {
	const [location] = useLocation();

	// wouter's client-side nav preserves window scroll, so the listing can mount at the bottom and
	// make react-infinite-scroll fetch pages forever. Reset scroll on every route change.
	useEffect(() => {
		window.scrollTo(0, 0);
		if (location === "/submit")
			document.title = "Submit a simulation — KQM Sim Database";
		else if (location.startsWith("/review"))
			document.title = "Review submissions — KQM Sim Database";
		else if (
			location.startsWith("/submission/") &&
			!location.endsWith("/results")
		)
			document.title = "Submission status — KQM Sim Database";
		else if (!location.startsWith("/db/") && !location.endsWith("/results"))
			document.title = "KQM Sim Database";
	}, [location]);

	return (
		<Layout>
			<Switch>
				<Route path="/">
					<Home />
				</Route>
				<Route path="/database">
					<Database />
				</Route>
				<Route path="/submit">
					<Suspense
						fallback={
							<p className="p-8" role="status">
								Loading submission form…
							</p>
						}
					>
						<Submit />
					</Suspense>
				</Route>
				<Route path="/submission/:id/results">
					{({ id }) => (
						<Suspense
							fallback={
								<p className="p-8" role="status">
									Loading results…
								</p>
							}
						>
							<Results id={id} submission />
						</Suspense>
					)}
				</Route>
				<Route path="/submission/:id">
					{({ id }) => (
						<Suspense
							fallback={
								<p className="p-8" role="status">
									Loading submission…
								</p>
							}
						>
							<Submission id={id} />
						</Suspense>
					)}
				</Route>
				<Route path="/review/:id?">
					{({ id }) => (
						<Suspense
							fallback={
								<p className="p-8" role="status">
									Loading review page…
								</p>
							}
						>
							<Review id={id} />
						</Suspense>
					)}
				</Route>
				<Route path="/db/:id">
					{(params) => (
						<Suspense fallback={<p className="p-8">Loading results…</p>}>
							<Results id={params.id} />
						</Suspense>
					)}
				</Route>
				<Route>
					<main className="p-8 text-center">
						Page not found.{" "}
						<a className="text-g-accent" href="/database">
							Browse the database
						</a>
						.
					</main>
				</Route>
			</Switch>
		</Layout>
	);
}
