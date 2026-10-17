"use client";

import { useState } from "react";

const MAX_LENGTH = 100;

type Props = {
  /** 追加に成功したら true を返す。失敗時は入力を残す。 */
  onAdd: (title: string) => Promise<boolean>;
};

export function TodoForm({ onAdd }: Props) {
  const [title, setTitle] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const length = [...title.trim()].length; // 絵文字なども1文字として数える

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault(); // ページ再読み込みを止める
    if (length < 1 || length > MAX_LENGTH) {
      setError(`タイトルは1〜${MAX_LENGTH}文字で入力してください`);
      return;
    }
    setError("");
    setSubmitting(true);
    try {
      if (await onAdd(title.trim())) setTitle("");
    } finally {
      setSubmitting(false); // 成功でも失敗でも必ず戻す
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <label htmlFor="title" className="sr-only">
        新しいタスク
      </label>
      <div className="flex gap-2">
        <input
          id="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="新しいタスクを入力"
          aria-invalid={error ? true : undefined}
          aria-describedby={error ? "title-error" : undefined}
          className="flex-1 rounded-lg border border-slate-300 bg-white px-3 py-2 text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <button
          type="submit"
          disabled={submitting}
          className="rounded-lg bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:opacity-50"
        >
          {submitting ? "追加中…" : "追加"}
        </button>
      </div>
      <div className="mt-1 flex justify-between text-sm">
        {error ? (
          <p id="title-error" role="alert" className="text-red-600">
            {error}
          </p>
        ) : (
          <span />
        )}
        <span className={length > MAX_LENGTH ? "text-red-600" : "text-slate-500"}>
          {length}/{MAX_LENGTH}
        </span>
      </div>
    </form>
  );
}
