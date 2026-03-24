import { Header } from "@/components/layout/header";
import { RelatoriosClient } from "@/components/reports/relatorios-client";
import { getCurrentMonth } from "@/lib/utils";

export default function RelatoriosPage() {
	const { month, year } = getCurrentMonth();
	return (
		<>
			<Header
				title="Relatórios"
				description="Análise detalhada das suas finanças"
			/>
			<RelatoriosClient initialMonth={month} initialYear={year} />
		</>
	);
}
