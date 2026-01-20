package db

import "context"

func ChangeStatus(pname string, status string) {
	query := "update projects set status=? where pname=?"
	ExecuteQuery(context.Background(), query, status, pname)
}
