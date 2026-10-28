package desk

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
)

const MaxAttachmentSize int64 = 10 << 20

type Attachment struct {
	ID         int64  `json:"id"`
	TicketID   int64  `json:"ticket_id"`
	UploaderID int64  `json:"uploader_id"`
	Name       string `json:"original_name"`
	MIME       string `json:"mime_type"`
	Size       int64  `json:"byte_size"`
	SHA256     string `json:"sha256"`
	Visibility string `json:"visibility"`
}

func attachmentRoot() string {
	if root := os.Getenv("ATTACHMENTS_DIR"); root != "" {
		return root
	}
	return filepath.Join("storage", "attachments")
}

func validAttachmentName(name string) bool {
	return name != "" && len(name) <= 180 && !strings.ContainsAny(name, `/\:`) && !strings.ContainsFunc(name, unicode.IsControl) && strings.TrimSpace(name) == name
}

func validateAttachment(name string, head []byte) (string, error) {
	if !validAttachmentName(name) {
		return "", bad("nombre de archivo inválido")
	}
	expected := map[string]string{".pdf": "application/pdf", ".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".txt": "text/plain"}[strings.ToLower(filepath.Ext(name))]
	detected := strings.Split(http.DetectContentType(head), ";")[0]
	if expected == "" || expected != detected {
		return "", &Error{415, "tipo o contenido de archivo no permitido"}
	}
	return detected, nil
}

func (s *Service) attachmentAccess(ctx context.Context, q queryer, u *auth.User, ticket int64, write bool, visibility string) error {
	t, err := object(ctx, q, `SELECT to_jsonb(t) FROM tickets t WHERE id=$1`, ticket)
	if err != nil {
		return err
	}
	internal := !u.ReadOnly() && s.allowed(ctx, q, u, int64(t["queue_id"].(float64)), "handle")
	if !s.readable(ctx, q, u, t) || visibility == "internal" && !internal {
		return forbidden()
	}
	if write {
		if u.ReadOnly() || !internal && int64(t["requester_id"].(float64)) != u.ID {
			return forbidden()
		}
		if t["state"] == "cerrado" || t["state"] == "rechazado" {
			return conflict("reabra el ticket antes de adjuntar archivos")
		}
	}
	return nil
}

func (s *Service) CheckAttachmentUpload(ctx context.Context, u *auth.User, ticket int64, visibility string) error {
	if visibility != "public" && visibility != "internal" {
		return bad("visibilidad requerida: public o internal")
	}
	return s.attachmentAccess(ctx, s.repo.pool, u, ticket, true, visibility)
}

func (s *Service) UploadAttachment(ctx context.Context, u *auth.User, ticket int64, name, visibility string, src io.Reader) (*Attachment, error) {
	if err := s.CheckAttachmentUpload(ctx, u, ticket, visibility); err != nil {
		return nil, err
	}
	if !validAttachmentName(name) {
		return nil, bad("nombre de archivo inválido")
	}
	root := attachmentRoot()
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, &Error{500, "almacenamiento no disponible"}
	}
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}
	key := hex.EncodeToString(keyBytes)
	path := filepath.Join(root, key)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return nil, &Error{500, "almacenamiento no disponible"}
	}
	keep := false
	defer func() {
		f.Close()
		if !keep {
			os.Remove(path)
		}
	}()
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(f, hash), io.LimitReader(src, MaxAttachmentSize+1))
	if err != nil {
		return nil, bad("no se pudo recibir el archivo")
	}
	if size > MaxAttachmentSize {
		return nil, &Error{413, "el archivo supera 10 MiB"}
	}
	if size == 0 {
		return nil, bad("archivo vacío")
	}
	head := make([]byte, 512)
	n, err := f.ReadAt(head, 0)
	if err != nil && err != io.EOF {
		return nil, &Error{500, "almacenamiento no disponible"}
	}
	mimeType, err := validateAttachment(name, head[:n])
	if err != nil {
		return nil, err
	}
	if err = f.Sync(); err != nil {
		return nil, &Error{500, "almacenamiento no disponible"}
	}
	a := &Attachment{TicketID: ticket, UploaderID: u.ID, Name: name, MIME: mimeType, Size: size, SHA256: hex.EncodeToString(hash.Sum(nil)), Visibility: visibility}
	err = s.transaction(ctx, u, func(tx pgx.Tx) error {
		// Serialize attachment metadata with ticket state and queue changes.
		var locked int64
		if e := tx.QueryRow(ctx, `SELECT id FROM tickets WHERE id=$1 FOR UPDATE`, ticket).Scan(&locked); e != nil {
			return e
		}
		if e := tx.QueryRow(ctx, `SELECT q.id FROM queues q JOIN tickets t ON t.queue_id=q.id WHERE t.id=$1 FOR SHARE OF q`, ticket).Scan(&locked); e != nil {
			return e
		}
		if e := s.attachmentAccess(ctx, tx, u, ticket, true, visibility); e != nil {
			return e
		}
		if e := tx.QueryRow(ctx, `INSERT INTO ticket_attachments(ticket_id,uploader_id,original_name,storage_key,mime_type,byte_size,sha256,visibility) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, ticket, u.ID, name, key, mimeType, size, a.SHA256, visibility).Scan(&a.ID); e != nil {
			return e
		}
		// Internal attachment metadata must not leak through public ticket history.
		return s.audit(ctx, tx, ticket, u.ID, "attachment_uploaded", nil, map[string]any{"attachment_id": a.ID}, "")
	})
	if err != nil {
		return nil, err
	}
	keep = true
	return a, nil
}

func (s *Service) Attachments(ctx context.Context, u *auth.User, ticket int64) ([]map[string]any, error) {
	if err := s.attachmentAccess(ctx, s.repo.pool, u, ticket, false, "public"); err != nil {
		return nil, err
	}
	internal := s.attachmentAccess(ctx, s.repo.pool, u, ticket, false, "internal") == nil
	return objects(ctx, s.repo.pool, `SELECT to_jsonb(a)-'storage_key' FROM ticket_attachments a WHERE ticket_id=$1 AND ($2 OR visibility='public') ORDER BY id`, ticket, internal)
}

func (s *Service) DownloadAttachment(ctx context.Context, u *auth.User, ticket, id int64) (*Attachment, *os.File, error) {
	a := &Attachment{}
	var key string
	err := s.repo.pool.QueryRow(ctx, `SELECT id,ticket_id,uploader_id,original_name,storage_key,mime_type,byte_size,sha256,visibility FROM ticket_attachments WHERE ticket_id=$1 AND id=$2`, ticket, id).Scan(&a.ID, &a.TicketID, &a.UploaderID, &a.Name, &key, &a.MIME, &a.Size, &a.SHA256, &a.Visibility)
	if err == pgx.ErrNoRows {
		return nil, nil, &Error{404, "adjunto no encontrado"}
	}
	if err != nil {
		return nil, nil, err
	}
	if err = s.attachmentAccess(ctx, s.repo.pool, u, ticket, false, a.Visibility); err != nil {
		return nil, nil, err
	}
	decoded, e := hex.DecodeString(key)
	if e != nil || len(decoded) != 32 || len(key) != 64 {
		return nil, nil, &Error{500, "almacenamiento inválido"}
	}
	path := filepath.Join(attachmentRoot(), key)
	stat, e := os.Lstat(path)
	if e != nil || !stat.Mode().IsRegular() {
		return nil, nil, &Error{404, "archivo no disponible"}
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, nil, &Error{404, "archivo no disponible"}
	}
	actual, e := f.Stat()
	if e != nil || !os.SameFile(stat, actual) || actual.Size() != a.Size {
		f.Close()
		return nil, nil, &Error{500, "integridad del archivo inválida"}
	}
	hash := sha256.New()
	_, e = io.Copy(hash, io.LimitReader(f, MaxAttachmentSize+1))
	if e != nil || hex.EncodeToString(hash.Sum(nil)) != a.SHA256 {
		f.Close()
		return nil, nil, &Error{500, "integridad del archivo inválida"}
	}
	if _, e = f.Seek(0, 0); e != nil {
		f.Close()
		return nil, nil, &Error{500, "archivo no disponible"}
	}
	return a, f, nil
}
