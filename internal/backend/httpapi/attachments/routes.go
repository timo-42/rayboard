package attachments

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	attachmentservice "github.com/timo-42/rayboard/internal/backend/attachments"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

const attachmentsTag = "Attachments"

type routes struct {
	provider Provider
}

func Register(api huma.API, provider Provider) {
	route := routes{provider: provider}

	huma.Register(api, shared.Operation(http.MethodGet, "/api/tickets/{ticket_id}/attachments", attachmentsTag, "List ticket attachments"), route.list)
	route.registerUpload(api)
	huma.Register(api, downloadAttachmentOperation(), route.download)
	huma.Register(api, shared.Operation(http.MethodDelete, "/api/attachments/{attachment_id}", attachmentsTag, "Delete attachment"), route.delete)
}

func (r routes) list(ctx context.Context, input *listAttachmentsInput) (*listAttachmentsOutput, error) {
	ctx, principal, _, err := r.provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}

	items, err := r.provider.Attachments.List(ctx, principal, input.TicketID)
	if err != nil {
		return nil, shared.AttachmentError(err)
	}
	return &listAttachmentsOutput{Body: shared.ItemList[AttachmentResource]{Items: attachmentResources(items)}}, nil
}

func (r routes) registerUpload(api huma.API) {
	op := uploadAttachmentOperation(api)
	api.OpenAPI().AddOperation(&op)
	api.Adapter().Handle(&op, api.Middlewares().Handler(op.Middlewares.Handler(func(ctx huma.Context) {
		r.upload(api, ctx)
	})))
}

func (r routes) upload(api huma.API, ctx huma.Context) {
	authInput := authInputFromContext(ctx)
	requestContext, principal, _, err := r.provider.Authenticator.Authenticate(ctx.Context(), authInput, true)
	if err != nil {
		writeError(api, ctx, err)
		return
	}

	upload, err := readUpload(ctx)
	if err != nil {
		writeError(api, ctx, err)
		return
	}

	meta, err := r.provider.Attachments.Upload(requestContext, principal, attachmentservice.UploadInput{
		TicketID:    ctx.Param("ticket_id"),
		FileName:    upload.fileName,
		ContentType: upload.contentType,
		Data:        upload.data,
	})
	if err != nil {
		writeError(api, ctx, shared.AttachmentError(err))
		return
	}
	writeOutput(api, ctx, http.StatusCreated, attachmentResource(meta))
}

func (r routes) download(ctx context.Context, input *downloadAttachmentInput) (*downloadAttachmentOutput, error) {
	ctx, principal, _, err := r.provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}

	file, err := r.provider.Attachments.Download(ctx, principal, input.AttachmentID)
	if err != nil {
		return nil, shared.AttachmentError(err)
	}
	return &downloadAttachmentOutput{
		ContentType:        file.ContentType,
		ContentDisposition: mime.FormatMediaType("attachment", map[string]string{"filename": file.FileName}),
		Body:               file.Data,
	}, nil
}

func (r routes) delete(ctx context.Context, input *deleteAttachmentInput) (*shared.EmptyOutput, error) {
	ctx, principal, _, err := r.provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}

	if err := r.provider.Attachments.Delete(ctx, principal, input.AttachmentID); err != nil {
		return nil, shared.AttachmentError(err)
	}
	return &shared.EmptyOutput{}, nil
}

type uploadedFile struct {
	fileName    string
	contentType string
	data        []byte
}

func readUpload(ctx huma.Context) (uploadedFile, error) {
	mediaType, params, err := mime.ParseMediaType(ctx.Header("Content-Type"))
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") || params["boundary"] == "" {
		return uploadedFile{}, missingFileError()
	}

	reader := multipart.NewReader(io.LimitReader(ctx.BodyReader(), uploadReadLimitBytes), params["boundary"])
	form, err := reader.ReadForm(1 << 20)
	if err != nil {
		return uploadedFile{}, missingFileError()
	}
	defer form.RemoveAll()

	files := form.File["file"]
	if len(files) == 0 {
		return uploadedFile{}, missingFileError()
	}

	file, err := files[0].Open()
	if err != nil {
		return uploadedFile{}, unreadableFileError()
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, attachmentservice.MaxAttachmentSizeBytes+1))
	if err != nil {
		return uploadedFile{}, unreadableFileError()
	}

	contentType := files[0].Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	return uploadedFile{
		fileName:    files[0].Filename,
		contentType: contentType,
		data:        data,
	}, nil
}

func missingFileError() error {
	return shared.AttachmentError(&attachmentservice.ValidationError{
		Message: "Multipart form with file is required",
		Fields:  map[string]string{"file": "Required"},
	})
}

func unreadableFileError() error {
	return shared.AttachmentError(&attachmentservice.ValidationError{
		Message: "Could not read uploaded file",
		Fields:  map[string]string{"file": "Unreadable"},
	})
}

func authInputFromContext(ctx huma.Context) shared.AuthInput {
	return shared.AuthInput{
		Authorization: ctx.Header("Authorization"),
		CookieHeader:  ctx.Header("Cookie"),
		SessionCookie: cookieValueFromContext(ctx, "rayboard_session"),
		CSRFCookie:    cookieValueFromContext(ctx, shared.CSRFCookieName),
		CSRFToken:     ctx.Header("X-CSRF-Token"),
	}
}

func cookieValueFromContext(ctx huma.Context, name string) string {
	cookie, err := huma.ReadCookie(ctx, name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func writeError(api huma.API, ctx huma.Context, err error) {
	var headers huma.HeadersError
	if errors.As(err, &headers) {
		for name, values := range headers.GetHeaders() {
			for _, value := range values {
				ctx.AppendHeader(name, value)
			}
		}
	}

	var status huma.StatusError
	if errors.As(err, &status) {
		_ = huma.WriteErr(api, ctx, status.GetStatus(), status.Error())
		return
	}
	_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "Request failed")
}

func writeOutput(api huma.API, ctx huma.Context, status int, body any) {
	contentType, err := api.Negotiate(ctx.Header("Accept"))
	if err != nil {
		writeError(api, ctx, huma.Error406NotAcceptable("unable to marshal response"))
		return
	}

	transformed, err := api.Transform(ctx, strconv.Itoa(status), body)
	if err != nil {
		writeError(api, ctx, huma.Error500InternalServerError("Request failed"))
		return
	}

	ctx.SetHeader("Content-Type", contentType)
	ctx.SetStatus(status)
	if err := api.Marshal(ctx.BodyWriter(), contentType, transformed); err != nil {
		panic(err)
	}
}

func reflectType[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
