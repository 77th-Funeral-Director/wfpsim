import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import * as git from "git-rev-sync";
import { defineConfig } from "vite";
process.env.VITE_GIT_COMMIT_HASH = git.long(import.meta.dirname);
process.env.VITE_GIT_BRANCH = git.branch(import.meta.dirname);
export default defineConfig({
 plugins: [tailwindcss(), react()],
 resolve: { tsconfigPaths: true },
 server: { proxy: { "/api": { target: "http://localhost:8788", changeOrigin: true } } },
});
