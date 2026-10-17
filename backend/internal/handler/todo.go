// Package handler は HTTP のリクエスト解析とレスポンス生成を担当する。
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"todoapp/backend/internal/model"
	"todoapp/backend/internal/store"
)

type TodoHandler struct {
	store store.TodoStore // 具体型ではなくインターフェース
}

func NewTodoHandler(s store.TodoStore) *TodoHandler {
	return &TodoHandler{store: s}
}

// Register はルートを登録する。protect で包んだルートはログイン必須になる。
func (h *TodoHandler) Register(mux *http.ServeMux, protect func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /todos", protect(h.list))
	mux.HandleFunc("POST /todos", protect(h.create))
	mux.HandleFunc("GET /todos/{id}", protect(h.get))
	mux.HandleFunc("PATCH /todos/{id}", protect(h.update))
	mux.HandleFunc("DELETE /todos/{id}", protect(h.delete))
}

// ---------- 共通ヘルパー ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status) // ボディを書く前に呼ぶ
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

// decodeJSON は本文サイズを1MBに制限し、未知のフィールドを拒否する。
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func parseID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

// writeStoreError は store のエラーを HTTP に変換する。詳細はログにだけ出す。
func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "タスクが見つかりません")
		return
	}
	log.Printf("internal error: %v", err)
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバーエラーが発生しました")
}

// ---------- ハンドラー ----------

func (h *TodoHandler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *TodoHandler) list(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	switch status {
	case "":
		status = "all"
	case "all", "active", "done":
	default:
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "status は all / active / done のいずれかです")
		return
	}
	todos, err := h.store.List(r.Context(), userID(r), status)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todos)
}

func (h *TodoHandler) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの形式が正しくありません")
		return
	}
	title, err := model.NormalizeTitle(req.Title)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	todo, err := h.store.Create(r.Context(), userID(r), title)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/todos/%d", todo.ID))
	writeJSON(w, http.StatusCreated, todo)
}

func (h *TodoHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "IDが正しくありません")
		return
	}
	todo, err := h.store.Get(r.Context(), userID(r), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "IDが正しくありません")
		return
	}
	// ポインタ型: 「送られなかった」と「false / 空文字が送られた」を区別できる
	var req struct {
		Title *string `json:"title"`
		Done  *bool   `json:"done"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの形式が正しくありません")
		return
	}
	if req.Title == nil && req.Done == nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "title または done を指定してください")
		return
	}
	in := store.UpdateInput{Done: req.Done}
	if req.Title != nil {
		title, err := model.NormalizeTitle(*req.Title)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		in.Title = &title
	}
	todo, err := h.store.Update(r.Context(), userID(r), id, in)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "IDが正しくありません")
		return
	}
	if err := h.store.Delete(r.Context(), userID(r), id); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
