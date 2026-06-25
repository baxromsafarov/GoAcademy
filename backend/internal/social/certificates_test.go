package social

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/progress"
)

func scanID(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&id); err != nil {
		t.Fatalf("seed (%s): %v", sql, err)
	}
	return id
}

func TestCertificates_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	prog := progress.NewService(pool, activity.NoopRecorder{})
	svc := NewCertificatesService(pool, prog)
	owner := seedUserStats(t, pool, "cert-owner", true, false, 0)

	// A track with a single video item.
	marker := newUUID(t)
	vid := scanID(t, pool, "INSERT INTO videos (title, youtube_id, tags) VALUES ('CV','y',ARRAY[$1]) RETURNING id::text", marker)
	track := scanID(t, pool, "INSERT INTO tracks (title, language) VALUES ($1,'en') RETURNING id::text", "Cert Track "+marker)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM tracks WHERE id = $1", track)
		_, _ = pool.Exec(context.Background(), "DELETE FROM videos WHERE $1 = ANY(tags)", marker)
	})
	if _, err := pool.Exec(ctx, "INSERT INTO track_items (track_id, content_type, content_id, position) VALUES ($1,'video',$2::uuid,1)", track, vid); err != nil {
		t.Fatalf("seed track item: %v", err)
	}

	// Track not complete -> issuing is a conflict.
	if _, err := svc.IssueForTrack(ctx, owner, track); err == nil {
		t.Error("issuing before completion should fail")
	}

	// Complete the only item.
	if _, err := pool.Exec(ctx, "INSERT INTO video_progress (user_id, video_id, watched_percent, completed) VALUES ($1::uuid,$2::uuid,100,true)", owner, vid); err != nil {
		t.Fatalf("complete video: %v", err)
	}

	// Now a certificate is issued.
	cert, err := svc.IssueForTrack(ctx, owner, track)
	if err != nil {
		t.Fatalf("IssueForTrack: %v", err)
	}
	if cert.Code == "" || cert.TrackID != track {
		t.Fatalf("unexpected certificate: %+v", cert)
	}

	// Issuing again returns the SAME certificate (one per user+track).
	cert2, err := svc.IssueForTrack(ctx, owner, track)
	if err != nil || cert2.Code != cert.Code {
		t.Errorf("idempotent issue: got (%+v, %v), want same code %s", cert2, err, cert.Code)
	}

	// Public verification resolves the owner and track.
	v, err := svc.Verify(ctx, cert.Code)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if v.DisplayName != "cert-owner" || v.TrackTitle != "Cert Track "+marker {
		t.Errorf("verification = %+v", v)
	}

	// It appears in the owner's certificate list.
	if certs, _ := svc.ListForUser(ctx, owner); len(certs) != 1 || certs[0].Code != cert.Code {
		t.Errorf("list = %+v, want the one certificate", certs)
	}

	// Unknown code / unknown track.
	if _, err := svc.Verify(ctx, "GOAC-DOESNOTEXIST"); err == nil {
		t.Error("unknown code should be not found")
	}
	if _, err := svc.IssueForTrack(ctx, owner, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Error("unknown track should error")
	}
}
