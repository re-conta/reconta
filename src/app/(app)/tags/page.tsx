import { Header } from "@/components/layout/header";
import { TagsClient } from "@/components/tags/tags-client";

export default function TagsPage() {
	return (
		<>
			<Header
				title="Tags"
				description="Gerencie as tags de receitas e despesas"
			/>
			<TagsClient />
		</>
	);
}
