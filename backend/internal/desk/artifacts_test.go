package desk

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestSourceIntegrity(t *testing.T) {
	root := "."
	if _, e := os.Stat("data/CatalogoServicios.xlsx"); e != nil {
		root = "../../.."
	}
	for path, want := range map[string]string{
		"data/CatalogoServicios.xlsx":                      "de3b478a5faeeeaebce1aa7726e0e3321188a68e41bbb656e1d17b0c5b74dcf0",
		"data/fuentes-service-desk/CMDB.xlsx":              "d182796bb40b492be2cb87e2c5fe19a8f910f2e754183dc48b5e89bd360464dc",
		"data/fuentes-service-desk/Tarea_Service_Desk.pdf": "0a36d768f0ef305334b4f56ad65e8dda1be0410f6f6e636254ba0090ec29396e",
	} {
		b, e := os.ReadFile(filepath.Join(root, path))
		if e != nil {
			t.Fatal(e)
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatalf("original modified: %s", path)
		}
	}
	t.Log("PASS original hashes: catalogue, CMDB and assignment")
	if _, e := os.Stat(filepath.Join(root, "frontend/src")); e != nil {
		t.Log("source tree not included in runtime, secret scan runs in image builder")
		return
	}
	patterns := []*regexp.Regexp{regexp.MustCompile(`-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----`), regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{36,}`), regexp.MustCompile(`sk-proj-[A-Za-z0-9_-]{48,}`)}
	e := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "dist", "outputs", "tmp", "storage":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(d.Name(), ".env") && d.Name() != ".env.example" {
			return nil
		}
		switch filepath.Ext(path) {
		case ".go", ".jsx", ".js", ".mjs", ".md", ".txt", ".yaml", ".yml", ".sh", ".json":
		default:
			return nil
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		for _, pattern := range patterns {
			if pattern.Match(b) {
				t.Errorf("possible secret in %s, value withheld", path)
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log("PASS heuristic scan of source tree for private keys and provider tokens, not a complete secrets audit")
}
