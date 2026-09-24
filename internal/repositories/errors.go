package repositories

import "github.com/georgysavva/scany/v2/sqlscan"

func NotFound(err error) bool {
	return sqlscan.NotFound(err)
}
