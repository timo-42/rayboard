package comments

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	commentservice "github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

const commentsTag = "Comments"

type routes struct {
	provider Provider
}

func Register(api huma.API, provider Provider) {
	route := routes{provider: provider}

	huma.Register(api, shared.Operation(http.MethodGet, "/api/tickets/{ticket_id}/comments", commentsTag, "List ticket comments"), route.list)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/tickets/{ticket_id}/comments", commentsTag, "Create ticket comment", http.StatusCreated), route.create)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/comments/{comment_id}", commentsTag, "Delete comment", http.StatusNoContent), route.delete)
}

func (r routes) list(ctx context.Context, input *listCommentsInput) (*listCommentsOutput, error) {
	ctx, principal, _, err := r.provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}

	items, err := r.provider.Comments.List(ctx, principal, input.TicketID)
	if err != nil {
		return nil, shared.CommentError(err)
	}
	return &listCommentsOutput{Body: shared.NewListResource[CommentResource](commentResources(items))}, nil
}

func (r routes) create(ctx context.Context, input *createCommentInput) (*createCommentOutput, error) {
	ctx, principal, _, err := r.provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}

	comment, err := r.provider.Comments.Create(ctx, principal, commentservice.CreateInput{
		TicketID: input.TicketID,
		Body:     input.Body.Spec.Body,
	})
	if err != nil {
		return nil, shared.CommentError(err)
	}
	return &createCommentOutput{Body: commentResource(comment)}, nil
}

func (r routes) delete(ctx context.Context, input *deleteCommentInput) (*shared.EmptyOutput, error) {
	ctx, principal, _, err := r.provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}

	if err := r.provider.Comments.Delete(ctx, principal, input.CommentID); err != nil {
		return nil, shared.CommentError(err)
	}
	return &shared.EmptyOutput{}, nil
}
