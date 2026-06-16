package attachments

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	attachmentservice "github.com/timo-42/rayboard/internal/backend/attachments"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

const uploadReadLimitBytes = attachmentservice.MaxAttachmentSizeBytes + 1<<20

type listAttachmentsInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id" doc:"Ticket ID."`
}

type listAttachmentsOutput struct {
	Body shared.ListResource[AttachmentResource]
}

type uploadAttachmentInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id" doc:"Ticket ID."`
}

type uploadAttachmentOutput struct {
	Status int `status:"201"`
	Body   AttachmentResource
}

type downloadAttachmentInput struct {
	shared.AuthInput
	AttachmentID string `path:"attachment_id" doc:"Attachment ID."`
}

type downloadAttachmentOutput struct {
	ContentType        string `header:"Content-Type" doc:"Attachment media type."`
	ContentDisposition string `header:"Content-Disposition" doc:"Attachment download filename."`
	Body               []byte
}

type deleteAttachmentInput struct {
	shared.AuthInput
	AttachmentID string `path:"attachment_id" doc:"Attachment ID."`
}

type AttachmentMetadata struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	CreatedAt time.Time `json:"created_at"`
}

type AttachmentSpec struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

type AttachmentStatus struct {
	SizeBytes  int64  `json:"size_bytes"`
	UploaderID string `json:"uploader_id,omitempty"`
}

type AttachmentResource = shared.Resource[AttachmentMetadata, AttachmentSpec, AttachmentStatus]

func attachmentResource(meta attachmentservice.Metadata) AttachmentResource {
	return AttachmentResource{
		Metadata: AttachmentMetadata{
			ID:        meta.ID,
			TicketID:  meta.TicketID,
			CreatedAt: meta.CreatedAt,
		},
		Spec: AttachmentSpec{
			FileName:    meta.FileName,
			ContentType: meta.ContentType,
		},
		Status: AttachmentStatus{
			SizeBytes:  meta.SizeBytes,
			UploaderID: meta.UploaderID,
		},
	}
}

func attachmentResources(items []attachmentservice.Metadata) []AttachmentResource {
	resources := make([]AttachmentResource, 0, len(items))
	for _, item := range items {
		resources = append(resources, attachmentResource(item))
	}
	return resources
}

func uploadAttachmentRequestBody() *huma.RequestBody {
	return &huma.RequestBody{
		Required: true,
		Content: map[string]*huma.MediaType{
			"multipart/form-data": {
				Schema: &huma.Schema{
					Type: "object",
					Properties: map[string]*huma.Schema{
						"file": {
							Type:            "string",
							Format:          "binary",
							ContentEncoding: "binary",
							Description:     "Attachment file.",
						},
					},
					Required: []string{"file"},
				},
				Encoding: map[string]*huma.Encoding{
					"file": {ContentType: "application/octet-stream"},
				},
			},
		},
	}
}

func uploadAttachmentParameters() []*huma.Param {
	return []*huma.Param{
		{
			Name:        "ticket_id",
			In:          "path",
			Description: "Ticket ID.",
			Required:    true,
			Schema:      &huma.Schema{Type: "string"},
		},
		{
			Name:        "Authorization",
			In:          "header",
			Description: "Bearer API token.",
			Schema:      &huma.Schema{Type: "string"},
		},
		{
			Name:        "rayboard_session",
			In:          "cookie",
			Description: "Browser session cookie.",
			Schema:      &huma.Schema{Type: "string"},
		},
		{
			Name:        shared.CSRFCookieName,
			In:          "cookie",
			Description: "Browser CSRF cookie.",
			Schema:      &huma.Schema{Type: "string"},
		},
		{
			Name:        "X-CSRF-Token",
			In:          "header",
			Description: "CSRF token for mutating cookie-authenticated requests.",
			Schema:      &huma.Schema{Type: "string"},
		},
	}
}

func uploadAttachmentOperation(api huma.API) huma.Operation {
	op := shared.Operation(http.MethodPost, "/api/tickets/{ticket_id}/attachments", attachmentsTag, "Upload ticket attachment")
	op.DefaultStatus = http.StatusCreated
	op.MaxBodyBytes = uploadReadLimitBytes
	op.Parameters = uploadAttachmentParameters()
	op.RequestBody = uploadAttachmentRequestBody()
	op.Responses["201"] = &huma.Response{
		Description: http.StatusText(http.StatusCreated),
		Content: map[string]*huma.MediaType{
			"application/json": {
				Schema: huma.SchemaFromType(api.OpenAPI().Components.Schemas, reflectType[AttachmentResource]()),
			},
		},
	}
	return op
}

func downloadAttachmentOperation() huma.Operation {
	op := shared.Operation(http.MethodGet, "/api/attachments/{attachment_id}/download", attachmentsTag, "Download attachment")
	op.Responses["200"] = &huma.Response{
		Description: "Attachment file",
		Content: map[string]*huma.MediaType{
			"application/octet-stream": {
				Schema: &huma.Schema{Type: "string", Format: "binary"},
			},
		},
	}
	return op
}
