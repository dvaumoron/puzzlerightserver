// Generated from model.crn - do not edit.

package model

import "context"

type Role struct {
	Id          uint64
	NameId      uint64
	ObjectId    uint64
	ActionFlags uint8
}

func MakeRole(id uint64, nameId uint64, objectId uint64, actionFlags uint8) Role {
	return Role{
		ActionFlags: actionFlags,
		Id:          id,
		NameId:      nameId,
		ObjectId:    objectId,
	}
}

func (o Role) Create(pool ExecerContext, ctx context.Context) error {
	_, err := createRole(pool, ctx, o.NameId, o.ObjectId, o.ActionFlags)
	return err
}

func ReadRole(pool RowQueryerContext, ctx context.Context, id uint64) (Role, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "select o.id, o.name_id, o.object_id, o.action_flags from roles as o where o.id = $1;"
	var idTemp uint64
	var nameIdTemp uint64
	var objectIdTemp uint64
	var actionFlagsTemp uint8
	err := pool.QueryRowContext(ctx, query, id).Scan(&idTemp, &nameIdTemp, &objectIdTemp, &actionFlagsTemp)
	return MakeRole(idTemp, nameIdTemp, objectIdTemp, actionFlagsTemp), err
}

func (o Role) Update(pool ExecerContext, ctx context.Context) error {
	_, err := updateRole(pool, ctx, o.Id, o.NameId, o.ObjectId, o.ActionFlags)
	return err
}

func (o Role) Delete(pool ExecerContext, ctx context.Context) error {
	_, err := deleteRole(pool, ctx, o.Id)
	return err
}

func createRole(pool ExecerContext, ctx context.Context, nameId uint64, objectId uint64, actionFlags uint8) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "insert into roles(name_id, object_id, action_flags) values($1, $2, $3);"
	result, err := pool.ExecContext(ctx, query, nameId, objectId, actionFlags)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func updateRole(pool ExecerContext, ctx context.Context, id uint64, nameId uint64, objectId uint64, actionFlags uint8) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "update roles set name_id = $2, object_id = $3, action_flags = $4 where id = $1;"
	result, err := pool.ExecContext(ctx, query, id, nameId, objectId, actionFlags)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func deleteRole(pool ExecerContext, ctx context.Context, id uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "delete from roles where id = $1;"
	result, err := pool.ExecContext(ctx, query, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func GetRolesByUserId(pool QueryerContext, ctx context.Context, userId uint64) ([]Role, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "select r.id, r.name_id, r.object_id, r.action_flags from roles as r where r.id in (select o.role_id from user_roles as o where o.user_id = $1);"
	var idTemp uint64
	var nameIdTemp uint64
	var objectIdTemp uint64
	var actionFlagsTemp uint8
	rows, err := pool.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []Role{}
	for rows.Next() {
		err := rows.Scan(&idTemp, &nameIdTemp, &objectIdTemp, &actionFlagsTemp)
		if err != nil {
			return nil, err
		}
		results = append(results, MakeRole(idTemp, nameIdTemp, objectIdTemp, actionFlagsTemp))
	}
	return results, nil
}

func GetRolesByObjectIds(pool QueryerContext, ctx context.Context, objectIds []uint64) ([]Role, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := varArgsFilter("select r.id, r.name_id, r.object_id, r.action_flags from roles as r where r.object_id in ($1);", "$1", len(objectIds))
	var idTemp uint64
	var nameIdTemp uint64
	var objectIdTemp uint64
	var actionFlagsTemp uint8
	rows, err := pool.QueryContext(ctx, query, anyConverter(objectIds)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []Role{}
	for rows.Next() {
		err := rows.Scan(&idTemp, &nameIdTemp, &objectIdTemp, &actionFlagsTemp)
		if err != nil {
			return nil, err
		}
		results = append(results, MakeRole(idTemp, nameIdTemp, objectIdTemp, actionFlagsTemp))
	}
	return results, nil
}

func GetRoleByNameAndObjectId(pool RowQueryerContext, ctx context.Context, name string, objectId uint64) (Role, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "select r.id, r.name_id, r.object_id, r.action_flags from roles as r, role_names as n where r.name_id = n.id and n.name = $1 and r.object_id = $2;"
	var idTemp uint64
	var nameIdTemp uint64
	var objectIdTemp uint64
	var actionFlagsTemp uint8
	err := pool.QueryRowContext(ctx, query, name, objectId).Scan(&idTemp, &nameIdTemp, &objectIdTemp, &actionFlagsTemp)
	return MakeRole(idTemp, nameIdTemp, objectIdTemp, actionFlagsTemp), err
}

func GetRoleByNameIdAndObjectId(pool RowQueryerContext, ctx context.Context, nameId uint64, objectId uint64) (Role, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "select r.id, r.name_id, r.object_id, r.action_flags from roles as r where r.name_id = $1 and r.object_id = $2;"
	var idTemp uint64
	var nameIdTemp uint64
	var objectIdTemp uint64
	var actionFlagsTemp uint8
	err := pool.QueryRowContext(ctx, query, nameId, objectId).Scan(&idTemp, &nameIdTemp, &objectIdTemp, &actionFlagsTemp)
	return MakeRole(idTemp, nameIdTemp, objectIdTemp, actionFlagsTemp), err
}

func GetRolesByNameAndObjectIds(pool QueryerContext, ctx context.Context, name string, objectIds []uint64) ([]Role, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	size := len(objectIds)
	queryArgs := make([]any, 0, size)
	queryArgs = append(queryArgs, name)
	queryArgs = append(queryArgs, anyConverter(objectIds)...)
	query := varArgsFilter("select r.id, r.name_id, r.object_id, r.action_flags from roles as r, role_names as n where r.name_id = n.id and n.name = $1 and r.object_id in ($2);", "$2", size)
	var idTemp uint64
	var nameIdTemp uint64
	var objectIdTemp uint64
	var actionFlagsTemp uint8
	rows, err := pool.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []Role{}
	for rows.Next() {
		err := rows.Scan(&idTemp, &nameIdTemp, &objectIdTemp, &actionFlagsTemp)
		if err != nil {
			return nil, err
		}
		results = append(results, MakeRole(idTemp, nameIdTemp, objectIdTemp, actionFlagsTemp))
	}
	return results, nil
}
