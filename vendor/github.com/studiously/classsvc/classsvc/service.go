package classsvc

import (
	"context"

	"github.com/google/uuid"
	"github.com/studiously/classsvc/models"
)

type Service interface {
	// ListClasses gets all classes the current user is enrolled in.
	ListClasses(ctx context.Context) ([]uuid.UUID, error)
	// GetClass gets details for a specific class.
	GetClass(ctx context.Context, id uuid.UUID) (*models.Class, error)
	// CreateClass creates a class and enrolls the current user in it as an administrator.
	CreateClass(ctx context.Context, name string) (uuid.UUID, error)
	// UpdateClass updates a class.
	UpdateClass(ctx context.Context, id uuid.UUID, name string, currentUnit uuid.UUID) error
	// DeleteClass deactivates a class.
	DeleteClass(ctx context.Context, id uuid.UUID) error
	// JoinClass enrolls the current user in a class.
	JoinClass(ctx context.Context, class uuid.UUID) (error)
	// LeaveClass causes a user to be un-enrolled from a class.
	// If user is not `uuid.Nil`, then LeaveClass removes the other user, requiring the current user to have elevated permissions.
	LeaveClass(ctx context.Context, user uuid.UUID, class uuid.UUID) error
	// SetRole sets the role of a user in a class.
	// The current user must have a higher role than the target user.
	SetRole(ctx context.Context, user uuid.UUID, class uuid.UUID, role models.UserRole) error
	// ListMembers lists all members of a class and their role.
	ListMembers(ctx context.Context, classId uuid.UUID) ([]*models.Member, error)
}
