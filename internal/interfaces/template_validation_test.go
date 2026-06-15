package interfaces

import (
	"formatting-documents/internal/domain"
	"strings"
	"testing"
)

func TestValidateTemplateName(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantValue string
		wantError bool
	}{
		{name: "valid", value: "Учебный шаблон", wantValue: "Учебный шаблон"},
		{name: "trim spaces", value: "  Шаблон  ", wantValue: "Шаблон"},
		{name: "empty", value: "   ", wantError: true},
		{name: "sixty unicode characters", value: strings.Repeat("я", 60), wantValue: strings.Repeat("я", 60)},
		{name: "too long", value: strings.Repeat("я", 61), wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := domain.FormattingTemplate{Name: tt.value}
			err := validateTemplateName(&template)
			if (err != nil) != tt.wantError {
				t.Fatalf("validateTemplateName() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && template.Name != tt.wantValue {
				t.Fatalf("validateTemplateName() name = %q, want %q", template.Name, tt.wantValue)
			}
		})
	}
}
