import type { Todo } from "@/types/todo";
import { TodoItem } from "./TodoItem";

type Props = {
  todos: Todo[];
  onToggle: (id: number) => void;
  onDelete: (id: number) => void;
  onEdit: (id: number, title: string) => void;
};

export function TodoList({ todos, ...handlers }: Props) {
  if (todos.length === 0) {
    return <p className="mt-6 text-center text-slate-500">タスクがありません</p>;
  }
  return (
    <ul className="mt-4 space-y-2">
      {todos.map((todo) => (
        <TodoItem key={todo.id} todo={todo} {...handlers} />
      ))}
    </ul>
  );
}
