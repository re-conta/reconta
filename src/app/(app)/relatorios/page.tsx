import { Header } from "@/components/layout/header";
import { RelatoriosClient } from "@/components/reports/relatorios-client";

export default function RelatoriosPage() {
	return (
		<>
			<Header
				title="Relatórios"
				description="Análise detalhada das suas finanças"
			/>
			<RelatoriosClient />
		</>
	);
}
