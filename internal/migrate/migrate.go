package migrate

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type migrationFile struct {
	version  string
	upFile   string
	downFile string
}

func collect(dir string) ([]migrationFile, error) {
	if dir == "" {
		return nil, fmt.Errorf("migrations directory is empty")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	byVersion := map[string]migrationFile{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		switch {
		case strings.HasSuffix(name, ".up.sql"):
			v := strings.TrimSuffix(name, ".up.sql")
			m := byVersion[v]
			m.version = v
			m.upFile = name
			byVersion[v] = m
		case strings.HasSuffix(name, ".down.sql"):
			v := strings.TrimSuffix(name, ".down.sql")
			m := byVersion[v]
			m.version = v
			m.downFile = name
			byVersion[v] = m
		}
	}
	if len(byVersion) == 0 {
		return nil, fmt.Errorf("no *.up.sql / *.down.sql migrations found in %s", dir)
	}

	versions := make([]string, 0, len(byVersion))
	for v := range byVersion {
		versions = append(versions, v)
	}
	sort.Strings(versions)

	out := make([]migrationFile, 0, len(versions))
	for _, v := range versions {
		out = append(out, byVersion[v])
	}
	return out, nil
}

func ensureMetaTable(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`); err != nil {
		return fmt.Errorf("schema_migrations: %w", err)
	}
	return nil
}

func Run(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	return RunUp(ctx, pool, dir)
}

func RunUp(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	files, err := collect(dir)
	if err != nil {
		return err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := ensureMetaTable(ctx, tx); err != nil {
		return err
	}

	applied := map[string]bool{}
	rows, err := tx.Query(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		applied[v] = true
	}
	rows.Close()

	for _, m := range files {
		if m.upFile == "" {
			return fmt.Errorf("missing up migration for version %s", m.version)
		}
		if applied[m.version] {
			slog.Info("migration skip", slog.String("version", m.version), slog.String("reason", "already applied"))
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, m.upFile))
		if err != nil {
			return fmt.Errorf("read %s: %w", m.upFile, err)
		}
		slog.Info("migration up", slog.String("version", m.version), slog.String("file", m.upFile))
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			return fmt.Errorf("exec %s: %w", m.upFile, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, m.version); err != nil {
			return fmt.Errorf("record %s: %w", m.version, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	slog.Info("migrations complete")
	return nil
}

func RunDown(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	files, err := collect(dir)
	if err != nil {
		return err
	}
	byVersion := map[string]migrationFile{}
	for _, m := range files {
		byVersion[m.version] = m
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := ensureMetaTable(ctx, tx); err != nil {
		return err
	}

	var version string
	err = tx.QueryRow(ctx, `
		SELECT version
		FROM schema_migrations
		ORDER BY applied_at DESC, version DESC
		LIMIT 1
	`).Scan(&version)
	if err != nil {
		if err == pgx.ErrNoRows {
			slog.Info("migration down skip", slog.String("reason", "no applied migrations"))
			return nil
		}
		return fmt.Errorf("select latest applied migration: %w", err)
	}

	m, ok := byVersion[version]
	if !ok || m.downFile == "" {
		return fmt.Errorf("missing down migration for applied version %s", version)
	}

	body, err := os.ReadFile(filepath.Join(dir, m.downFile))
	if err != nil {
		return fmt.Errorf("read %s: %w", m.downFile, err)
	}

	slog.Info("migration down", slog.String("version", version), slog.String("file", m.downFile))
	if _, err := tx.Exec(ctx, string(body)); err != nil {
		return fmt.Errorf("exec %s: %w", m.downFile, err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, version); err != nil {
		return fmt.Errorf("remove version %s: %w", version, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	slog.Info("migrations rollback complete", slog.String("version", version))
	return nil
}
