import type { TodoStatus } from "@/lib/api";

const OPTIONS: { value: TodoStatus; label: string }[] = [
  { value: "all", label: "全て" },
  { value: "active", label: "未完了" },
  { value: "done", label: "完了" },
];

type Props = {
  value: TodoStatus;
  onChange: (value: TodoStatus) => void;
};

export function StatusFilter({ value, onChange }: Props) {
  return (
    <div role="group" aria-label="表示の絞り込み" className="mt-4 flex gap-2">
      {OPTIONS.map((o) => (
        <button
          key={o.value}
          type="button"
          aria-pressed={value === o.value}
          onClick={() => onChange(o.value)}
          className={`rounded-full px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 ${
            value === o.value
              ? "bg-indigo-600 text-white"
              : "bg-slate-100 text-slate-700 hover:bg-slate-200"
          }`}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}
