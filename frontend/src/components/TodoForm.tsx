"use client";

import { useState } from "react";

type Props = {
  onAdd: (title: string) => void;
};

export function TodoForm({ onAdd }: Props) {
  const [title, setTitle] = useState("");
  const [error, setError] = useState("");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault(); // ページ再読み込みを止める
    const trimmed = title.trim();
    if (trimmed.length < 1 || trimmed.length > 100) {
      setError("タイトルは1〜100文字で入力してください");
      return;
    }
    setError("");
    onAdd(trimmed);
    setTitle("");
  }

  return (
    <form onSubmit={handleSubmit}>
      <label htmlFor="title">新しいタスク</label>
      <input
        id="title"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="新しいタスクを入力"
      />
      <button type="submit">追加</button>
      {error && <p role="alert">{error}</p>}
    </form>
  );
}
