import { SharedOwnerProvider } from "@/components/layout/shared-owner-context";

interface Props {
	params: Promise<{ ownerId: string }>;
	children: React.ReactNode;
}

export default async function SharedLayout({ params, children }: Props) {
	const { ownerId } = await params;
	return (
		<SharedOwnerProvider ownerId={ownerId}>{children}</SharedOwnerProvider>
	);
}
