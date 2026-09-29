import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";

import { initI18n } from "@gcsim/localization";
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
import "./index.css";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
	<React.StrictMode>
		<App />
	</React.StrictMode>,
);
