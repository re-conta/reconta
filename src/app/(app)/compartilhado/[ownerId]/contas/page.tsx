import { Header } from "@/components/layout/header";
import { ContasClient } from "@/components/bills/contas-client";
import { getCurrentMonth } from "@/lib/utils";

export default function SharedContasPage() {
	const { month, year } = getCurrentMonth();
	return (
		<>
			<Header
				title="Contas Fixas"
				description="Visualizando contas fixas compartilhadas"
			/>
			<ContasClient initialMonth={month} initialYear={year} />
		</>
	);
}
