// Test query sqlc pertama untuk pengguna dan paket (test case TC-M01-020,
// docs/test-cases/M01-database.md).

package store_test

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"pfmea/backend/internal/store/sqlc"
)

// TestQueries_TC_M01_020 memastikan query sqlc membaca pengguna dan paket demo dengan tipe Go
// yang diharapkan (uuid.UUID, time.Time), id yang tidak ada menghasilkan pgx.ErrNoRows, dan
// daftar pengguna tidak membawa hash kata sandi.
func TestQueries_TC_M01_020(t *testing.T) {
	ctx := context.Background()
	db, _ := newStore(t)
	q := sqlc.New(db.Pool)

	u, err := q.GetUserByUsername(ctx, "apratama")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	// Tipe Go hasil generator diperiksa saat kompilasi (override di sqlc.yaml).
	id := hasType[uuid.UUID](u.ID)
	created := hasType[time.Time](u.CreatedAt)
	if id != apratama || u.Role != sqlc.UserRoleAuthor || created.IsZero() || u.PasswordHash == nil {
		t.Errorf("apratama = %+v, want role author with a password hash", u)
	}

	byID, err := q.GetUserByID(ctx, apratama)
	if err != nil || byID.Username != "apratama" {
		t.Errorf("GetUserByID = %+v, %v", byID, err)
	}
	if _, err := q.GetUserByID(ctx, uuid.New()); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("unknown user: err = %v, want pgx.ErrNoRows", err)
	}

	users, err := q.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	names := make([]string, len(users))
	for i, u := range users {
		names[i] = u.Username
	}
	want := []string{"admin", "apratama", "dhidayat", "operator1", "rsaputri", "swulandari"}
	if !slices.Equal(names, want) {
		t.Errorf("ListUsers = %v, want %v", names, want)
	}
	if len(users) > 0 {
		if _, ok := reflect.TypeOf(users[0]).FieldByName("PasswordHash"); ok {
			t.Errorf("ListUsers rows must not carry the password hash")
		}
	}

	pkg, err := q.GetPackageByID(ctx, ps07)
	if err != nil || pkg.Code != "PS-07" || pkg.Kind != sqlc.PackageKindModel {
		t.Errorf("GetPackageByID(PS-07) = %+v, %v", pkg, err)
	}
	if _, err := q.GetPackageByID(ctx, uuid.New()); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("unknown package: err = %v, want pgx.ErrNoRows", err)
	}

	pkgs, err := q.ListPackages(ctx)
	if err != nil {
		t.Fatalf("ListPackages: %v", err)
	}
	if len(pkgs) != 2 || pkgs[0].Code != "GENERAL" || pkgs[0].Kind != sqlc.PackageKindGeneral ||
		pkgs[1].Code != "PS-07" || pkgs[1].Kind != sqlc.PackageKindModel {
		t.Errorf("ListPackages = %+v, want GENERAL (general) then PS-07 (model)", pkgs)
	}
}

// hasType mengembalikan v apa adanya; kompilasi gagal bila v bukan bertipe T.
func hasType[T any](v T) T { return v }
