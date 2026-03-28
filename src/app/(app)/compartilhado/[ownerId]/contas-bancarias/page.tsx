import { Header } from "@/components/layout/header";
import { ContasBancariasClient } from "@/components/accounts/contas-bancarias-client";

export default function SharedContasBancariasPage() {
	return (
		<>
			<Header
				title="Contas Bancárias"
				description="Visualizando contas bancárias compartilhadas"
			/>
			<ContasBancariasClient />
		</>
	);
}
