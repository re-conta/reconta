"use client";

import {
	AlertCircle,
	BarChart3,
	BookOpen,
	Download,
	Home,
	LogOut,
	Menu,
	Settings,
	Share2,
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

const navigation = [
	{ name: "Dashboard", href: "/", icon: Home },
	{ name: "Lançamentos", href: "/transacoes", icon: BookOpen },
	{ name: "Contas Fixas", href: "/contas", icon: AlertCircle },
	{ name: "Relatórios", href: "/relatorios", icon: BarChart3 },
	{ name: "Importar Extrato", href: "/importar", icon: Upload },
	{ name: "Exportar", href: "/exportar", icon: Download },
	{ name: "Categorias", href: "/categorias", icon: Tags },
	{ name: "Contas Bancárias", href: "/contas-bancarias", icon: Wallet },
	{ name: "Compartilhamentos", href: "/compartilhamentos", icon: Share2 },
	{ name: "Ajustes", href: "/ajustes", icon: Settings },
];

export function Sidebar() {
	const pathname = usePathname();
	const router = useRouter();
	const { data: session } = useSession();
	const [mobileOpen, setMobileOpen] = useState(false);

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
			<header className="lg:hidden fixed inset-x-0 top-0 z-50 flex h-14 items-center gap-3 border-b border-zinc-800 bg-zinc-950 px-4">
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
					"fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r border-zinc-800 bg-zinc-950 transition-transform duration-200",
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

				{/* Navigation */}
				<nav className="flex-1 overflow-y-auto p-4">
					<ul className="space-y-1">
						{navigation.map((item) => {
							const isActive =
								item.href === "/"
									? pathname === "/"
									: pathname.startsWith(item.href);
							return (
								<li key={item.href}>
									<Link
										href={item.href}
										onClick={() => setMobileOpen(false)}
										className={cn(
											"flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
											isActive
												? "bg-indigo-600/20 text-indigo-400"
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
