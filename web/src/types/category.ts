export interface Category {
  id: number;
  name: string;
  color: string;
  icon: string;
  type: "income" | "expense" | "both";
  patterns: string;
  isTaxable: boolean;
}

export interface CategoryInput {
  name: string;
  color: string;
  icon: string;
  type: "income" | "expense" | "both";
  patterns: string;
  isTaxable: boolean;
}
