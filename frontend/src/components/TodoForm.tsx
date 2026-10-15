"use client";

import { useState } from "react";

type Props = {
  onAdd: (title: string) => void | Promise<void>;
};

export function TodoForm({ onAdd }: Props) {
  const [title, setTitle] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = title.trim();
    if (trimmed.length < 1 || trimmed.length > 100) {
      setError("タイトルは1〜100文字で入力してください");
      return;
    }
    setError("");
    setSubmitting(true);
    try {
      await onAdd(trimmed);
      setTitle("");
    } finally {
      setSubmitting(false); // 成功でも失敗でも必ず戻す
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <label htmlFor="title" className="sr-only">新しいタスク</label>
      <div className="flex gap-2">
        <input
          id="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="新しいタスクを入力"
          aria-invalid={error ? true : undefined}
          aria-describedby={error ? "title-error" : undefined}
          className="flex-1 rounded-lg border border-slate-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <button
          type="submit"
          disabled={submitting}
          className="rounded-lg bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700 disabled:opacity-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        >
          {submitting ? "追加中…" : "追加"}
        </button>
      </div>
      <div className="mt-1 flex justify-between text-sm">
        {error ? (
          <p id="title-error" role="alert" className="text-red-600">{error}</p>
        ) : (
          <span />
        )}
        <span className={title.trim().length > 100 ? "text-red-600" : "text-slate-500"}>
          {title.trim().length}/100
        </span>
      </div>
    </form>
  );
}
