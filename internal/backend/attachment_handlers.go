package backend

import (
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/timo-42/rayboard/internal/backend/attachments"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpjson"
)

type attachmentRoute struct {
	attachments *attachments.Service
}

func registerAttachmentRoutes(mux *http.ServeMux, authService *auth.Service, attachmentService *attachments.Service) {
	authRoute := authRoute{auth: authService}
	route := attachmentRoute{attachments: attachmentService}

	mux.HandleFunc("GET /api/tickets/{ticket_id}/attachments", authRoute.requireAuth(route.list))
	mux.HandleFunc("POST /api/tickets/{ticket_id}/attachments", authRoute.requireAuth(route.upload))
	mux.HandleFunc("GET /api/attachments/{attachment_id}/download", authRoute.requireAuth(route.download))
	mux.HandleFunc("DELETE /api/attachments/{attachment_id}", authRoute.requireAuth(route.delete))
}

func (route attachmentRoute) list(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	items, err := route.attachments.List(r.Context(), principal, r.PathValue("ticket_id"))
	if err != nil {
		writeAttachmentError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": items})
}

func (route attachmentRoute) upload(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	r.Body = http.MaxBytesReader(w, r.Body, attachments.MaxAttachmentSizeBytes+1<<20)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", "Multipart form with file is required", map[string]string{"file": "Required"})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", "Multipart form with file is required", map[string]string{"file": "Required"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, attachments.MaxAttachmentSizeBytes+1))
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", "Could not read uploaded file", map[string]string{"file": "Unreadable"})
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	meta, err := route.attachments.Upload(r.Context(), principal, attachments.UploadInput{
		TicketID:    r.PathValue("ticket_id"),
		FileName:    header.Filename,
		ContentType: contentType,
		Data:        data,
	})
	if err != nil {
		writeAttachmentError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, meta)
}

func (route attachmentRoute) download(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	file, err := route.attachments.Download(r.Context(), principal, r.PathValue("attachment_id"))
	if err != nil {
		writeAttachmentError(w, err)
		return
	}
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": file.FileName}))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(file.Data)
}

func (route attachmentRoute) delete(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if err := route.attachments.Delete(r.Context(), principal, r.PathValue("attachment_id")); err != nil {
		writeAttachmentError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeAttachmentError(w http.ResponseWriter, err error) {
	var validation *attachments.ValidationError
	switch {
	case errors.As(err, &validation):
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", validation.Message, validation.Fields)
	case errors.Is(err, attachments.ErrValidation):
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", "Validation failed", nil)
	case errors.Is(err, attachments.ErrTooLarge):
		httpjson.Error(w, http.StatusRequestEntityTooLarge, "validation_failed", "Attachment is too large", map[string]string{"file": "Maximum size is 10 MiB"})
	case errors.Is(err, attachments.ErrNotFound):
		httpjson.Error(w, http.StatusNotFound, "not_found", "Resource was not found", nil)
	case errors.Is(err, authz.ErrForbidden):
		httpjson.Error(w, http.StatusForbidden, "forbidden", "Permission denied", nil)
	default:
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "Request failed", nil)
	}
}
