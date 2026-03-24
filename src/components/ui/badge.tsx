import type * as React from "react";
import { cn } from "@/lib/utils";

interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
	variant?: "default" | "success" | "danger" | "warning" | "info";
}

const variants = {
	default: "bg-zinc-800 text-zinc-300",
	success: "bg-emerald-900/50 text-emerald-400",
	danger: "bg-red-900/50 text-red-400",
	warning: "bg-amber-900/50 text-amber-400",
	info: "bg-blue-900/50 text-blue-400",
};

export function Badge({
	className,
	variant = "default",
	...props
}: BadgeProps) {
	return (
		<span
			className={cn(
				"inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium",
				variants[variant],
				className,
			)}
			{...props}
		/>
	);
}
