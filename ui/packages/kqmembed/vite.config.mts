import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import * as git from "git-rev-sync";
import { defineConfig } from "vite";
process.env.VITE_GIT_COMMIT_HASH = git.long(import.meta.dirname);
process.env.VITE_GIT_BRANCH = git.branch(import.meta.dirname);
export default defineConfig({
 base: "/embed/",
 build: { outDir: "../kqmsim/dist/embed", emptyOutDir: true },
 plugins: [tailwindcss(), react()],
 resolve: { tsconfigPaths: true },
 server: { proxy: { "/api": { target: "https://sim.kqm.gg", changeOrigin: true } } },
});
