package migrations

import (
	"context"

	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/migrations"
)

func GetMigrations() migrations.MigrationSource {
	builder := migrations.NewMigrationBuilder("gorest-likeable")

	builder.Add(
		"20260102000002000",
		"create_likes_table",
		func(ctx context.Context, db database.Database) error {
			if err := migrations.SQL(ctx, db, migrations.DialectSQL{
				Postgres: `CREATE TABLE IF NOT EXISTS likes (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					liker_id UUID,
					liked_id UUID,
					likeable_id UUID NOT NULL,
					likeable TEXT NOT NULL,
					ip_address TEXT,
					user_agent TEXT,
					liked_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP(0) WITH TIME ZONE,
					created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
					UNIQUE (liker_id, likeable, likeable_id)
				)`,
				MySQL: `CREATE TABLE IF NOT EXISTS likes (
					id CHAR(36) PRIMARY KEY,
					liker_id CHAR(36),
					liked_id CHAR(36),
					likeable_id CHAR(36) NOT NULL,
					likeable VARCHAR(255) NOT NULL,
					ip_address VARCHAR(255),
					user_agent TEXT,
					liked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					UNIQUE KEY unique_like (liker_id, likeable, likeable_id),
					INDEX idx_likeable (likeable, likeable_id, liked_at),
					INDEX idx_liker_id (liker_id),
					INDEX idx_anonymous_like (ip_address, user_agent(255))
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
				SQLite: `CREATE TABLE IF NOT EXISTS likes (
					id TEXT PRIMARY KEY,
					liker_id TEXT,
					liked_id TEXT,
					likeable_id TEXT NOT NULL,
					likeable TEXT NOT NULL,
					ip_address TEXT,
					user_agent TEXT,
					liked_at TEXT NOT NULL DEFAULT (datetime('now')),
					updated_at TEXT,
					created_at TEXT NOT NULL DEFAULT (datetime('now')),
					UNIQUE (liker_id, likeable, likeable_id)
				)`,
			}); err != nil {
				return err
			}

			// Create indexes for Postgres and SQLite
			if db.DriverName() == "postgres" {
				// Composite index for likeable queries
				if err := migrations.SQL(ctx, db, migrations.DialectSQL{
					Postgres: `CREATE INDEX IF NOT EXISTS idx_likeable ON likes(likeable, likeable_id, liked_at)`,
				}); err != nil {
					return err
				}
				if err := migrations.CreateIndex(ctx, db, "idx_liker_id", "likes", "liker_id"); err != nil {
					return err
				}
				// Index for anonymous likes
				if err := migrations.SQL(ctx, db, migrations.DialectSQL{
					Postgres: `CREATE INDEX IF NOT EXISTS idx_anonymous_like ON likes(ip_address, user_agent)`,
				}); err != nil {
					return err
				}
			}

			if db.DriverName() == "sqlite" {
				// Composite index for likeable queries
				if err := migrations.SQL(ctx, db, migrations.DialectSQL{
					SQLite: `CREATE INDEX IF NOT EXISTS idx_likeable ON likes(likeable, likeable_id, liked_at)`,
				}); err != nil {
					return err
				}
				if err := migrations.CreateIndex(ctx, db, "idx_liker_id", "likes", "liker_id"); err != nil {
					return err
				}
				// Index for anonymous likes
				if err := migrations.SQL(ctx, db, migrations.DialectSQL{
					SQLite: `CREATE INDEX IF NOT EXISTS idx_anonymous_like ON likes(ip_address, user_agent)`,
				}); err != nil {
					return err
				}
			}

			return nil
		},
		func(ctx context.Context, db database.Database) error {
			// Drop indexes first
			if db.DriverName() == "postgres" || db.DriverName() == "sqlite" {
				_ = migrations.DropIndex(ctx, db, "idx_likeable", "likes")
				_ = migrations.DropIndex(ctx, db, "idx_liker_id", "likes")
				_ = migrations.DropIndex(ctx, db, "idx_anonymous_like", "likes")
			}

			return migrations.DropTableIfExists(ctx, db, "likes")
		},
	)

	builder.Add(
		"20260209000003000",
		"fix_unique_constraints_for_anonymous_likes",
		func(ctx context.Context, db database.Database) error {
			if err := migrations.SQL(ctx, db, migrations.DialectSQL{
				Postgres: `
					ALTER TABLE likes DROP CONSTRAINT IF EXISTS likes_liker_id_likeable_likeable_id_key;

					CREATE UNIQUE INDEX IF NOT EXISTS unique_authenticated_like
					ON likes(liker_id, likeable, likeable_id)
					WHERE liker_id IS NOT NULL;

					CREATE UNIQUE INDEX IF NOT EXISTS unique_anonymous_like
					ON likes(ip_address, user_agent, likeable, likeable_id)
					WHERE liker_id IS NULL;
				`,
				MySQL: `
					ALTER TABLE likes DROP INDEX IF EXISTS unique_like;

					CREATE UNIQUE INDEX unique_authenticated_like
					ON likes(liker_id, likeable, likeable_id);

					CREATE UNIQUE INDEX unique_anonymous_like
					ON likes(ip_address(255), user_agent(255), likeable, likeable_id, liker_id);
				`,
				SQLite: `
					DROP INDEX IF EXISTS unique_like;

					CREATE UNIQUE INDEX IF NOT EXISTS unique_authenticated_like
					ON likes(liker_id, likeable, likeable_id)
					WHERE liker_id IS NOT NULL;

					CREATE UNIQUE INDEX IF NOT EXISTS unique_anonymous_like
					ON likes(ip_address, user_agent, likeable, likeable_id)
					WHERE liker_id IS NULL;
				`,
			}); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, db database.Database) error {
			if err := migrations.SQL(ctx, db, migrations.DialectSQL{
				Postgres: `
					DROP INDEX IF EXISTS unique_authenticated_like;
					DROP INDEX IF EXISTS unique_anonymous_like;
					ALTER TABLE likes ADD CONSTRAINT likes_liker_id_likeable_likeable_id_key
					UNIQUE (liker_id, likeable, likeable_id);
				`,
				MySQL: `
					DROP INDEX IF EXISTS unique_authenticated_like ON likes;
					DROP INDEX IF EXISTS unique_anonymous_like ON likes;
					CREATE UNIQUE INDEX unique_like ON likes(liker_id, likeable, likeable_id);
				`,
				SQLite: `
					DROP INDEX IF EXISTS unique_authenticated_like;
					DROP INDEX IF EXISTS unique_anonymous_like;
				`,
			}); err != nil {
				return err
			}

			return nil
		},
	)

	return builder.Build()
}
