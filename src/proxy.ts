import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const PUBLIC_PATHS = ["/login", "/cadastro", "/api/auth", "/api/whatsapp/webhook"];

export function proxy(request: NextRequest) {
	const { pathname } = request.nextUrl;

	const isPublic = PUBLIC_PATHS.some((p) => pathname.startsWith(p));
	if (isPublic) return NextResponse.next();

	const sessionCookie =
		request.cookies.get("better-auth.session_token") ??
		request.cookies.get("__Secure-better-auth.session_token");

	if (!sessionCookie?.value) {
		const loginUrl = new URL("/login", request.url);
		loginUrl.searchParams.set("next", pathname);
		return NextResponse.redirect(loginUrl);
	}

	return NextResponse.next();
}

export const config = {
	matcher: [
		"/((?!_next/static|_next/image|favicon.ico|.*\\.png$|.*\\.svg$|.*\\.ico$|.*\\.jpg$|.*\\.jpeg$|.*\\.gif$|.*\\.webp$).*)",
	],
};
