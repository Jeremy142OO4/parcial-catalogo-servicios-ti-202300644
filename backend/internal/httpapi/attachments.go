package httpapi

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
)

func (s *Server) attachmentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/desk/tickets/{id}/attachments", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		v, e := s.desk.Attachments(r.Context(), u, idOf(r))
		deskResult(w, v, e)
	}))
	mux.HandleFunc("POST /api/desk/tickets/{id}/attachments", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		// Reject unauthorized senders before parsing their body.
		if e := s.desk.CheckAttachmentUpload(r.Context(), u, idOf(r), "public"); e != nil {
			deskResult(w, nil, e)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 11<<20)
		e := r.ParseMultipartForm(1 << 20)
		if r.MultipartForm != nil {
			defer r.MultipartForm.RemoveAll()
		}
		if e != nil {
			var limit *http.MaxBytesError
			if errors.As(e, &limit) {
				writeError(w, 413, "solicitud supera 11 MiB")
			} else {
				writeError(w, 400, "multipart inválido")
			}
			return
		}
		files := r.MultipartForm.File["file"]
		if len(files) != 1 || len(r.MultipartForm.File) != 1 || len(r.MultipartForm.Value) != 1 || len(r.MultipartForm.Value["visibility"]) != 1 {
			writeError(w, 400, "envíe un file y una visibility")
			return
		}
		_, params, e := mime.ParseMediaType(files[0].Header.Get("Content-Disposition"))
		if e != nil {
			writeError(w, 400, "nombre inválido")
			return
		}
		file, e := files[0].Open()
		if e != nil {
			writeError(w, 400, "archivo inválido")
			return
		}
		defer file.Close()
		v, e := s.desk.UploadAttachment(r.Context(), u, idOf(r), params["filename"], r.FormValue("visibility"), file)
		if e != nil {
			deskResult(w, nil, e)
			return
		}
		writeJSON(w, 201, v)
	}))
	mux.HandleFunc("GET /api/desk/tickets/{id}/attachments/{attachment}/download", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		id, e := strconv.ParseInt(r.PathValue("attachment"), 10, 64)
		if e != nil || id < 1 {
			writeError(w, 400, "id de adjunto inválido")
			return
		}
		a, f, e := s.desk.DownloadAttachment(r.Context(), u, idOf(r), id)
		if e != nil {
			deskResult(w, nil, e)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", a.MIME)
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": a.Name}))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("Content-Length", strconv.FormatInt(a.Size, 10))
		io.Copy(w, f)
	}))
}
