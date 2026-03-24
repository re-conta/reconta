import { Header } from "@/components/layout/header";
import { ContasClient } from "@/components/bills/contas-client";
import { getCurrentMonth } from "@/lib/utils";

export default function ContasPage() {
	const { month, year } = getCurrentMonth();
	return (
		<>
			<Header
				title="Contas Fixas"
				description="Alertas de contas recorrentes e pagamentos mensais"
			/>
			<ContasClient initialMonth={month} initialYear={year} />
		</>
	);
}
