package shared

type EmptyOutput struct {
}

type ItemList[T any] struct {
	Items []T `json:"items"`
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
