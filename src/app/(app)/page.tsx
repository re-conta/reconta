import { DashboardClient } from "@/components/dashboard/dashboard-client";
import { Header } from "@/components/layout/header";

export default function DashboardPage() {
	return (
		<>
			<Header title="Dashboard" description="Visão geral das suas finanças" />
			<DashboardClient />
		</>
	);
}
