//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

// Example: the happy path. Create a user, then read it back both ways.
func TestUserRepo_CreateAndGet(t *testing.T) {
	repo := newUserRepo(t)
	ctx := context.Background()

	// Arrange + Act: create a user.
	created, err := repo.Create(ctx, "vlad@example.com", "hash-123")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Assert: the database filled in id and created_at for us.
	if len(created.ID) != 36 { // uuid: 8-4-4-4-12 characters
		t.Errorf("ID = %q, want a uuid", created.ID)
	}
	if created.Email != "vlad@example.com" || created.PasswordHash != "hash-123" {
		t.Errorf("created = %+v", created)
	}
	if time.Since(created.CreatedAt) > time.Minute {
		t.Errorf("CreatedAt = %v, want about now", created.CreatedAt)
	}

	// Read it back by ID and by email: both must return the same row.
	byID, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	byEmail, err := repo.GetByEmail(ctx, "vlad@example.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	for name, got := range map[string]string{"GetByID": byID.ID, "GetByEmail": byEmail.ID} {
		if got != created.ID {
			t.Errorf("%s returned ID %q, want %q", name, got, created.ID)
		}
	}
}

// ---------------------------------------------------------------------
// ЗАДАНИЕ ДЛЯ ВЛАДА. Убери t.Skip и напиши тело теста.
// Образец — тест выше. Ошибки сравнивай через errors.Is(err, user.ErrXxx)
// (импорт: "errors" и "github.com/vdyabkin01-ux/goledger/internal/user").
// Запуск: make test-integration
// ---------------------------------------------------------------------

func TestUserRepo_Create_DuplicateEmail(t *testing.T) {
	t.Skip("TODO(Влад): создать пользователя, создать второго с тем же email — ждём user.ErrEmailTaken")
}

func TestUserRepo_GetByEmail_NotFound(t *testing.T) {
	t.Skip("TODO(Влад): пустая база, GetByEmail любого email — ждём user.ErrNotFound")
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	t.Skip("TODO(Влад): GetByID с валидным, но несуществующим uuid — ждём user.ErrNotFound")
}

func TestUserRepo_GetByID_InvalidUUID(t *testing.T) {
	t.Skip(`TODO(Влад): GetByID("not-a-uuid") — ждём user.ErrNotFound, а не какую-то другую ошибку`)
}

func TestUserRepo_Create_UppercaseEmailRejected(t *testing.T) {
	t.Skip(`TODO(Влад): Create("Vlad@Example.com", ...) — база должна отказать (CHECK в миграции).
Ждём: err != nil и при этом это НЕ user.ErrEmailTaken`)
}
