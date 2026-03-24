import { DashboardClient } from "@/components/dashboard/dashboard-client";
import { Header } from "@/components/layout/header";
import { getCurrentMonth } from "@/lib/utils";

export default function DashboardPage() {
	const { month, year } = getCurrentMonth();
	return (
		<>
			<Header title="Dashboard" description="Visão geral das suas finanças" />
			<DashboardClient initialMonth={month} initialYear={year} />
		</>
	);
}
