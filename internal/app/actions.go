package app

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

func normalized(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }
func parseActions(raw json.RawMessage, transcript []any) ([]any, error) {
	var list []map[string]any
	if err := json.Unmarshal(raw, &list); err != nil || list == nil || len(list) > 100 {
		return nil, errors.New("Некорректный список поручений.")
	}
	sources := map[string]string{}
	for _, item := range transcript {
		s := object(item)
		sources[text(s["id"])] = text(s["text"])
	}
	out := []any{}
	ids := map[string]bool{}
	for _, action := range list {
		if action == nil {
			return nil, errors.New("Некорректное поручение.")
		}
		for key := range action {
			if key != "id" && key != "text" && key != "owner" && key != "deadline" && key != "evidence" && key != "segment_id" && key != "status" {
				return nil, errors.New("Неизвестное поле поручения.")
			}
		}
		id := text(action["id"])
		if id == "" {
			id = randomID()
		}
		if len(id) > 100 || ids[id] {
			return nil, errors.New("Идентификаторы поручений должны быть уникальны.")
		}
		ids[id] = true
		action["id"] = id
		actionText := strings.TrimSpace(text(action["text"]))
		if actionText == "" || utf8.RuneCountInString(actionText) > 4000 {
			return nil, errors.New("Укажите текст поручения.")
		}
		action["text"] = actionText
		for _, field := range []string{"owner", "deadline", "evidence", "segment_id"} {
			value, exists := action[field]
			if !exists || value == nil {
				action[field] = nil
				continue
			}
			s, valid := value.(string)
			if !valid {
				return nil, errors.New("Некорректные поля поручения.")
			}
			s = strings.TrimSpace(s)
			if s == "" {
				action[field] = nil
			} else {
				action[field] = s
			}
		}
		if utf8.RuneCountInString(text(action["owner"])) > 200 || utf8.RuneCountInString(text(action["evidence"])) > 10000 || len(text(action["segment_id"])) > 100 {
			return nil, errors.New("Поля поручения слишком длинные.")
		}
		deadline := text(action["deadline"])
		if deadline != "" {
			if _, err := time.Parse("2006-01-02", deadline); err != nil {
				return nil, errors.New("Срок должен иметь формат YYYY-MM-DD.")
			}
		}
		status := text(action["status"])
		if status == "" {
			status = "pending"
		}
		if status != "pending" && status != "confirmed" {
			return nil, errors.New("Некорректный статус поручения.")
		}
		action["status"] = status
		segment := text(action["segment_id"])
		if segment != "" {
			source, exists := sources[segment]
			evidence := text(action["evidence"])
			if !exists || (evidence != "" && !strings.Contains(normalized(source), normalized(evidence))) {
				return nil, errors.New("Цитата должна соответствовать указанному фрагменту транскрипта.")
			}
		}
		out = append(out, action)
	}
	return out, nil
}
