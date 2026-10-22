package places

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("place not found")
	ErrForbidden     = errors.New("place belongs to another user")
	ErrInvalidTags   = errors.New("one or more tags do not exist")
	ErrInvalidCursor = errors.New("invalid cursor")
)

type SQLiteRepository struct {
	db *sql.DB
}

type queryer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type cursorKey struct {
	Sort      string    `json:"s"`
	Value     float64   `json:"v"`
	CreatedAt time.Time `json:"t"`
	ID        uuid.UUID `json:"i"`
}

func NewSQLiteRepository(db *sql.DB) (*SQLiteRepository, error) {
	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) ListTags(ctx context.Context) ([]Tag, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT slug, name FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags := make([]Tag, 0)
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.Slug, &tag.Label); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *SQLiteRepository) ValidateTags(ctx context.Context, slugs []string) error {
	if len(slugs) == 0 {
		return nil
	}
	query, args := inClause(`SELECT COUNT(*) FROM tags WHERE slug IN (%s)`, slugs)
	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return err
	}
	if count != len(slugs) {
		return ErrInvalidTags
	}
	return nil
}

func (r *SQLiteRepository) Create(ctx context.Context, authorID uuid.UUID, request CreatePlaceRequest) (Place, error) {
	id := uuid.New()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Place{}, err
	}
	defer tx.Rollback()
	if err := insertPlace(ctx, tx, id, authorID, request); err != nil {
		return Place{}, err
	}
	if err := replaceTags(ctx, tx, id, request.Tags); err != nil {
		return Place{}, err
	}
	if err := tx.Commit(); err != nil {
		return Place{}, err
	}
	return r.Get(ctx, id, nil)
}

func insertPlace(ctx context.Context, tx *sql.Tx, id, authorID uuid.UUID, request CreatePlaceRequest) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO places
		(id, author_id, title, description, lat, lng, category, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, id.String(), authorID.String(), request.Title,
		nullableText(request.Description), request.Lat, request.Lng, request.Category,
		time.Now().UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func (r *SQLiteRepository) Get(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (Place, error) {
	place, err := scanPlace(r.db.QueryRowContext(ctx, `
		SELECT p.id, p.author_id, p.title, COALESCE(p.description, ''), p.lat, p.lng, p.category,
		p.created_at, p.updated_at,
		CASE WHEN p.rating_count > 0 THEN p.rating_avg ELSE NULL END,
		p.rating_count,
		(SELECT r.rating FROM ratings r WHERE r.place_id = p.id AND r.user_id = ?),
		(SELECT COUNT(*) FROM comments c WHERE c.place_id = p.id AND c.deleted = 0)
		FROM places p WHERE p.id = ?`, nullableUUID(viewerID), id.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Place{}, ErrNotFound
		}
		return Place{}, err
	}
	if err := loadPlaceChildren(ctx, r.db, &place); err != nil {
		return Place{}, err
	}
	return place, nil
}

func (r *SQLiteRepository) Update(ctx context.Context, id, authorID uuid.UUID, request UpdatePlaceRequest) (Place, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Place{}, err
	}
	defer tx.Rollback()
	var storedAuthor string
	if err := tx.QueryRowContext(ctx, `SELECT author_id FROM places WHERE id = ?`, id.String()).Scan(&storedAuthor); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Place{}, ErrNotFound
		}
		return Place{}, err
	}
	if storedAuthor != authorID.String() {
		return Place{}, ErrForbidden
	}
	current, err := r.getForUpdate(ctx, tx, id)
	if err != nil {
		return Place{}, err
	}
	if request.Title != nil {
		current.Title = *request.Title
	}
	if request.Description != nil {
		current.Description = *request.Description
	}
	if request.Lat != nil {
		current.Lat = *request.Lat
	}
	if request.Lng != nil {
		current.Lng = *request.Lng
	}
	if request.Category != nil {
		current.Category = *request.Category
	}
	_, err = tx.ExecContext(ctx, `UPDATE places SET title = ?, description = ?, lat = ?, lng = ?, category = ?, updated_at = ? WHERE id = ?`,
		current.Title, nullableText(current.Description), current.Lat, current.Lng, current.Category,
		time.Now().UTC().Format(time.RFC3339Nano), id.String())
	if err != nil {
		return Place{}, err
	}
	if request.Tags != nil {
		if err := replaceTags(ctx, tx, id, *request.Tags); err != nil {
			return Place{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Place{}, err
	}
	return r.Get(ctx, id, nil)
}

func (r *SQLiteRepository) getForUpdate(ctx context.Context, tx *sql.Tx, id uuid.UUID) (Place, error) {
	return scanPlace(tx.QueryRowContext(ctx, `SELECT p.id, p.author_id, p.title, COALESCE(p.description, ''),
		p.lat, p.lng, p.category, p.created_at, p.updated_at, NULL, 0, NULL, 0 FROM places p WHERE p.id = ?`, id.String()))
}

func (r *SQLiteRepository) Delete(ctx context.Context, id, authorID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM places WHERE id = ? AND author_id = ?`, id.String(), authorID.String())
	if err != nil {
		return err
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if deleted != 0 {
		return nil
	}
	var exists int
	if err := r.db.QueryRowContext(ctx, `SELECT 1 FROM places WHERE id = ?`, id.String()).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	return ErrForbidden
}

func (r *SQLiteRepository) List(ctx context.Context, options ListOptions) (PlaceList, error) {
	query := `SELECT p.id, p.title, p.lat, p.lng, p.category,
		CASE WHEN p.rating_count > 0 THEN p.rating_avg ELSE NULL END,
		p.rating_count,
		(SELECT ph.thumb_url FROM photos ph WHERE ph.place_id = p.id ORDER BY COALESCE(ph.position, 2147483647), ph.created_at, ph.id LIMIT 1),
		p.created_at
		FROM places p WHERE 1 = 1`
	args := make([]any, 0)
	if options.AuthorID != nil {
		query += ` AND p.author_id = ?`
		args = append(args, options.AuthorID.String())
	}
	if options.MinLat != nil && options.MaxLat != nil {
		query += ` AND p.lat BETWEEN ? AND ?`
		args = append(args, *options.MinLat, *options.MaxLat)
	}
	if options.MinLng != nil && options.MaxLng != nil {
		if *options.MinLng <= *options.MaxLng {
			query += ` AND p.lng BETWEEN ? AND ?`
			args = append(args, *options.MinLng, *options.MaxLng)
		} else {
			query += ` AND (p.lng >= ? OR p.lng <= ?)`
			args = append(args, *options.MinLng, *options.MaxLng)
		}
	}
	if options.Lat != nil && options.Lng != nil {
		latDelta := float64(options.RadiusM) / 111320
		minLat := math.Max(-90, *options.Lat-latDelta)
		maxLat := math.Min(90, *options.Lat+latDelta)
		query += ` AND p.lat BETWEEN ? AND ?`
		args = append(args, minLat, maxLat)
		cosLat := math.Cos(*options.Lat * math.Pi / 180)
		lngDelta := 180.0
		if math.Abs(cosLat) > 1e-9 {
			lngDelta = math.Min(180, float64(options.RadiusM)/(111320*math.Abs(cosLat)))
		}
		if lngDelta < 180 {
			minLng, maxLng := *options.Lng-lngDelta, *options.Lng+lngDelta
			if minLng < -180 {
				query += ` AND (p.lng >= ? OR p.lng <= ?)`
				args = append(args, minLng+360, maxLng)
			} else if maxLng > 180 {
				query += ` AND (p.lng >= ? OR p.lng <= ?)`
				args = append(args, minLng, maxLng-360)
			} else {
				query += ` AND p.lng BETWEEN ? AND ?`
				args = append(args, minLng, maxLng)
			}
		}
	}
	if options.Category != "" {
		query += ` AND p.category = ?`
		args = append(args, options.Category)
	}
	if len(options.Tags) != 0 {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(options.Tags)), ",")
		query += ` AND (SELECT COUNT(DISTINCT t.slug) FROM place_tags pt JOIN tags t ON t.id = pt.tag_id WHERE pt.place_id = p.id AND t.slug IN (` + placeholders + `)) = ?`
		for _, slug := range options.Tags {
			args = append(args, slug)
		}
		args = append(args, len(options.Tags))
	}
	if options.MinRating != nil {
		query += ` AND COALESCE(p.rating_avg, 0) >= ?`
		args = append(args, *options.MinRating)
	}
	if options.Query != "" {
		query += ` AND (LOWER(p.title) LIKE ? OR LOWER(COALESCE(p.description, '')) LIKE ?)`
		search := "%" + strings.ToLower(options.Query) + "%"
		args = append(args, search, search)
	}
	radius := options.Lat != nil && options.Lng != nil

	var cursor *cursorKey
	if options.Cursor != "" {
		decoded, err := decodeCursor(options.Cursor, options.Sort)
		if err != nil {
			return PlaceList{}, ErrInvalidCursor
		}
		cursor = &decoded
	}
	if cursor != nil && options.Sort != "distance" {
		createdAt := formatTime(cursor.CreatedAt)
		if options.Sort == "rating" {
			query += ` AND (COALESCE(p.rating_avg, 0) < ? OR (COALESCE(p.rating_avg, 0) = ? AND (p.created_at < ? OR (p.created_at = ? AND p.id < ?))))`
			args = append(args, cursor.Value, cursor.Value, createdAt, createdAt, cursor.ID.String())
		} else {
			query += ` AND (p.created_at < ? OR (p.created_at = ? AND p.id < ?))`
			args = append(args, createdAt, createdAt, cursor.ID.String())
		}
	}
	if options.Sort == "rating" {
		query += ` ORDER BY COALESCE(p.rating_avg, 0) DESC, p.created_at DESC, p.id DESC`
	} else {
		query += ` ORDER BY p.created_at DESC, p.id DESC`
	}
	// Con radio exacta hay que descartar filas fuera del círculo en Go; el
	// resto de consultas puede paginar en SQL.
	limitInSQL := options.Sort != "distance" && !radius
	if limitInSQL {
		query += ` LIMIT ?`
		args = append(args, options.Limit+1)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return PlaceList{}, err
	}
	defer rows.Close()
	items := make([]PlaceSummary, 0)
	for rows.Next() {
		var item PlaceSummary
		var rating sql.NullFloat64
		var cover sql.NullString
		var createdAt string
		if err := rows.Scan(&item.ID, &item.Title, &item.Lat, &item.Lng, &item.Category, &rating,
			&item.RatingCount, &cover, &createdAt); err != nil {
			return PlaceList{}, err
		}
		item.CreatedAt, err = parseTime(createdAt)
		if err != nil {
			return PlaceList{}, err
		}
		if rating.Valid {
			item.RatingAvg = &rating.Float64
		}
		if cover.Valid {
			item.CoverThumbURL = &cover.String
		}
		if radius {
			distance := haversineMeters(*options.Lat, *options.Lng, item.Lat, item.Lng)
			if distance > float64(options.RadiusM) {
				continue
			}
			distanceM := int(math.Round(distance))
			item.DistanceM = &distanceM
		}
		items = append(items, item)
		if !limitInSQL && options.Sort != "distance" && len(items) > options.Limit {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return PlaceList{}, err
	}

	if options.Sort == "distance" {
		sort.Slice(items, func(i, j int) bool { return summaryBefore(items[i], items[j], options.Sort) })
		if cursor != nil {
			filtered := items[:0]
			for _, item := range items {
				if summaryAfter(item, *cursor, options.Sort) {
					filtered = append(filtered, item)
				}
			}
			items = filtered
		}
	}

	result := PlaceList{Items: items}
	if len(items) > options.Limit {
		result.Items = items[:options.Limit]
		last := result.Items[len(result.Items)-1]
		key := cursorKey{Sort: options.Sort, CreatedAt: last.CreatedAt, ID: last.ID}
		switch options.Sort {
		case "rating":
			if last.RatingAvg != nil {
				key.Value = *last.RatingAvg
			}
		case "distance":
			if last.DistanceM != nil {
				key.Value = float64(*last.DistanceM)
			}
		}
		encoded, err := encodeCursor(key)
		if err != nil {
			return PlaceList{}, err
		}
		result.NextCursor = &encoded
	}
	return result, nil
}

func replaceTags(ctx context.Context, tx *sql.Tx, placeID uuid.UUID, slugs []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM place_tags WHERE place_id = ?`, placeID.String()); err != nil {
		return err
	}
	for _, slug := range slugs {
		var tagID string
		if err := tx.QueryRowContext(ctx, `SELECT id FROM tags WHERE slug = ?`, slug).Scan(&tagID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrInvalidTags
			}
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO place_tags (place_id, tag_id) VALUES (?, ?)`, placeID.String(), tagID); err != nil {
			return err
		}
	}
	return nil
}

func loadPlaceChildren(ctx context.Context, q queryer, place *Place) error {
	var authorName string
	var placesCount int
	if err := q.QueryRowContext(ctx, `SELECT name, (SELECT COUNT(*) FROM places WHERE author_id = ?) FROM users WHERE id = ?`,
		place.Author.ID.String(), place.Author.ID.String()).Scan(&authorName, &placesCount); err != nil {
		return err
	}
	place.Author.DisplayName = authorName
	place.Author.PlacesCount = placesCount

	tagRows, err := q.QueryContext(ctx, `SELECT t.slug, t.name FROM tags t JOIN place_tags pt ON pt.tag_id = t.id WHERE pt.place_id = ? ORDER BY t.name`, place.ID.String())
	if err != nil {
		return err
	}
	place.Tags = make([]Tag, 0)
	for tagRows.Next() {
		var tag Tag
		if err := tagRows.Scan(&tag.Slug, &tag.Label); err != nil {
			tagRows.Close()
			return err
		}
		place.Tags = append(place.Tags, tag)
	}
	if err := tagRows.Err(); err != nil {
		tagRows.Close()
		return err
	}
	tagRows.Close()

	photoRows, err := q.QueryContext(ctx, `SELECT id, url, thumb_url, caption, width, height, position, uploaded_by, created_at FROM photos WHERE place_id = ? ORDER BY COALESCE(position, 2147483647), created_at, id`, place.ID.String())
	if err != nil {
		return err
	}
	place.Photos = make([]Photo, 0)
	for photoRows.Next() {
		var photo Photo
		var id, uploadedBy, createdAt string
		var caption sql.NullString
		var width, height, position sql.NullInt64
		if err := photoRows.Scan(&id, &photo.URL, &photo.ThumbURL, &caption, &width, &height, &position, &uploadedBy, &createdAt); err != nil {
			photoRows.Close()
			return err
		}
		photo.ID, err = uuid.Parse(id)
		if err != nil {
			photoRows.Close()
			return err
		}
		photo.UploadedBy, err = uuid.Parse(uploadedBy)
		if err != nil {
			photoRows.Close()
			return err
		}
		photo.CreatedAt, err = parseTime(createdAt)
		if err != nil {
			photoRows.Close()
			return err
		}
		if caption.Valid {
			photo.Caption = &caption.String
		}
		if width.Valid {
			value := int(width.Int64)
			photo.Width = &value
		}
		if height.Valid {
			value := int(height.Int64)
			photo.Height = &value
		}
		if position.Valid {
			value := int(position.Int64)
			photo.Position = &value
		}
		place.Photos = append(place.Photos, photo)
	}
	if err := photoRows.Err(); err != nil {
		photoRows.Close()
		return err
	}
	return photoRows.Close()
}

func scanPlace(row *sql.Row) (Place, error) {
	var place Place
	var id, authorID, createdAt, updatedAt string
	var description sql.NullString
	var ratingAvg sql.NullFloat64
	var myRating sql.NullInt64
	if err := row.Scan(&id, &authorID, &place.Title, &description, &place.Lat, &place.Lng, &place.Category,
		&createdAt, &updatedAt, &ratingAvg, &place.RatingCount, &myRating, &place.CommentCount); err != nil {
		return Place{}, err
	}
	var err error
	place.ID, err = uuid.Parse(id)
	if err != nil {
		return Place{}, err
	}
	place.Author.ID, err = uuid.Parse(authorID)
	if err != nil {
		return Place{}, err
	}
	place.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return Place{}, err
	}
	place.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return Place{}, err
	}
	if description.Valid {
		place.Description = description.String
	}
	if ratingAvg.Valid {
		place.RatingAvg = &ratingAvg.Float64
	}
	if myRating.Valid {
		value := int(myRating.Int64)
		place.MyRating = &value
	}
	place.Tags = make([]Tag, 0)
	place.Photos = make([]Photo, 0)
	return place, nil
}

func summaryBefore(a, b PlaceSummary, sortMode string) bool {
	switch sortMode {
	case "rating":
		av, bv := ratingValue(a), ratingValue(b)
		if av != bv {
			return av > bv
		}
	case "distance":
		av, bv := distanceValue(a), distanceValue(b)
		if av != bv {
			return av < bv
		}
	}
	if !a.CreatedAt.Equal(b.CreatedAt) {
		return a.CreatedAt.After(b.CreatedAt)
	}
	return a.ID.String() > b.ID.String()
}

func summaryAfter(item PlaceSummary, key cursorKey, sortMode string) bool {
	switch sortMode {
	case "rating":
		value := ratingValue(item)
		if value != key.Value {
			return value < key.Value
		}
	case "distance":
		value := distanceValue(item)
		if value != key.Value {
			return value > key.Value
		}
	}
	if !item.CreatedAt.Equal(key.CreatedAt) {
		return item.CreatedAt.Before(key.CreatedAt)
	}
	return item.ID.String() < key.ID.String()
}

func ratingValue(item PlaceSummary) float64 {
	if item.RatingAvg == nil {
		return 0
	}
	return *item.RatingAvg
}

func distanceValue(item PlaceSummary) float64 {
	if item.DistanceM == nil {
		return 0
	}
	return float64(*item.DistanceM)
}

func encodeCursor(key cursorKey) (string, error) {
	data, err := json.Marshal(key)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeCursor(value, sortMode string) (cursorKey, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return cursorKey{}, err
	}
	var key cursorKey
	if err := json.Unmarshal(data, &key); err != nil {
		return cursorKey{}, err
	}
	if key.Sort != sortMode || key.ID == uuid.Nil || key.CreatedAt.IsZero() {
		return cursorKey{}, ErrInvalidCursor
	}
	return key, nil
}

func inClause(query string, values []string) (string, []any) {
	placeholders := strings.TrimRight(strings.Repeat("?,", len(values)), ",")
	args := make([]any, len(values))
	for i := range values {
		args[i] = values[i]
	}
	return fmt.Sprintf(query, placeholders), args
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableUUID(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02 15:04:05Z07:00"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid stored timestamp %q", value)
}

func haversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371000
	lat1, lat2 = lat1*math.Pi/180, lat2*math.Pi/180
	dLat, dLng := lat2-lat1, (lng2-lng1)*math.Pi/180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
