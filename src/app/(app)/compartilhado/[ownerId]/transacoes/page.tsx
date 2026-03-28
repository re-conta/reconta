import { Header } from "@/components/layout/header";
import { TransacoesClient } from "@/components/transactions/transacoes-client";

export default function SharedTransacoesPage() {
	return (
		<>
			<Header
				title="Lançamentos"
				description="Visualizando lançamentos compartilhados"
			/>
			<TransacoesClient />
		</>
	);
}
