package comments

import (
	commentservice "github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

type listCommentsInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id" doc:"Ticket ID."`
}

type listCommentsOutput struct {
	Body shared.ItemList[commentservice.Comment]
}

type createCommentInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id" doc:"Ticket ID."`
	Body     *createCommentBody
}

type createCommentBody struct {
	Body string `json:"body" doc:"Comment body."`
}

type createCommentOutput struct {
	Status int `status:"201"`
	Body   commentservice.Comment
}

type deleteCommentInput struct {
	shared.AuthInput
	CommentID string `path:"comment_id" doc:"Comment ID."`
}
