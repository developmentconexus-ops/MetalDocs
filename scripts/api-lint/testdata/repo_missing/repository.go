package fixture

import "context"

type txer interface {
	ExecContext(context.Context, string, ...any) (any, error)
}

func SaveWidget(ctx context.Context, tx txer, area string, id string) error {
	_, err := tx.ExecContext(ctx, "UPDATE widgets SET area_code = $1 WHERE id = $2", area, id)
	_ = area
	return err
}
