package places

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	ListTags(context.Context) ([]Tag, error)
	Create(context.Context, uuid.UUID, CreatePlaceRequest) (Place, error)
	Get(context.Context, uuid.UUID, *uuid.UUID) (Place, error)
	Update(context.Context, uuid.UUID, uuid.UUID, UpdatePlaceRequest) (Place, error)
	Delete(context.Context, uuid.UUID, uuid.UUID) error
	List(context.Context, ListOptions) (PlaceList, error)
	ValidateTags(context.Context, []string) error
}
