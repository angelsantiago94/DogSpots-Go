package places

import (
	"context"
	"database/sql"
	"fmt"
)

type developmentPlace struct {
	id          string
	title       string
	description string
	lat         float64
	lng         float64
	category    string
}

var developmentPlaces = []developmentPlace{
	{
		id:          "b9f57e80-4f73-4dc0-9db5-19c5c995bd11",
		title:       "Parque de Pradolongo",
		description: "Parque urbano de Madrid. Punto de referencia aproximado dentro del parque.",
		lat:         40.3758401,
		lng:         -3.7072362,
		category:    "park",
	},
	{
		id:          "0f8f4ed6-22b7-49d8-87af-cf1a4ab59e21",
		title:       "El Retiro",
		description: "Parque urbano de Madrid. Punto de referencia aproximado dentro del parque.",
		lat:         40.414940,
		lng:         -3.6832845,
		category:    "park",
	},
	{
		id:          "8aacadce-0813-48f9-bbbc-6a3350d54876",
		title:       "Casa de Campo",
		description: "Parque forestal de Madrid. Punto de referencia aproximado dentro del parque.",
		lat:         40.424190,
		lng:         -3.7558580,
		category:    "park",
	},
	{
		id:          "18a07a90-7470-4627-963b-58a782789e0f",
		title:       "Madrid Río",
		description: "Parque urbano junto al río Manzanares. Punto de referencia aproximado.",
		lat:         40.4005966,
		lng:         -3.7207547,
		category:    "park",
	},
}

func SeedDevelopmentData(ctx context.Context, database *sql.DB) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin development seed transaction: %w", err)
	}
	defer tx.Rollback()

	const demoUserID = "8e399959-697a-4dd6-b068-e5064d0f4a4b"
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO users
		(id, email, password_hash, name, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		demoUserID, "demo@dogspots.local", "!disabled-demo-account!", "DogSpots Demo"); err != nil {
		return fmt.Errorf("insert development demo user: %w", err)
	}
	for _, place := range developmentPlaces {
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO places
			(id, author_id, title, description, lat, lng, category, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			place.id, demoUserID, place.title, place.description, place.lat, place.lng, place.category); err != nil {
			return fmt.Errorf("insert development place %q: %w", place.title, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit development seed: %w", err)
	}
	return nil
}
