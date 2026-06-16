package comments

import (
	"time"

	commentservice "github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

type listCommentsInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id" doc:"Ticket ID."`
}

type listCommentsOutput struct {
	Body shared.ListResource[CommentResource]
}

type createCommentInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id" doc:"Ticket ID."`
	Body     shared.ResourceInput[CommentSpec]
}

type createCommentOutput = shared.CreatedOutput[CommentResource]

type deleteCommentInput struct {
	shared.AuthInput
	CommentID string `path:"comment_id" doc:"Comment ID."`
}

type CommentMetadata struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CommentSpec struct {
	Body string `json:"body" doc:"Comment body."`
}

type CommentStatus struct {
	AuthorID string `json:"author_id,omitempty"`
}

type CommentResource = shared.Resource[CommentMetadata, CommentSpec, CommentStatus]

func commentResource(comment commentservice.Comment) CommentResource {
	return CommentResource{
		Metadata: CommentMetadata{
			ID:        comment.ID,
			TicketID:  comment.TicketID,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		},
		Spec: CommentSpec{
			Body: comment.Body,
		},
		Status: CommentStatus{
			AuthorID: comment.AuthorID,
		},
	}
}

func commentResources(comments []commentservice.Comment) []CommentResource {
	resources := make([]CommentResource, 0, len(comments))
	for _, comment := range comments {
		resources = append(resources, commentResource(comment))
	}
	return resources
}
