package sqlite

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"webaviable/storage"
)

func newTestStorage(t *testing.T) *Storage {
	t.Helper()

	s, err := New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	return s
}

func mustSaveURL(t *testing.T, s *Storage, alias, url string) int64 {
	t.Helper()

	id, err := s.SaveURL(alias, url)
	if err != nil {
		t.Fatalf("failed to save URL %q: %v", alias, err)
	}
	return id
}

func mustSaveCheck(t *testing.T, s *Storage, urlID int64, status int, timeMs int64, at time.Time) int64 {
	t.Helper()

	id, err := s.SaveCheck(urlID, status, timeMs, at)
	if err != nil {
		t.Fatalf("failed to save check: %v", err)
	}
	return id
}

func mustGetChecks(t *testing.T, s *Storage, urlID int64) []storage.Check {
	t.Helper()

	checks, err := s.GetChecks(urlID)
	if err != nil {
		t.Fatalf("failed to get checks: %v", err)
	}
	return checks
}

func TestNewIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	s1, err := New(path)
	if err != nil {
		t.Fatalf("first New failed: %v", err)
	}
	urlID := mustSaveURL(t, s1, "example", "https://example.com")
	mustSaveCheck(t, s1, urlID, 200, 50, time.Now())
	if err := s1.Close(); err != nil {
		t.Fatalf("failed to close storage: %v", err)
	}

	s2, err := New(path)
	if err != nil {
		t.Fatalf("second New failed: %v", err)
	}
	defer s2.Close()

	got, err := s2.GetURL("example")
	if err != nil {
		t.Fatalf("failed to get URL after reopen: %v", err)
	}
	if got.ID != urlID {
		t.Fatalf("expected URL ID %d after reopen, got %d", urlID, got.ID)
	}
	if n := len(mustGetChecks(t, s2, urlID)); n != 1 {
		t.Fatalf("expected 1 check after reopen, got %d", n)
	}
}

func TestNewInvalidPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no", "such", "dir", "test.db")

	s, err := New(path)
	if err == nil {
		s.Close()
		t.Fatal("expected error for invalid path, got nil")
	}
}

func TestGetURL(t *testing.T) {
	s := newTestStorage(t)

	id := mustSaveURL(t, s, "example", "https://example.com")

	got, err := s.GetURL("example")
	if err != nil {
		t.Fatalf("failed to get URL: %v", err)
	}
	if got.ID != id {
		t.Fatalf("expected URL ID %d, got %d", id, got.ID)
	}
	if got.Alias != "example" {
		t.Fatalf("expected alias 'example', got '%s'", got.Alias)
	}
	if got.URL != "https://example.com" {
		t.Fatalf("expected URL 'https://example.com', got '%s'", got.URL)
	}
}

func TestGetURLNotFound(t *testing.T) {
	s := newTestStorage(t)

	_, err := s.GetURL("nonexistent")
	if !errors.Is(err, storage.ErrURLNotFound) {
		t.Fatalf("expected ErrURLNotFound, got %v", err)
	}
}

func TestSaveURLDuplicateAlias(t *testing.T) {
	s := newTestStorage(t)

	mustSaveURL(t, s, "example", "https://example.com")

	_, err := s.SaveURL("example", "https://other.com")
	if !errors.Is(err, storage.ErrURLExists) {
		t.Fatalf("expected ErrURLExists, got %v", err)
	}

	got, err := s.GetURL("example")
	if err != nil {
		t.Fatalf("failed to get URL: %v", err)
	}
	if got.URL != "https://example.com" {
		t.Fatalf("original URL was overwritten: got '%s'", got.URL)
	}
}

func TestSaveURLSameURLDifferentAliases(t *testing.T) {
	s := newTestStorage(t)

	id1 := mustSaveURL(t, s, "a", "https://example.com")
	id2 := mustSaveURL(t, s, "b", "https://example.com")
	if id1 == id2 {
		t.Fatalf("expected different IDs, both are %d", id1)
	}
}

func TestGetAllURLsEmpty(t *testing.T) {
	s := newTestStorage(t)

	urls, err := s.GetAllURLs()
	if err != nil {
		t.Fatalf("failed to get all URLs: %v", err)
	}
	if len(urls) != 0 {
		t.Fatalf("expected 0 URLs, got %d", len(urls))
	}
}

func TestGetAllURLsOrderedByID(t *testing.T) {
	s := newTestStorage(t)

	want := []storage.URL{
		{Alias: "google", URL: "https://google.com"},
		{Alias: "github", URL: "https://github.com"},
		{Alias: "golang", URL: "https://go.dev"},
	}
	for i := range want {
		want[i].ID = mustSaveURL(t, s, want[i].Alias, want[i].URL)
	}

	got, err := s.GetAllURLs()
	if err != nil {
		t.Fatalf("failed to get all URLs: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d URLs, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("URL #%d: expected %+v, got %+v", i, want[i], got[i])
		}
	}
}

func TestDeleteURL(t *testing.T) {
	s := newTestStorage(t)

	mustSaveURL(t, s, "example", "https://example.com")
	mustSaveURL(t, s, "keep", "https://keep.com")

	if err := s.DeleteURL("example"); err != nil {
		t.Fatalf("failed to delete URL: %v", err)
	}

	if _, err := s.GetURL("example"); !errors.Is(err, storage.ErrURLNotFound) {
		t.Fatalf("expected ErrURLNotFound after delete, got %v", err)
	}
	if _, err := s.GetURL("keep"); err != nil {
		t.Fatalf("unrelated URL was affected by delete: %v", err)
	}
}

func TestDeleteURLNotFound(t *testing.T) {
	s := newTestStorage(t)

	err := s.DeleteURL("nonexistent")
	if !errors.Is(err, storage.ErrURLNotFound) {
		t.Fatalf("expected ErrURLNotFound, got %v", err)
	}
}

func TestDeleteURLTwice(t *testing.T) {
	s := newTestStorage(t)

	mustSaveURL(t, s, "example", "https://example.com")

	if err := s.DeleteURL("example"); err != nil {
		t.Fatalf("first delete failed: %v", err)
	}
	if err := s.DeleteURL("example"); !errors.Is(err, storage.ErrURLNotFound) {
		t.Fatalf("expected ErrURLNotFound on second delete, got %v", err)
	}
}

func TestDeleteURLAllowsAliasReuse(t *testing.T) {
	s := newTestStorage(t)

	mustSaveURL(t, s, "example", "https://old.com")
	if err := s.DeleteURL("example"); err != nil {
		t.Fatalf("failed to delete URL: %v", err)
	}
	mustSaveURL(t, s, "example", "https://new.com")

	got, err := s.GetURL("example")
	if err != nil {
		t.Fatalf("failed to get URL: %v", err)
	}
	if got.URL != "https://new.com" {
		t.Fatalf("expected 'https://new.com', got '%s'", got.URL)
	}
}

func TestDeleteURLCascadesChecks(t *testing.T) {
	s := newTestStorage(t)

	delID := mustSaveURL(t, s, "del", "https://del.com")
	keepID := mustSaveURL(t, s, "keep", "https://keep.com")

	mustSaveCheck(t, s, delID, 200, 10, time.Now())
	mustSaveCheck(t, s, delID, 500, 20, time.Now())
	mustSaveCheck(t, s, keepID, 200, 30, time.Now())

	if err := s.DeleteURL("del"); err != nil {
		t.Fatalf("failed to delete URL: %v", err)
	}

	var n int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM checks WHERE url_id = ?", delID).Scan(&n); err != nil {
		t.Fatalf("failed to count checks: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected checks to be deleted by cascade, %d left", n)
	}

	if got := len(mustGetChecks(t, s, keepID)); got != 1 {
		t.Fatalf("expected 1 check for untouched URL, got %d", got)
	}
}

func TestSaveCheckUnknownURL(t *testing.T) {
	s := newTestStorage(t)

	if _, err := s.SaveCheck(9999, 200, 10, time.Now()); err == nil {
		t.Fatal("expected foreign key error, got nil")
	}
}

func TestSaveCheckReturnsUniqueIDs(t *testing.T) {
	s := newTestStorage(t)

	urlID := mustSaveURL(t, s, "example", "https://example.com")

	id1 := mustSaveCheck(t, s, urlID, 200, 1, time.Now())
	id2 := mustSaveCheck(t, s, urlID, 200, 2, time.Now())
	if id1 == id2 {
		t.Fatalf("expected different check IDs, both are %d", id1)
	}
}

func TestSaveCheckStoresUTC(t *testing.T) {
	s := newTestStorage(t)

	urlID := mustSaveURL(t, s, "example", "https://example.com")

	loc := time.FixedZone("UTC+5", 5*60*60)
	at := time.Now().Add(-time.Hour).In(loc).Truncate(time.Millisecond)
	mustSaveCheck(t, s, urlID, 200, 10, at)

	checks := mustGetChecks(t, s, urlID)
	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}
	if !checks[0].CheckedAt.Equal(at) {
		t.Fatalf("expected CheckedAt %v, got %v", at, checks[0].CheckedAt)
	}
}

func TestGetChecks(t *testing.T) {
	s := newTestStorage(t)

	urlID := mustSaveURL(t, s, "example", "https://example.com")
	at := time.Now().Truncate(time.Millisecond)
	checkID := mustSaveCheck(t, s, urlID, 200, 123, at)

	checks := mustGetChecks(t, s, urlID)
	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}
	c := checks[0]
	if c.ID != checkID {
		t.Fatalf("expected check ID %d, got %d", checkID, c.ID)
	}
	if c.URLID != urlID {
		t.Fatalf("expected URL ID %d, got %d", urlID, c.URLID)
	}
	if c.HTTPStatus != 200 {
		t.Fatalf("expected HTTP status 200, got %d", c.HTTPStatus)
	}
	if c.ResponseTimeMs != 123 {
		t.Fatalf("expected response time 123, got %d", c.ResponseTimeMs)
	}
	if !c.CheckedAt.Equal(at) {
		t.Fatalf("expected CheckedAt %v, got %v", at, c.CheckedAt)
	}
}

func TestGetChecksNoChecks(t *testing.T) {
	s := newTestStorage(t)

	urlID := mustSaveURL(t, s, "example", "https://example.com")

	if n := len(mustGetChecks(t, s, urlID)); n != 0 {
		t.Fatalf("expected 0 checks, got %d", n)
	}
}

func TestGetChecksUnknownURL(t *testing.T) {
	s := newTestStorage(t)

	if n := len(mustGetChecks(t, s, 9999)); n != 0 {
		t.Fatalf("expected 0 checks for unknown URL, got %d", n)
	}
}

func TestGetChecksOrderedNewestFirst(t *testing.T) {
	s := newTestStorage(t)

	urlID := mustSaveURL(t, s, "example", "https://example.com")
	now := time.Now().Truncate(time.Millisecond)

	mustSaveCheck(t, s, urlID, 301, 30, now.Add(-3*time.Hour))
	mustSaveCheck(t, s, urlID, 200, 10, now.Add(-1*time.Hour))
	mustSaveCheck(t, s, urlID, 500, 20, now.Add(-2*time.Hour))

	checks := mustGetChecks(t, s, urlID)

	wantStatus := []int{200, 500, 301}
	if len(checks) != len(wantStatus) {
		t.Fatalf("expected %d checks, got %d", len(wantStatus), len(checks))
	}
	for i, want := range wantStatus {
		if int(checks[i].HTTPStatus) != want {
			t.Fatalf("check #%d: expected status %d, got %d", i, want, checks[i].HTTPStatus)
		}
	}
	for i := 1; i < len(checks); i++ {
		if checks[i].CheckedAt.After(checks[i-1].CheckedAt) {
			t.Fatalf("checks are not sorted DESC at #%d: %v after %v",
				i, checks[i].CheckedAt, checks[i-1].CheckedAt)
		}
	}
}

func TestGetChecksOnlyLast24Hours(t *testing.T) {
	s := newTestStorage(t)

	urlID := mustSaveURL(t, s, "example", "https://example.com")
	now := time.Now()

	mustSaveCheck(t, s, urlID, 200, 10, now.Add(-1*time.Hour))
	mustSaveCheck(t, s, urlID, 200, 20, now.Add(-23*time.Hour))
	mustSaveCheck(t, s, urlID, 500, 30, now.Add(-25*time.Hour))
	mustSaveCheck(t, s, urlID, 500, 40, now.Add(-7*24*time.Hour))

	checks := mustGetChecks(t, s, urlID)
	if len(checks) != 2 {
		t.Fatalf("expected 2 checks within 24h, got %d", len(checks))
	}
	for _, c := range checks {
		if c.HTTPStatus != 200 {
			t.Fatalf("old check leaked into result: %+v", c)
		}
	}
}

func TestGetChecksIsolatedByURL(t *testing.T) {
	s := newTestStorage(t)

	firstID := mustSaveURL(t, s, "first", "https://first.com")
	secondID := mustSaveURL(t, s, "second", "https://second.com")

	for i := 0; i < 3; i++ {
		mustSaveCheck(t, s, firstID, 200, 10, time.Now())
	}
	mustSaveCheck(t, s, secondID, 500, 20, time.Now())

	first := mustGetChecks(t, s, firstID)
	if len(first) != 3 {
		t.Fatalf("expected 3 checks for first URL, got %d", len(first))
	}
	for _, c := range first {
		if c.URLID != firstID {
			t.Fatalf("check of URL %d leaked into URL %d", c.URLID, firstID)
		}
	}

	second := mustGetChecks(t, s, secondID)
	if len(second) != 1 {
		t.Fatalf("expected 1 check for second URL, got %d", len(second))
	}
	if second[0].HTTPStatus != 500 {
		t.Fatalf("expected HTTP status 500, got %d", second[0].HTTPStatus)
	}
}
