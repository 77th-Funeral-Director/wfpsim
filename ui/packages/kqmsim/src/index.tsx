import { initI18n } from "@gcsim/localization";
import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import { kqmCharacterNames } from "./characters";

initI18n().addResourceBundle(
	"en",
	"game",
	{
		character_names: kqmCharacterNames,
	},
	true,
	true,
);
import "@gcsim/components/src/index.css";
import "./index.css";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
	<React.StrictMode>
		<App />
	</React.StrictMode>,
);
