export interface Category {
  id: number;
  name: string;
  color: string;
  icon: string;
  type: "income" | "expense" | "both";
  patterns: string;
}

export interface CategoryInput {
  name: string;
  color: string;
  icon: string;
  type: "income" | "expense" | "both";
  patterns: string;
}
