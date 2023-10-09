// Generated from model.crn - do not edit.

package model

import (
	"context"
	"database/sql"
)

type RoleName struct {
	ID   uint64
	Name string
}

func MakeRoleName(id uint64, name string) RoleName {
	return RoleName{
		ID:   id,
		Name: name,
	}
}

func (o RoleName) Create(pool ExecerContext, ctx context.Context) error {
	_, err := createRoleName(pool, ctx, o.ID, o.Name)
	return err
}

func ReadRoleName(pool RowQueryerContext, ctx context.Context, ID uint64) (RoleName, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var IDTemp uint64
	var NameTemp string
	err := pool.QueryRowContext(ctx, "select o.i_d, o.name from role_names as o where o.i_d = :ID;", sql.Named("ID", ID)).Scan(&IDTemp, &NameTemp)
	return MakeRoleName(IDTemp, NameTemp), err
}

func (o RoleName) Update(pool ExecerContext, ctx context.Context) error {
	_, err := updateRoleName(pool, ctx, o.ID, o.Name)
	return err
}

func (o RoleName) Delete(pool ExecerContext, ctx context.Context) error {
	_, err := deleteRoleName(pool, ctx, o.ID)
	return err
}

func createRoleName(pool ExecerContext, ctx context.Context, ID uint64, Name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := pool.ExecContext(ctx, "insert into role_names(i_d, name) values(:ID, :Name);", sql.Named("ID", ID), sql.Named("Name", Name))
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func updateRoleName(pool ExecerContext, ctx context.Context, ID uint64, Name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := pool.ExecContext(ctx, "update role_names set name = :Name where i_d = :ID;", sql.Named("ID", ID), sql.Named("Name", Name))
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func deleteRoleName(pool ExecerContext, ctx context.Context, ID uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := pool.ExecContext(ctx, "delete from role_names where i_d = :ID;", sql.Named("ID", ID))
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func GetRoleNameByName(pool RowQueryerContext, ctx context.Context, name string) (RoleName, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var IDTemp uint64
	var NameTemp string
	err := pool.QueryRowContext(ctx, "select n.i_d, n.name from role_names as n where n.name = :name;", sql.Named("name", name)).Scan(&IDTemp, &NameTemp)
	return MakeRoleName(IDTemp, NameTemp), err
}

func GetRoleNamesByIds(pool QueryerContext, ctx context.Context, ids []uint64) ([]RoleName, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var IDTemp uint64
	var NameTemp string
	rows, err := pool.QueryContext(ctx, "select n.i_d, n.name from role_names as n where n.id in (:ids);", sql.Named("ids", ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []RoleName{}
	for rows.Next() {
		err := rows.Scan(&IDTemp, &NameTemp)
		if err != nil {
			return nil, err
		}
		results = append(results, MakeRoleName(IDTemp, NameTemp))
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
