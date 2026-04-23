package source

import (
	"fmt"
	"strings"
)

func extractTextField(data map[string]any) (string, error) {
	candidates := []string{"content", "text", "body", "message"}
	for _, key := range candidates {
		value, ok := data[key]
		if !ok {
			continue
		}
		text, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("field %q must be a string", key)
		}
		text = strings.TrimSpace(text)
		if text != "" {
			return text, nil
		}
	}

	for _, value := range data {
		if text, ok := value.(string); ok {
			text = strings.TrimSpace(text)
			if text != "" {
				return text, nil
			}
		}
	}

	return "", fmt.Errorf("no text field found (expected one of: content,text,body,message)")
}
