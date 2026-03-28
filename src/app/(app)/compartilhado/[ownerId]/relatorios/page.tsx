import { Header } from "@/components/layout/header";
import { RelatoriosClient } from "@/components/reports/relatorios-client";

export default function SharedRelatoriosPage() {
	return (
		<>
			<Header
				title="Relatórios"
				description="Visualizando relatórios compartilhados"
			/>
			<RelatoriosClient />
		</>
	);
}
