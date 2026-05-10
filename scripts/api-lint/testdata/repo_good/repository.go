package fixture

import "context"

var authz = struct {
	Require    func(any)
	RequireAll func(any)
}{}

type txer interface {
	ExecContext(context.Context, string, ...any) (any, error)
}

func SaveWidget(ctx context.Context, tx txer, area string, id string) error {
	authz.Require(area)
	_, err := tx.ExecContext(ctx, "UPDATE widgets SET area_code = $1 WHERE id = $2", area, id)
	return err
}
