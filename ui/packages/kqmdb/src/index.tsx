import { initI18n } from "@gcsim/localization";
import App from "./App";

// all the css styling we need (except tailwind)
import "@gcsim/components/src/index.css";
import "./index.css";

import React from "react";
import ReactDOM from "react-dom/client";

initI18n().addResourceBundle(
	"en",
	"game",
	{
		character_names: {
			alyosha: "Alyosha",
			linnea: "Linnea",
			lohen: "Lohen",
			sandrone: "Sandrone",
			vesna: "Vesna",
			vodyanitsa: "Vodyanitsa",
			zibai: "Zibai",
		},
	},
	true,
	true,
);

const root = ReactDOM.createRoot(
	document.getElementById("root") as HTMLElement,
);
root.render(
	<React.StrictMode>
		<App />
	</React.StrictMode>,
);
