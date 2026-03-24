import { Header } from "@/components/layout/header";
import { ContasBancariasClient } from "@/components/accounts/contas-bancarias-client";

export default function ContasBancariasPage() {
	return (
		<>
			<Header
				title="Contas Bancárias"
				description="Gerencie suas contas e carteiras"
			/>
			<ContasBancariasClient />
		</>
	);
}
