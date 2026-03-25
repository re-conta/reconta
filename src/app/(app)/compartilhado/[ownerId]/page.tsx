import { SharedViewClient } from "@/components/sharing/shared-view-client";

interface Props {
	params: Promise<{ ownerId: string }>;
}

export default async function CompartilhadoPage({ params }: Props) {
	const { ownerId } = await params;
	return <SharedViewClient ownerId={ownerId} />;
}
