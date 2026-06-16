package shared

type EmptyOutput struct {
}

type ItemList[T any] struct {
	Items []T `json:"items"`
}

type ResourceInput[Spec any] struct {
	Spec Spec `json:"spec"`
}

type Resource[Metadata any, Spec any, Status any] struct {
	Metadata Metadata `json:"metadata"`
	Spec     Spec     `json:"spec"`
	Status   Status   `json:"status"`
}

type ListOutput[T any] struct {
	Body ItemList[T]
}

type CreatedOutput[T any] struct {
	Body T
}

type AcceptedOutput[T any] struct {
	Body T
}
