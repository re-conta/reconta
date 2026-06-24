"use client";

import {
	AlertCircle,
	ArrowLeft,
	BarChart3,
	BookOpen,
	Download,
	Eye,
	Home,
	LogOut,
	Menu,
	Settings,
	Share2,
	Tag,
	Tags,
	Upload,
	Wallet,
	X,
} from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState } from "react";
import { signOut, useSession } from "@/lib/auth-client";
import { cn } from "@/lib/utils";
import { useSharedOwner } from "@/components/layout/shared-owner-context";

const navigation = [
	{ name: "Dashboard", href: "/", icon: Home },
	{ name: "Lançamentos", href: "/transacoes", icon: BookOpen },
	{ name: "Contas Fixas", href: "/contas", icon: AlertCircle },
	{ name: "Contas Bancárias", href: "/contas-bancarias", icon: Wallet },
	{ name: "Categorias", href: "/categorias", icon: Tags },
	{ name: "Tags", href: "/tags", icon: Tag },
	{ name: "Relatórios", href: "/relatorios", icon: BarChart3 },
	{ name: "Compartilhamentos", href: "/compartilhamentos", icon: Share2 },
	{ name: "Importar Extrato", href: "/importar", icon: Upload },
	{ name: "Exportar", href: "/exportar", icon: Download },
	{ name: "Ajustes", href: "/ajustes", icon: Settings },
];

/** Items available when viewing shared data (read-only) */
const sharedNavigation = [
	{ name: "Dashboard", href: "", icon: Home },
	{ name: "Lançamentos", href: "/transacoes", icon: BookOpen },
	{ name: "Contas Fixas", href: "/contas", icon: AlertCircle },
	{ name: "Categorias", href: "/categorias", icon: Tags },
	{ name: "Tags", href: "/tags", icon: Tag },
	{ name: "Relatórios", href: "/relatorios", icon: BarChart3 },
	{ name: "Contas Bancárias", href: "/contas-bancarias", icon: Wallet },
	{ name: "Exportar", href: "/exportar", icon: Download },
];

export function Sidebar() {
	const pathname = usePathname();
	const router = useRouter();
	const { data: session } = useSession();
	const [mobileOpen, setMobileOpen] = useState(false);
	const shared = useSharedOwner();

	const isSharedMode = !!shared;
	const navItems = isSharedMode
		? sharedNavigation.map((item) => ({
				...item,
				href: `${shared.routeBase}${item.href}`,
			}))
		: navigation;

	async function handleSignOut() {
		try {
			await signOut();
		} catch {
			// ignore network errors during sign-out
		}
		router.push("/login");
	}

	return (
		<>
			{/* Mobile top bar */}
			<header className="lg:hidden fixed inset-x-0 top-0 z-50 flex h-14 items-center gap-3 border-b border-zinc-800 bg-zinc-950 px-4 font-nunito">
				<button
					type="button"
					onClick={() => setMobileOpen(true)}
					className="rounded-md p-1.5 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800"
					aria-label="Abrir menu"
				>
					<Menu className="h-5 w-5" />
				</button>
				<div className="flex items-center gap-2">
					<Image
						src="/images/favicon.svg"
						alt="ReConta"
						width={32}
						height={32}
						loading="eager"
					/>
					<span className="text-base font-bold text-white">ReConta</span>
				</div>
				{isSharedMode && shared.owner.ownerName && (
					<span className="ml-auto text-xs text-amber-400 flex items-center gap-1">
						<Eye className="h-3.5 w-3.5" />
						{shared.owner.ownerName}
					</span>
				)}
			</header>

			{/* Overlay */}
			{mobileOpen && (
				<button
					type="button"
					aria-label="Fechar menu"
					className="lg:hidden fixed inset-0 z-40 bg-black/60 cursor-default"
					onClick={() => setMobileOpen(false)}
				/>
			)}

			{/* Sidebar */}
			<aside
				className={cn(
					"fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r border-zinc-800 bg-background transition-transform duration-200 font-nunito",
					isSharedMode && "border-r-amber-600/30",
					mobileOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0",
				)}
			>
				{/* Logo */}
				<div className="flex h-16 items-center justify-between gap-2 px-6 border-b border-zinc-800">
					<div className="flex items-center gap-2">
						<Image
							src="/images/favicon.svg"
							alt="ReConta"
							width={32}
							height={32}
							loading="eager"
						/>
						<div>
							<span className="text-lg font-bold text-white">ReConta</span>
						</div>
					</div>
					<button
						type="button"
						onClick={() => setMobileOpen(false)}
						className="lg:hidden rounded-md p-1 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800"
						aria-label="Fechar menu"
					>
						<X className="h-5 w-5" />
					</button>
				</div>

				{/* Shared-mode banner */}
				{isSharedMode && (
					<div className="mx-3 mt-3 rounded-lg bg-amber-500/10 border border-amber-500/20 px-3 py-2.5">
						<div className="flex items-center gap-2 text-amber-400 mb-1">
							<Eye className="h-4 w-4 shrink-0" />
							<span className="text-xs font-semibold uppercase tracking-wide">
								Visualizando conta
							</span>
						</div>
						<p className="text-sm font-medium text-amber-200 truncate">
							{shared.owner.ownerName || "Carregando..."}
						</p>
						{shared.owner.ownerEmail && (
							<p className="text-xs text-amber-400/70 truncate">
								{shared.owner.ownerEmail}
							</p>
						)}
						<Link
							href="/compartilhamentos"
							onClick={() => setMobileOpen(false)}
							className="mt-2 flex items-center gap-1.5 text-xs text-amber-400 hover:text-amber-300 transition-colors"
						>
							<ArrowLeft className="h-3 w-3" />
							Voltar para minha conta
						</Link>
					</div>
				)}

				{/* Navigation */}
				<nav className="flex-1 overflow-y-auto p-4">
					<ul className="space-y-1">
						{navItems.map((item) => {
							const isActive = isSharedMode
								? item.href === shared.routeBase
									? pathname === shared.routeBase ||
										pathname === `${shared.routeBase}/`
									: pathname === item.href ||
										pathname.startsWith(`${item.href}/`)
								: item.href === "/"
									? pathname === "/"
									: pathname === item.href ||
										pathname.startsWith(`${item.href}/`);
							return (
								<li key={item.href}>
									<Link
										href={item.href}
										onClick={() => setMobileOpen(false)}
										className={cn(
											"flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
											isActive
												? isSharedMode
													? "bg-amber-600/20 text-amber-400"
													: "bg-indigo-600/20 text-indigo-400"
												: "text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100",
										)}
									>
										<item.icon className="h-4 w-4 shrink-0" />
										{item.name}
									</Link>
								</li>
							);
						})}
					</ul>
				</nav>

				{/* User + Logout */}
				<div className="p-4 border-t border-zinc-800 space-y-3">
					{session?.user && (
						<div className="flex items-center gap-2.5 px-2">
							{session.user.image ? (
								<Image
									src={session.user.image}
									alt={session.user.name ?? "Avatar"}
									width={32}
									height={32}
									unoptimized
									className="h-8 w-8 rounded-full object-cover shrink-0"
								/>
							) : (
								<div className="h-8 w-8 rounded-full bg-indigo-600 flex items-center justify-center text-sm font-bold text-white shrink-0">
									{session.user.name?.[0]?.toUpperCase() ?? "U"}
								</div>
							)}
							<div className="flex-1 min-w-0">
								<p className="text-xs font-medium text-zinc-200 truncate">
									{session.user.name}
								</p>
								<p className="text-xs text-zinc-500 truncate">
									{session.user.email}
								</p>
							</div>
						</div>
					)}
					<button
						type="button"
						onClick={handleSignOut}
						className="w-full flex items-center gap-2 px-3 py-2 text-sm text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 rounded-lg transition-colors"
					>
						<LogOut className="h-4 w-4" />
						Sair
					</button>
				</div>
			</aside>
		</>
	);
}
