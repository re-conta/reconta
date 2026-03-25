"use client";

import {
	AlertCircle,
	BarChart3,
	BookOpen,
	Download,
	Home,
	LogOut,
	Tags,
	Upload,
	Wallet,
} from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { signOut, useSession } from "@/lib/auth-client";
import { cn } from "@/lib/utils";

const navigation = [
	{ name: "Dashboard", href: "/", icon: Home },
	{ name: "Lançamentos", href: "/transacoes", icon: BookOpen },
	{ name: "Contas Fixas", href: "/contas", icon: AlertCircle },
	{ name: "Relatórios", href: "/relatorios", icon: BarChart3 },
	{ name: "Importar PDF", href: "/importar", icon: Upload },
	{ name: "Exportar", href: "/exportar", icon: Download },
	{ name: "Categorias", href: "/categorias", icon: Tags },
	{ name: "Contas Bancárias", href: "/contas-bancarias", icon: Wallet },
];

export function Sidebar() {
	const pathname = usePathname();
	const router = useRouter();
	const { data: session } = useSession();

	async function handleSignOut() {
		await signOut();
		router.push("/login");
	}

	return (
		<aside className="fixed inset-y-0 left-0 z-40 flex w-64 flex-col border-r border-zinc-800 bg-zinc-950">
			{/* Logo */}
			<div className="flex h-16 items-center gap-2 px-6 border-b border-zinc-800">
				<div className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-600">
					<span className="text-sm font-bold text-white">R</span>
				</div>
				<div>
					<span className="text-lg font-bold text-white">ReConta</span>
					<span className="ml-0.5 text-xs text-zinc-400">.app</span>
				</div>
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
						<div className="h-8 w-8 rounded-full bg-indigo-600 flex items-center justify-center text-sm font-bold text-white shrink-0">
							{session.user.name?.[0]?.toUpperCase() ?? "U"}
						</div>
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
	);
}
