// Generated from model.crn - do not edit.

package model

import "context"

type RoleName struct {
	Id   uint64
	Name string
}

func MakeRoleName(id uint64, name string) RoleName {
	return RoleName{
		Id:   id,
		Name: name,
	}
}

func (o RoleName) Create(pool ExecerContext, ctx context.Context) error {
	_, err := createRoleName(pool, ctx, o.Name)
	return err
}

func ReadRoleName(pool RowQueryerContext, ctx context.Context, id uint64) (RoleName, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "select o.id, o.name from role_names as o where o.id = $1;"
	var idTemp uint64
	var nameTemp string
	err := pool.QueryRowContext(ctx, query, id).Scan(&idTemp, &nameTemp)
	return MakeRoleName(idTemp, nameTemp), err
}

func (o RoleName) Update(pool ExecerContext, ctx context.Context) error {
	_, err := updateRoleName(pool, ctx, o.Id, o.Name)
	return err
}

func (o RoleName) Delete(pool ExecerContext, ctx context.Context) error {
	_, err := deleteRoleName(pool, ctx, o.Id)
	return err
}

func createRoleName(pool ExecerContext, ctx context.Context, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "insert into role_names(name) values($1);"
	result, err := pool.ExecContext(ctx, query, name)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func updateRoleName(pool ExecerContext, ctx context.Context, id uint64, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "update role_names set name = $2 where id = $1;"
	result, err := pool.ExecContext(ctx, query, id, name)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func deleteRoleName(pool ExecerContext, ctx context.Context, id uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "delete from role_names where id = $1;"
	result, err := pool.ExecContext(ctx, query, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func GetRoleNameByName(pool RowQueryerContext, ctx context.Context, name string) (RoleName, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "select n.id, n.name from role_names as n where n.name = $1;"
	var IdTemp uint64
	var NameTemp string
	err := pool.QueryRowContext(ctx, query, name).Scan(&IdTemp, &NameTemp)
	return MakeRoleName(IdTemp, NameTemp), err
}

func GetRoleNamesByIds(pool QueryerContext, ctx context.Context, ids []uint64) ([]RoleName, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := varArgsFilter("select n.id, n.name from role_names as n where n.id in ($1);", "$1", len(ids))
	var IdTemp uint64
	var NameTemp string
	rows, err := pool.QueryContext(ctx, query, anyConverter(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []RoleName{}
	for rows.Next() {
		err := rows.Scan(&IdTemp, &NameTemp)
		if err != nil {
			return nil, err
		}
		results = append(results, MakeRoleName(IdTemp, NameTemp))
	}
	return results, nil
}

func DeleteUnusedRoleNames(pool ExecerContext, ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "delete from role_names where id not in (select distinct(name_id) from roles);"
	result, err := pool.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
