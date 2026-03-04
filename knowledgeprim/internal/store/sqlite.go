package store

import (
	"context"
	"database/sql"
	"embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/propifly/primkit/primkit/db"
)

//go:embed migrations/*.sql
var migrations embed.FS

// SQLiteStore implements the Store interface backed by SQLite.
type SQLiteStore struct {
	db     *sql.DB
	dbPath string
}

// New creates a new SQLiteStore from a file path.
func New(dbPath string) (*SQLiteStore, error) {
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening store: %w", err)
	}
	if err := db.Migrate(database, migrations, "migrations"); err != nil {
		database.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return &SQLiteStore{db: database, dbPath: dbPath}, nil
}

// NewFromDB creates a SQLiteStore from an existing *sql.DB (for tests).
func NewFromDB(database *sql.DB) (*SQLiteStore, error) {
	if err := db.Migrate(database, migrations, "migrations"); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return &SQLiteStore{db: database, dbPath: ":memory:"}, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// ---------------------------------------------------------------------------
// Entity CRUD
// ---------------------------------------------------------------------------

func (s *SQLiteStore) CaptureEntity(ctx context.Context, entity *model.Entity, embedding []float32) error {
	if err := entity.Validate(); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidEntity, err)
	}

	if entity.ID == "" {
		entity.ID = model.NewEntityID()
	}
	now := time.Now().UTC()
	entity.CreatedAt = now
	entity.UpdatedAt = now

	var props *string
	if len(entity.Properties) > 0 {
		p := string(entity.Properties)
		props = &p
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO entities (id, type, title, body, url, source, properties, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entity.ID, entity.Type, entity.Title, entity.Body, entity.URL, entity.Source,
		props, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("inserting entity: %w", err)
	}

	// Store embedding if provided.
	if len(embedding) > 0 {
		blob := float32sToBytes(embedding)
		_, err = tx.ExecContext(ctx,
			`INSERT INTO entity_vectors (entity_id, embedding, dimensions)
			 VALUES (?, ?, ?)`,
			entity.ID, blob, len(embedding),
		)
		if err != nil {
			return fmt.Errorf("inserting vector: %w", err)
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) GetEntity(ctx context.Context, id string) (*model.Entity, error) {
	entity, err := s.getEntity(ctx, s.db, id)
	if err != nil {
		return nil, err
	}

	// Load edges for this entity.
	edges, err := s.getEdgesForEntity(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("loading edges: %w", err)
	}
	entity.Edges = edges

	return entity, nil
}

func (s *SQLiteStore) getEntity(ctx context.Context, q querier, id string) (*model.Entity, error) {
	var e model.Entity
	var body, url, props sql.NullString
	var createdAt, updatedAt string

	err := q.QueryRowContext(ctx,
		`SELECT id, type, title, body, url, source, properties, created_at, updated_at
		 FROM entities WHERE id = ?`, id,
	).Scan(&e.ID, &e.Type, &e.Title, &body, &url, &e.Source, &props, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying entity: %w", err)
	}

	if body.Valid {
		e.Body = &body.String
	}
	if url.Valid {
		e.URL = &url.String
	}
	if props.Valid {
		e.Properties = json.RawMessage(props.String)
	}
	e.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	e.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)

	return &e, nil
}

func (s *SQLiteStore) getEdgesForEntity(ctx context.Context, q querier, id string) ([]*model.Edge, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT source_id, target_id, relationship, weight, context, created_at, updated_at
		 FROM edges WHERE source_id = ? OR target_id = ?
		 ORDER BY weight DESC, created_at DESC`, id, id,
	)
	if err != nil {
		return nil, fmt.Errorf("querying edges: %w", err)
	}
	defer rows.Close()

	var edges []*model.Edge
	for rows.Next() {
		var edge model.Edge
		var edgeCtx sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&edge.SourceID, &edge.TargetID, &edge.Relationship,
			&edge.Weight, &edgeCtx, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scanning edge: %w", err)
		}
		if edgeCtx.Valid {
			edge.Context = &edgeCtx.String
		}
		edge.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		edge.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
		edges = append(edges, &edge)
	}
	return edges, rows.Err()
}

func (s *SQLiteStore) UpdateEntity(ctx context.Context, id string, update *model.EntityUpdate) error {
	// Verify entity exists.
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM entities WHERE id = ?`, id).Scan(&exists)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("checking entity: %w", err)
	}

	// Build dynamic SET clause.
	var sets []string
	var args []interface{}

	if update.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *update.Title)
	}
	if update.Body != nil {
		sets = append(sets, "body = ?")
		args = append(args, *update.Body)
	}
	if update.URL != nil {
		sets = append(sets, "url = ?")
		args = append(args, *update.URL)
	}
	if update.Properties != nil {
		sets = append(sets, "properties = ?")
		args = append(args, string(*update.Properties))
	}

	if len(sets) == 0 {
		return nil // Nothing to update.
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	sets = append(sets, "updated_at = ?")
	args = append(args, now)
	args = append(args, id)

	query := fmt.Sprintf("UPDATE entities SET %s WHERE id = ?", strings.Join(sets, ", "))
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating entity: %w", err)
	}

	return nil
}

func (s *SQLiteStore) DeleteEntity(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM entities WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting entity: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// Edge operations
// ---------------------------------------------------------------------------

func (s *SQLiteStore) CreateEdge(ctx context.Context, edge *model.Edge) error {
	if err := edge.Validate(); err != nil {
		return err
	}

	if edge.Weight == 0 {
		edge.Weight = 1.0
	}
	now := time.Now().UTC()
	edge.CreatedAt = now
	edge.UpdatedAt = now

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO edges (source_id, target_id, relationship, weight, context, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		edge.SourceID, edge.TargetID, edge.Relationship, edge.Weight, edge.Context,
		now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrEdgeExists
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return fmt.Errorf("one or both entities do not exist")
		}
		return fmt.Errorf("inserting edge: %w", err)
	}

	return nil
}

func (s *SQLiteStore) UpdateEdge(ctx context.Context, source, target, relationship string, update *model.EdgeUpdate) error {
	var sets []string
	var args []interface{}

	if update.Context != nil {
		sets = append(sets, "context = ?")
		args = append(args, *update.Context)
	}
	if update.Weight != nil {
		sets = append(sets, "weight = ?")
		args = append(args, *update.Weight)
	}

	if len(sets) == 0 {
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	sets = append(sets, "updated_at = ?")
	args = append(args, now)
	args = append(args, source, target, relationship)

	query := fmt.Sprintf(
		"UPDATE edges SET %s WHERE source_id = ? AND target_id = ? AND relationship = ?",
		strings.Join(sets, ", "),
	)
	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating edge: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrEdgeNotFound
	}
	return nil
}

func (s *SQLiteStore) StrengthenEdge(ctx context.Context, source, target, relationship string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := s.db.ExecContext(ctx,
		`UPDATE edges SET weight = weight + 1.0, updated_at = ?
		 WHERE source_id = ? AND target_id = ? AND relationship = ?`,
		now, source, target, relationship,
	)
	if err != nil {
		return fmt.Errorf("strengthening edge: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrEdgeNotFound
	}
	return nil
}

func (s *SQLiteStore) DeleteEdge(ctx context.Context, source, target, relationship string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM edges WHERE source_id = ? AND target_id = ? AND relationship = ?`,
		source, target, relationship,
	)
	if err != nil {
		return fmt.Errorf("deleting edge: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrEdgeNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// Search: FTS5, Vector, Hybrid (RRF)
// ---------------------------------------------------------------------------

func (s *SQLiteStore) SearchFTS(ctx context.Context, query string, filter *model.SearchFilter) ([]*model.SearchResult, error) {
	limit := resolveLimit(filter)

	var whereClause string
	var args []interface{}
	args = append(args, query)

	if filter != nil && filter.Type != "" {
		whereClause = " AND e.type = ?"
		args = append(args, filter.Type)
	}
	args = append(args, limit)

	sqlQuery := fmt.Sprintf(`
		SELECT e.id, e.type, e.title, e.body, e.url, e.source, e.properties,
		       e.created_at, e.updated_at,
		       bm25(entities_fts, 1.0, 0.5) AS score
		FROM entities_fts f
		JOIN entities e ON e.rowid = f.rowid
		WHERE entities_fts MATCH ?%s
		ORDER BY score ASC
		LIMIT ?`, whereClause)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("FTS search: %w", err)
	}
	defer rows.Close()

	return scanSearchResults(rows)
}

func (s *SQLiteStore) SearchVector(ctx context.Context, embedding []float32, filter *model.SearchFilter) ([]*model.SearchResult, error) {
	if len(embedding) == 0 {
		return nil, fmt.Errorf("embedding is required for vector search")
	}

	limit := resolveLimit(filter)

	// Load all vectors and compute cosine similarity in Go.
	type vecEntry struct {
		entityID  string
		embedding []float32
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT entity_id, embedding FROM entity_vectors`)
	if err != nil {
		return nil, fmt.Errorf("loading vectors: %w", err)
	}
	defer rows.Close()

	type scored struct {
		entityID string
		distance float64
	}
	var entries []scored

	for rows.Next() {
		var entityID string
		var blob []byte
		if err := rows.Scan(&entityID, &blob); err != nil {
			return nil, fmt.Errorf("scanning vector: %w", err)
		}
		vec := bytesToFloat32s(blob)
		dist := cosineDistance(embedding, vec)
		entries = append(entries, scored{entityID: entityID, distance: dist})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating vectors: %w", err)
	}

	// Sort by distance ascending (lower = more similar).
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].distance < entries[j].distance
	})

	// Filter by type and limit.
	var results []*model.SearchResult
	for _, entry := range entries {
		if len(results) >= limit {
			break
		}
		entity, err := s.getEntity(ctx, s.db, entry.entityID)
		if err != nil {
			continue // Entity might have been deleted.
		}
		if filter != nil && filter.Type != "" && entity.Type != filter.Type {
			continue
		}
		results = append(results, &model.SearchResult{
			Entity: entity,
			Score:  1.0 - entry.distance, // Convert distance to similarity.
		})
	}

	return results, nil
}

func (s *SQLiteStore) SearchHybrid(ctx context.Context, query string, embedding []float32, filter *model.SearchFilter) ([]*model.SearchResult, error) {
	// If no embedding, fall back to FTS only.
	if len(embedding) == 0 {
		return s.SearchFTS(ctx, query, filter)
	}

	// Run both searches with a larger limit for RRF fusion.
	rrfFilter := &model.SearchFilter{
		Limit: resolveLimit(filter) * 3,
	}
	if filter != nil {
		rrfFilter.Type = filter.Type
	}

	ftsResults, err := s.SearchFTS(ctx, query, rrfFilter)
	if err != nil {
		return nil, fmt.Errorf("FTS search: %w", err)
	}

	vecResults, err := s.SearchVector(ctx, embedding, rrfFilter)
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}

	// Reciprocal Rank Fusion with k=60.
	const k = 60.0
	rrfScores := make(map[string]float64)
	entityMap := make(map[string]*model.Entity)

	for rank, r := range ftsResults {
		rrfScores[r.Entity.ID] += 1.0 / (k + float64(rank+1))
		entityMap[r.Entity.ID] = r.Entity
	}
	for rank, r := range vecResults {
		rrfScores[r.Entity.ID] += 1.0 / (k + float64(rank+1))
		entityMap[r.Entity.ID] = r.Entity
	}

	// Sort by RRF score descending.
	type rrfEntry struct {
		id    string
		score float64
	}
	var ranked []rrfEntry
	for id, score := range rrfScores {
		ranked = append(ranked, rrfEntry{id: id, score: score})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	// Take top N.
	limit := resolveLimit(filter)
	var results []*model.SearchResult
	for _, entry := range ranked {
		if len(results) >= limit {
			break
		}
		results = append(results, &model.SearchResult{
			Entity: entityMap[entry.id],
			Score:  entry.score,
		})
	}

	return results, nil
}

// ---------------------------------------------------------------------------
// Graph traversal
// ---------------------------------------------------------------------------

func (s *SQLiteStore) Related(ctx context.Context, id string, opts *model.TraversalOpts) ([]*model.TraversalResult, error) {
	// Verify entity exists.
	if _, err := s.getEntity(ctx, s.db, id); err != nil {
		return nil, err
	}

	depth := 1
	direction := "both"
	if opts != nil {
		if opts.Depth > 0 {
			depth = opts.Depth
		}
		if opts.Direction != "" {
			direction = opts.Direction
		}
	}

	// BFS traversal with cycle prevention.
	visited := map[string]bool{id: true}
	var results []*model.TraversalResult

	type queueItem struct {
		entityID string
		depth    int
	}
	queue := []queueItem{{entityID: id, depth: 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.depth >= depth {
			continue
		}

		// Find edges for current entity.
		edges, err := s.getEdgesForEntity(ctx, s.db, current.entityID)
		if err != nil {
			return nil, fmt.Errorf("traversal at depth %d: %w", current.depth, err)
		}

		for _, edge := range edges {
			var neighborID, dir string
			if edge.SourceID == current.entityID {
				neighborID = edge.TargetID
				dir = "outgoing"
			} else {
				neighborID = edge.SourceID
				dir = "incoming"
			}

			// Direction filter.
			if direction != "both" && dir != direction {
				continue
			}

			// Relationship filter.
			if opts != nil && opts.Relationship != "" && edge.Relationship != opts.Relationship {
				continue
			}

			// Weight filter.
			if opts != nil && opts.MinWeight > 0 && edge.Weight < opts.MinWeight {
				continue
			}

			if visited[neighborID] {
				continue
			}
			visited[neighborID] = true

			neighbor, err := s.getEntity(ctx, s.db, neighborID)
			if err != nil {
				continue // Might have been deleted.
			}

			results = append(results, &model.TraversalResult{
				Entity:       neighbor,
				Relationship: edge.Relationship,
				Direction:    dir,
				Depth:        current.depth + 1,
				Weight:       edge.Weight,
			})

			queue = append(queue, queueItem{entityID: neighborID, depth: current.depth + 1})
		}
	}

	return results, nil
}

// ---------------------------------------------------------------------------
// Discovery
// ---------------------------------------------------------------------------

func (s *SQLiteStore) Discover(ctx context.Context, opts *model.DiscoverOpts) (*model.DiscoverReport, error) {
	if opts == nil {
		opts = model.DiscoverAll()
	}

	report := &model.DiscoverReport{}

	if opts.Orphans {
		orphans, err := s.discoverOrphans(ctx)
		if err != nil {
			return nil, fmt.Errorf("discovering orphans: %w", err)
		}
		report.Orphans = orphans
	}

	if opts.WeakEdges {
		weak, err := s.discoverWeakEdges(ctx)
		if err != nil {
			return nil, fmt.Errorf("discovering weak edges: %w", err)
		}
		report.WeakEdges = weak
	}

	if opts.Clusters {
		clusters, err := s.discoverClusters(ctx)
		if err != nil {
			return nil, fmt.Errorf("discovering clusters: %w", err)
		}
		report.Clusters = clusters
	}

	if opts.Bridges {
		bridges, err := s.discoverBridges(ctx)
		if err != nil {
			return nil, fmt.Errorf("discovering bridges: %w", err)
		}
		report.Bridges = bridges
	}

	if opts.Temporal {
		temporal, err := s.discoverTemporal(ctx)
		if err != nil {
			return nil, fmt.Errorf("discovering temporal: %w", err)
		}
		report.Temporal = temporal
	}

	return report, nil
}

func (s *SQLiteStore) discoverOrphans(ctx context.Context) ([]*model.Entity, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT e.id, e.type, e.title, e.body, e.url, e.source, e.properties,
		       e.created_at, e.updated_at
		FROM entities e
		WHERE e.id NOT IN (
			SELECT source_id FROM edges
			UNION
			SELECT target_id FROM edges
		)
		ORDER BY e.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEntities(rows)
}

func (s *SQLiteStore) discoverWeakEdges(ctx context.Context) ([]*model.Edge, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT source_id, target_id, relationship, weight, context, created_at, updated_at
		FROM edges
		WHERE context IS NULL OR context = ''
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []*model.Edge
	for rows.Next() {
		var edge model.Edge
		var edgeCtx sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&edge.SourceID, &edge.TargetID, &edge.Relationship,
			&edge.Weight, &edgeCtx, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if edgeCtx.Valid {
			edge.Context = &edgeCtx.String
		}
		edge.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		edge.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
		edges = append(edges, &edge)
	}
	return edges, rows.Err()
}

func (s *SQLiteStore) discoverClusters(ctx context.Context) ([]model.Cluster, error) {
	// Load adjacency list.
	neighbors := make(map[string]map[string]bool)

	rows, err := s.db.QueryContext(ctx, `SELECT source_id, target_id FROM edges`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var src, tgt string
		if err := rows.Scan(&src, &tgt); err != nil {
			return nil, err
		}
		if neighbors[src] == nil {
			neighbors[src] = make(map[string]bool)
		}
		if neighbors[tgt] == nil {
			neighbors[tgt] = make(map[string]bool)
		}
		neighbors[src][tgt] = true
		neighbors[tgt][src] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Connected components via BFS.
	visited := make(map[string]bool)
	var clusters []model.Cluster

	for nodeID := range neighbors {
		if visited[nodeID] {
			continue
		}
		// BFS from this node.
		var component []string
		queue := []string{nodeID}
		visited[nodeID] = true

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]
			component = append(component, curr)

			for nb := range neighbors[curr] {
				if !visited[nb] {
					visited[nb] = true
					queue = append(queue, nb)
				}
			}
		}

		// Only include clusters with 3+ entities.
		if len(component) >= 3 {
			var entities []*model.Entity
			for _, eid := range component {
				e, err := s.getEntity(ctx, s.db, eid)
				if err != nil {
					continue
				}
				entities = append(entities, e)
			}
			if len(entities) >= 3 {
				clusters = append(clusters, model.Cluster{
					Entities: entities,
					Size:     len(entities),
				})
			}
		}
	}

	// Sort clusters by size descending.
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Size > clusters[j].Size
	})

	return clusters, nil
}

func (s *SQLiteStore) discoverBridges(ctx context.Context) ([]*model.BridgeEntity, error) {
	// Entities with high edge count relative to average.
	rows, err := s.db.QueryContext(ctx, `
		SELECT entity_id, cnt FROM (
			SELECT source_id AS entity_id, COUNT(*) AS cnt FROM edges GROUP BY source_id
			UNION ALL
			SELECT target_id AS entity_id, COUNT(*) AS cnt FROM edges GROUP BY target_id
		)
		GROUP BY entity_id
		HAVING SUM(cnt) >= 3
		ORDER BY SUM(cnt) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bridges []*model.BridgeEntity
	for rows.Next() {
		var entityID string
		var count int
		if err := rows.Scan(&entityID, &count); err != nil {
			return nil, err
		}
		entity, err := s.getEntity(ctx, s.db, entityID)
		if err != nil {
			continue
		}
		bridges = append(bridges, &model.BridgeEntity{
			Entity:    entity,
			EdgeCount: count,
		})
	}
	return bridges, rows.Err()
}

func (s *SQLiteStore) discoverTemporal(ctx context.Context) ([]model.TemporalGroup, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT strftime('%Y-%W', created_at) AS period, type, COUNT(*) AS cnt
		FROM entities
		GROUP BY period, type
		ORDER BY period DESC, cnt DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []model.TemporalGroup
	for rows.Next() {
		var g model.TemporalGroup
		if err := rows.Scan(&g.Period, &g.Type, &g.Count); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// ---------------------------------------------------------------------------
// Aggregates
// ---------------------------------------------------------------------------

func (s *SQLiteStore) ListTypes(ctx context.Context) ([]model.TypeCount, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT type, COUNT(*) AS cnt FROM entities GROUP BY type ORDER BY cnt DESC`)
	if err != nil {
		return nil, fmt.Errorf("listing types: %w", err)
	}
	defer rows.Close()

	var types []model.TypeCount
	for rows.Next() {
		var tc model.TypeCount
		if err := rows.Scan(&tc.Type, &tc.Count); err != nil {
			return nil, err
		}
		types = append(types, tc)
	}
	return types, rows.Err()
}

func (s *SQLiteStore) ListRelationships(ctx context.Context) ([]model.RelationshipCount, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT relationship, COUNT(*) AS cnt FROM edges GROUP BY relationship ORDER BY cnt DESC`)
	if err != nil {
		return nil, fmt.Errorf("listing relationships: %w", err)
	}
	defer rows.Close()

	var rels []model.RelationshipCount
	for rows.Next() {
		var rc model.RelationshipCount
		if err := rows.Scan(&rc.Relationship, &rc.Count); err != nil {
			return nil, err
		}
		rels = append(rels, rc)
	}
	return rels, rows.Err()
}

func (s *SQLiteStore) Stats(ctx context.Context) (*model.Stats, error) {
	stats := &model.Stats{DBPath: s.dbPath}

	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM entities`).Scan(&stats.EntityCount)
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM edges`).Scan(&stats.EdgeCount)
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM entity_vectors`).Scan(&stats.VectorCount)
	s.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT type) FROM entities`).Scan(&stats.TypeCount)

	// Orphan count.
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM entities
		WHERE id NOT IN (
			SELECT source_id FROM edges
			UNION
			SELECT target_id FROM edges
		)`).Scan(&stats.OrphanCount)

	// DB size (only works for file-based DBs).
	if s.dbPath != "" && s.dbPath != ":memory:" {
		var pageCount, pageSize int64
		s.db.QueryRowContext(ctx, `PRAGMA page_count`).Scan(&pageCount)
		s.db.QueryRowContext(ctx, `PRAGMA page_size`).Scan(&pageSize)
		stats.DBSize = pageCount * pageSize
	}

	return stats, nil
}

// ---------------------------------------------------------------------------
// Export / Import
// ---------------------------------------------------------------------------

func (s *SQLiteStore) ExportEntities(ctx context.Context, filter *model.ExportFilter) (*model.ExportData, error) {
	var whereClause string
	var args []interface{}
	if filter != nil && filter.Type != "" {
		whereClause = " WHERE type = ?"
		args = append(args, filter.Type)
	}

	query := fmt.Sprintf(`SELECT id, type, title, body, url, source, properties, created_at, updated_at
		FROM entities%s ORDER BY created_at`, whereClause)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exporting entities: %w", err)
	}
	defer rows.Close()

	entities, err := scanEntities(rows)
	if err != nil {
		return nil, err
	}

	// Export edges.
	edgeRows, err := s.db.QueryContext(ctx,
		`SELECT source_id, target_id, relationship, weight, context, created_at, updated_at
		 FROM edges ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("exporting edges: %w", err)
	}
	defer edgeRows.Close()

	var edges []*model.Edge
	for edgeRows.Next() {
		var edge model.Edge
		var edgeCtx sql.NullString
		var createdAt, updatedAt string
		if err := edgeRows.Scan(&edge.SourceID, &edge.TargetID, &edge.Relationship,
			&edge.Weight, &edgeCtx, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if edgeCtx.Valid {
			edge.Context = &edgeCtx.String
		}
		edge.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		edge.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
		edges = append(edges, &edge)
	}

	return &model.ExportData{
		Entities: entities,
		Edges:    edges,
	}, edgeRows.Err()
}

func (s *SQLiteStore) ImportEntities(ctx context.Context, data *model.ExportData) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, e := range data.Entities {
		var props *string
		if len(e.Properties) > 0 {
			p := string(e.Properties)
			props = &p
		}

		_, err := tx.ExecContext(ctx,
			`INSERT OR REPLACE INTO entities (id, type, title, body, url, source, properties, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			e.ID, e.Type, e.Title, e.Body, e.URL, e.Source, props,
			e.CreatedAt.Format(time.RFC3339Nano), e.UpdatedAt.Format(time.RFC3339Nano),
		)
		if err != nil {
			return fmt.Errorf("importing entity %s: %w", e.ID, err)
		}
	}

	for _, edge := range data.Edges {
		_, err := tx.ExecContext(ctx,
			`INSERT OR REPLACE INTO edges (source_id, target_id, relationship, weight, context, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			edge.SourceID, edge.TargetID, edge.Relationship, edge.Weight, edge.Context,
			edge.CreatedAt.Format(time.RFC3339Nano), edge.UpdatedAt.Format(time.RFC3339Nano),
		)
		if err != nil {
			return fmt.Errorf("importing edge %s→%s: %w", edge.SourceID, edge.TargetID, err)
		}
	}

	return tx.Commit()
}

// ---------------------------------------------------------------------------
// Vector helpers
// ---------------------------------------------------------------------------

// float32sToBytes converts a slice of float32 to a byte slice (little-endian).
func float32sToBytes(floats []float32) []byte {
	buf := make([]byte, len(floats)*4)
	for i, f := range floats {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
	}
	return buf
}

// bytesToFloat32s converts a byte slice back to float32 (little-endian).
func bytesToFloat32s(data []byte) []float32 {
	floats := make([]float32, len(data)/4)
	for i := range floats {
		floats[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:]))
	}
	return floats
}

// cosineDistance computes the cosine distance between two vectors.
// Returns 0 for identical vectors, 2 for opposite vectors.
func cosineDistance(a, b []float32) float64 {
	if len(a) != len(b) {
		return 2.0 // Maximum distance for mismatched dimensions.
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 2.0
	}

	similarity := dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
	return 1.0 - similarity
}

// ---------------------------------------------------------------------------
// Scan helpers
// ---------------------------------------------------------------------------

// querier abstracts *sql.DB and *sql.Tx for shared query methods.
type querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func scanEntities(rows *sql.Rows) ([]*model.Entity, error) {
	var entities []*model.Entity
	for rows.Next() {
		var e model.Entity
		var body, url, props sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&e.ID, &e.Type, &e.Title, &body, &url, &e.Source,
			&props, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scanning entity: %w", err)
		}
		if body.Valid {
			e.Body = &body.String
		}
		if url.Valid {
			e.URL = &url.String
		}
		if props.Valid {
			e.Properties = json.RawMessage(props.String)
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		e.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
		entities = append(entities, &e)
	}
	return entities, rows.Err()
}

func scanSearchResults(rows *sql.Rows) ([]*model.SearchResult, error) {
	var results []*model.SearchResult
	for rows.Next() {
		var e model.Entity
		var body, url, props sql.NullString
		var createdAt, updatedAt string
		var score float64
		if err := rows.Scan(&e.ID, &e.Type, &e.Title, &body, &url, &e.Source,
			&props, &createdAt, &updatedAt, &score); err != nil {
			return nil, fmt.Errorf("scanning search result: %w", err)
		}
		if body.Valid {
			e.Body = &body.String
		}
		if url.Valid {
			e.URL = &url.String
		}
		if props.Valid {
			e.Properties = json.RawMessage(props.String)
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		e.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
		results = append(results, &model.SearchResult{
			Entity: &e,
			Score:  math.Abs(score), // BM25 returns negative scores.
		})
	}
	return results, rows.Err()
}

func resolveLimit(filter *model.SearchFilter) int {
	if filter != nil && filter.Limit > 0 {
		return filter.Limit
	}
	return model.DefaultSearchLimit
}
