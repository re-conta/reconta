import { Header } from "@/components/layout/header";
import { TagsClient } from "@/components/tags/tags-client";

export default function SharedTagsPage() {
	return (
		<>
			<Header title="Tags" description="Visualizando tags compartilhadas" />
			<TagsClient />
		</>
	);
}
