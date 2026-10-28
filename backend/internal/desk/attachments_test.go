package desk

import "testing"

func TestAttachmentTypes(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		valid      bool
	}{
		{"nota.txt", "evidencia de prueba", true},
		{"evidencia.pdf", "%PDF-1.7\n", true},
		{"imagen.png", "\x89PNG\r\n\x1a\n", true},
		{"imagen.jpg", "\xff\xd8\xff\xe0", true},
		{"script.html", "<!DOCTYPE html><html></html>", false},
		{"falso.txt", "<!DOCTYPE html><html></html>", false},
		{"falso.pdf", "texto", false},
		{"virus.exe", "MZ executable", false},
		{"../nota.txt", "texto", false},
		{`C:\nota.txt`, "texto", false},
		{"nota\r\n.txt", "texto", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, e := validateAttachment(tc.name, []byte(tc.body))
			if (e == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, e)
			}
		})
	}
}
