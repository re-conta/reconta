import { Sidebar } from "@/components/layout/sidebar";

export default function AppLayout({ children }: { children: React.ReactNode }) {
	return (
		<div className="min-h-screen">
			<Sidebar />
			<main className="lg:ml-64 pt-14 lg:pt-0 h-screen lg:h-screen overflow-y-auto">
				<div className="p-4 lg:p-8">{children}</div>
			</main>
		</div>
	);
}
