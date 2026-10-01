// Package audit records critical operations for later review.
package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/ids"
)

func Log(ctx context.Context, db database.DBTX, actor *uuid.UUID, action, entityType string, entityID uuid.UUID, meta map[string]any) error {
	if meta == nil {
		meta = map[string]any{}
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("audit: marshal: %w", err)
	}
	_, err = db.Exec(ctx, `
		INSERT INTO audit_logs (id, actor_user_id, action, entity_type, entity_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)`, ids.New(), actor, action, entityType, entityID, raw)
	if err != nil {
		return fmt.Errorf("audit: insert: %w", err)
	}
	return nil
}
