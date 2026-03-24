import { headers } from "next/headers";
import { NextResponse } from "next/server";
import { auth } from "@/lib/auth";

export async function getSession() {
	return auth.api.getSession({ headers: await headers() });
}

export async function requireSession() {
	const session = await getSession();
	if (!session) {
		return {
			session: null,
			userId: null as never,
			unauthorized: NextResponse.json(
				{ error: "Não autorizado" },
				{ status: 401 },
			),
		};
	}
	return { session, userId: session.user.id, unauthorized: null };
}
