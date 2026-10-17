"use client";

import { useState } from "react";
import type { Todo } from "@/types/todo";

type Props = {
  todo: Todo;
  onToggle: (id: number) => void;
  onDelete: (id: number) => void;
  onEdit: (id: number, title: string) => void;
};

const MAX_LENGTH = 100;

const buttonClass =
  "rounded px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500";

export function TodoItem({ todo, onToggle, onDelete, onEdit }: Props) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(todo.title);

  const draftLength = [...draft.trim()].length;
  const draftValid = draftLength >= 1 && draftLength <= MAX_LENGTH;

  function save(e: React.FormEvent) {
    e.preventDefault();
    if (!draftValid) return;
    onEdit(todo.id, draft.trim());
    setEditing(false);
  }

  function cancel() {
    setDraft(todo.title);
    setEditing(false);
  }

  return (
    <li className="flex flex-wrap items-center gap-3 rounded-lg border border-slate-200 bg-white p-3">
      <input
        type="checkbox"
        className="h-5 w-5 accent-indigo-600"
        checked={todo.done}
        onChange={() => onToggle(todo.id)}
        aria-label={`${todo.title} を完了にする`}
      />
      {editing ? (
        <form onSubmit={save} className="flex flex-1 flex-wrap items-center gap-2">
          <input
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            aria-label="タイトルを編集"
            aria-invalid={draftValid ? undefined : true}
            autoFocus
            className="min-w-0 flex-1 rounded border border-slate-300 px-2 py-1 text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
          <button
            type="submit"
            disabled={!draftValid}
            className={`${buttonClass} bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50`}
          >
            保存
          </button>
          <button
            type="button"
            onClick={cancel}
            className={`${buttonClass} text-slate-600 hover:bg-slate-100`}
          >
            キャンセル
          </button>
        </form>
      ) : (
        <>
          <span
            className={`min-w-0 flex-1 break-words ${
              todo.done ? "text-slate-400 line-through" : "text-slate-900"
            }`}
          >
            {todo.title}
          </span>
          <button
            type="button"
            onClick={() => setEditing(true)}
            aria-label={`${todo.title} を編集`}
            className={`${buttonClass} text-slate-600 hover:bg-slate-100`}
          >
            編集
          </button>
          <button
            type="button"
            onClick={() => onDelete(todo.id)}
            aria-label={`${todo.title} を削除`}
            className={`${buttonClass} text-red-600 hover:bg-red-50`}
          >
            削除
          </button>
        </>
      )}
    </li>
  );
}
