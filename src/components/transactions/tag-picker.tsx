"use client";

import * as DropdownMenuPrimitive from "@radix-ui/react-dropdown-menu";
import { Plus, Tag as TagIcon, X } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";

export interface Tag {
	id: number;
	name: string;
	color: string;
}

const TAG_COLORS = [
	"#ef4444",
	"#f97316",
	"#f59e0b",
	"#84cc16",
	"#10b981",
	"#06b6d4",
	"#3b82f6",
	"#8b5cf6",
	"#ec4899",
	"#6b7280",
];

function randomColor() {
	return TAG_COLORS[Math.floor(Math.random() * TAG_COLORS.length)];
}

interface Props {
	tags: Tag[];
	selectedIds: number[];
	onChange: (ids: number[]) => void;
	onTagCreated: (tag: Tag) => void;
	className?: string;
}

export function TagPicker({
	tags,
	selectedIds,
	onChange,
	onTagCreated,
	className,
}: Props) {
	const [open, setOpen] = useState(false);
	const [newTagName, setNewTagName] = useState("");
	const [creating, setCreating] = useState(false);

	const selected = tags.filter((t) => selectedIds.includes(t.id));

	function toggle(id: number) {
		if (selectedIds.includes(id)) {
			onChange(selectedIds.filter((i) => i !== id));
		} else {
			onChange([...selectedIds, id]);
		}
	}

	async function createTag(e: React.FormEvent) {
		e.preventDefault();
		const name = newTagName.trim();
		if (!name) return;
		setCreating(true);
		const res = await fetch("/api/tags", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ name, color: randomColor() }),
		});
		const tag = await res.json();
		setCreating(false);
		setNewTagName("");
		onTagCreated(tag);
		onChange([...selectedIds, tag.id]);
	}

	return (
		<DropdownMenuPrimitive.Root open={open} onOpenChange={setOpen}>
			<DropdownMenuPrimitive.Trigger asChild>
				<button
					type="button"
					className={cn(
						"flex items-center gap-1.5 rounded-lg border border-zinc-700 bg-zinc-800 px-2.5 h-9 text-sm text-zinc-300 hover:border-zinc-600 cursor-pointer",
						className,
					)}
				>
					<TagIcon className="h-3.5 w-3.5 text-zinc-500" />
					{selected.length === 0 ? (
						<span className="text-zinc-500">Tags</span>
					) : (
						<span className="flex flex-wrap gap-1 max-w-48">
							{selected.map((t) => (
								<span
									key={t.id}
									className="inline-flex items-center px-1.5 py-0.5 rounded-full text-xs font-medium truncate"
									style={{ backgroundColor: `${t.color}22`, color: t.color }}
								>
									{t.name}
								</span>
							))}
						</span>
					)}
				</button>
			</DropdownMenuPrimitive.Trigger>
			<DropdownMenuPrimitive.Portal>
				<DropdownMenuPrimitive.Content
					sideOffset={4}
					align="start"
					className="z-50 w-56 rounded-lg border border-zinc-700 bg-zinc-900 p-2 shadow-xl"
					onCloseAutoFocus={(e) => e.preventDefault()}
				>
					<div className="max-h-48 overflow-y-auto space-y-0.5">
						{tags.length === 0 && (
							<p className="px-2 py-1.5 text-xs text-zinc-500">
								Nenhuma tag criada ainda.
							</p>
						)}
						{tags.map((tag) => {
							const isSelected = selectedIds.includes(tag.id);
							return (
								<button
									key={tag.id}
									type="button"
									onClick={() => toggle(tag.id)}
									className={cn(
										"flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors cursor-pointer",
										isSelected
											? "bg-zinc-800 text-zinc-100"
											: "text-zinc-300 hover:bg-zinc-800/60",
									)}
								>
									<span
										className="h-3 w-3 rounded-full shrink-0"
										style={{ backgroundColor: tag.color }}
									/>
									<span className="flex-1 truncate text-left">{tag.name}</span>
									{isSelected && <X className="h-3 w-3 text-zinc-500" />}
								</button>
							);
						})}
					</div>
					<form
						onSubmit={createTag}
						className="mt-2 flex items-center gap-1 border-t border-zinc-800 pt-2"
					>
						<input
							type="text"
							value={newTagName}
							onChange={(e) => setNewTagName(e.target.value)}
							placeholder="Nova tag..."
							className="flex-1 h-8 rounded-md border border-zinc-700 bg-zinc-800 px-2 text-xs text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
						/>
						<button
							type="submit"
							disabled={creating || !newTagName.trim()}
							className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-indigo-600 text-white disabled:opacity-40 cursor-pointer"
						>
							<Plus className="h-3.5 w-3.5" />
						</button>
					</form>
				</DropdownMenuPrimitive.Content>
			</DropdownMenuPrimitive.Portal>
		</DropdownMenuPrimitive.Root>
	);
}
