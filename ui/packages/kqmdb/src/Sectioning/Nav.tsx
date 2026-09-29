import {
	Navbar,
	NavbarDivider,
	NavbarGroup,
	NavbarHeading,
} from "@gcsim/components";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@gcsim/primitives";
import { useTranslation } from "react-i18next";
import { Link } from "wouter";

const LANGUAGES = [
	{ value: "en", key: "nav.english" },
	{ value: "zh", key: "nav.chinese" },
	{ value: "ja", key: "nav.japanese" },
	{ value: "ko", key: "nav.korean" },
	{ value: "es", key: "nav.spanish" },
	{ value: "ru", key: "nav.russian" },
	{ value: "de", key: "nav.german" },
] as const;

export default function Nav() {
	const { t, i18n } = useTranslation();

	return (
		<Navbar className="h-16 border-b border-g-line-soft bg-g-surface">
			<div className="mx-auto flex w-full max-w-[1160px] px-4 sm:px-8">
				<NavbarHeading className="!mr-[10px]">
					<Link
						href="/"
						className="flex h-16 items-center gap-3"
						aria-label="KQM Sim Database home"
					>
						<img
							src="/kqm-logo.png"
							alt=""
							className="h-11 w-11 object-contain"
						/>
						<span className="font-g-display text-g-ink font-bold">
							KQM Sim <span className="text-g-accent">Database</span>
						</span>
					</Link>
				</NavbarHeading>
				<NavbarGroup className="min-[550px]:flex items-stretch">
					<NavbarDivider />
				</NavbarGroup>
				<NavbarGroup align="end">
					<Select
						value={i18n.resolvedLanguage}
						onValueChange={(value) => i18n.changeLanguage(value)}
					>
						<SelectTrigger aria-label="Language">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							{LANGUAGES.map(({ value, key }) => (
								<SelectItem key={value} value={value}>
									{t(key)}
								</SelectItem>
							))}
						</SelectContent>
					</Select>
				</NavbarGroup>
			</div>
		</Navbar>
	);
}
