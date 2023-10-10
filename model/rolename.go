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
	_, err := createRoleName(pool, ctx, o.Id, o.Name)
	return err
}

func ReadRoleName(pool RowQueryerContext, ctx context.Context, Id uint64) (RoleName, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var IdTemp uint64
	var NameTemp string
	err := pool.QueryRowContext(ctx, "select o.id, o.name from role_names as o where o.id = $1;", Id).Scan(&IdTemp, &NameTemp)
	return MakeRoleName(IdTemp, NameTemp), err
}

func (o RoleName) Update(pool ExecerContext, ctx context.Context) error {
	_, err := updateRoleName(pool, ctx, o.Id, o.Name)
	return err
}

func (o RoleName) Delete(pool ExecerContext, ctx context.Context) error {
	_, err := deleteRoleName(pool, ctx, o.Id)
	return err
}

func createRoleName(pool ExecerContext, ctx context.Context, Id uint64, Name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := pool.ExecContext(ctx, "insert into role_names(id, name) values($1, $2);", Id, Name)
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func updateRoleName(pool ExecerContext, ctx context.Context, Id uint64, Name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := pool.ExecContext(ctx, "update role_names set name = $2 where id = $1;", Id, Name)
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func deleteRoleName(pool ExecerContext, ctx context.Context, Id uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := pool.ExecContext(ctx, "delete from role_names where id = $1;", Id)
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func GetRoleNameByName(pool RowQueryerContext, ctx context.Context, name string) (RoleName, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var IdTemp uint64
	var NameTemp string
	err := pool.QueryRowContext(ctx, "select n.id, n.name from role_names as n where n.name = $1;", name).Scan(&IdTemp, &NameTemp)
	return MakeRoleName(IdTemp, NameTemp), err
}

func GetRoleNamesByIds(pool QueryerContext, ctx context.Context, ids []uint64) ([]RoleName, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var IdTemp uint64
	var NameTemp string
	rows, err := pool.QueryContext(ctx, "select n.id, n.name from role_names as n where n.id in ($1);", ids)
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

	result, err := pool.ExecContext(ctx, "delete from role_names where id not in (select distinct(name_id) from roles);")
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}
