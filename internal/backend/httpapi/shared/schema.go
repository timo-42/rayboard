package shared

type EmptyOutput struct {
}

type ListMetadata struct {
	Count int `json:"count"`
}

type ListSpec struct {
}

type ListStatus[T any] struct {
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

type ListResource[T any] = Resource[ListMetadata, ListSpec, ListStatus[T]]

func NewListResource[T any](items []T) ListResource[T] {
	return ListResource[T]{
		Metadata: ListMetadata{Count: len(items)},
		Spec:     ListSpec{},
		Status:   ListStatus[T]{Items: items},
	}
}

type ListOutput[T any] struct {
	Body ListResource[T]
}

type CreatedOutput[T any] struct {
	Body T
}

type AcceptedOutput[T any] struct {
	Body T
}
