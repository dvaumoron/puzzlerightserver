// Generated from model.crn - do not edit.

package model

import (
	"context"
	"database/sql"
)

type Role struct {
	ID          uint64
	NameId      uint64
	ObjectId    uint64
	ActionFlags uint8
}

func MakeRole(id uint64, nameid uint64, objectid uint64, actionflags uint8) Role {
	return Role{
		ActionFlags: actionflags,
		ID:          id,
		NameId:      nameid,
		ObjectId:    objectid,
	}
}

func (o Role) Create(pool ExecerContext, ctx context.Context) error {
	_, err := createRole(pool, ctx, o.ID, o.NameId, o.ObjectId, o.ActionFlags)
	return err
}

func ReadRole(pool RowQueryerContext, ctx context.Context, ID uint64) (Role, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var IDTemp uint64
	var NameIdTemp uint64
	var ObjectIdTemp uint64
	var ActionFlagsTemp uint8
	err := pool.QueryRowContext(ctx, "select o.i_d, o.name_id, o.object_id, o.action_flags from roles as o where o.i_d = :ID;", sql.Named("ID", ID)).Scan(&IDTemp, &NameIdTemp, &ObjectIdTemp, &ActionFlagsTemp)
	return MakeRole(IDTemp, NameIdTemp, ObjectIdTemp, ActionFlagsTemp), err
}

func (o Role) Update(pool ExecerContext, ctx context.Context) error {
	_, err := updateRole(pool, ctx, o.ID, o.NameId, o.ObjectId, o.ActionFlags)
	return err
}

func (o Role) Delete(pool ExecerContext, ctx context.Context) error {
	_, err := deleteRole(pool, ctx, o.ID)
	return err
}

func createRole(pool ExecerContext, ctx context.Context, ID uint64, NameId uint64, ObjectId uint64, ActionFlags uint8) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := pool.ExecContext(ctx, "insert into roles(i_d, name_id, object_id, action_flags) values(:ID, :NameId, :ObjectId, :ActionFlags);", sql.Named("ID", ID), sql.Named("NameId", NameId), sql.Named("ObjectId", ObjectId), sql.Named("ActionFlags", ActionFlags))
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func updateRole(pool ExecerContext, ctx context.Context, ID uint64, NameId uint64, ObjectId uint64, ActionFlags uint8) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := pool.ExecContext(ctx, "update roles set name_id = :NameId, object_id = :ObjectId, action_flags = :ActionFlags where i_d = :ID;", sql.Named("ID", ID), sql.Named("NameId", NameId), sql.Named("ObjectId", ObjectId), sql.Named("ActionFlags", ActionFlags))
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func deleteRole(pool ExecerContext, ctx context.Context, ID uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := pool.ExecContext(ctx, "delete from roles where i_d = :ID;", sql.Named("ID", ID))
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func GetRolesByUserId(pool QueryerContext, ctx context.Context, userId uint64) ([]Role, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var IDTemp uint64
	var NameIdTemp uint64
	var ObjectIdTemp uint64
	var ActionFlagsTemp uint8
	rows, err := pool.QueryContext(ctx, "select r.i_d, r.name_id, r.object_id, r.action_flags from roles as r where r.id in (select o.role_id from user_roles as o where o.user_id = :userId);", sql.Named("userId", userId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []Role{}
	for rows.Next() {
		err := rows.Scan(&IDTemp, &NameIdTemp, &ObjectIdTemp, &ActionFlagsTemp)
		if err != nil {
			return nil, err
		}
		results = append(results, MakeRole(IDTemp, NameIdTemp, ObjectIdTemp, ActionFlagsTemp))
	}
	return results, nil
}
func GetRolesByObjectIds(pool QueryerContext, ctx context.Context, objectIds []uint64) ([]Role, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var IDTemp uint64
	var NameIdTemp uint64
	var ObjectIdTemp uint64
	var ActionFlagsTemp uint8
	rows, err := pool.QueryContext(ctx, "select r.i_d, r.name_id, r.object_id, r.action_flags from roles as r where r.object_id in (:objectIds);", sql.Named("objectIds", objectIds))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []Role{}
	for rows.Next() {
		err := rows.Scan(&IDTemp, &NameIdTemp, &ObjectIdTemp, &ActionFlagsTemp)
		if err != nil {
			return nil, err
		}
		results = append(results, MakeRole(IDTemp, NameIdTemp, ObjectIdTemp, ActionFlagsTemp))
	}
	return results, nil
}
func GetRoleByNameAndObjectId(pool RowQueryerContext, ctx context.Context, name string, objectId uint64) (Role, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var IDTemp uint64
	var NameIdTemp uint64
	var ObjectIdTemp uint64
	var ActionFlagsTemp uint8
	err := pool.QueryRowContext(ctx, "select r.i_d, r.name_id, r.object_id, r.action_flags from roles as r, role_names as n where r.name_id = n.id and n.name = :name and r.object_id = :objectId;;", sql.Named("name", name), sql.Named("objectId", objectId)).Scan(&IDTemp, &NameIdTemp, &ObjectIdTemp, &ActionFlagsTemp)
	return MakeRole(IDTemp, NameIdTemp, ObjectIdTemp, ActionFlagsTemp), err
}
func GetRoleByNameIdAndObjectId(pool RowQueryerContext, ctx context.Context, nameId uint64, objectId uint64) (Role, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var IDTemp uint64
	var NameIdTemp uint64
	var ObjectIdTemp uint64
	var ActionFlagsTemp uint8
	err := pool.QueryRowContext(ctx, "select r.i_d, r.name_id, r.object_id, r.action_flags from roles as r, role_names as n where r.name_id = :nameId and r.object_id = :objectId;;", sql.Named("nameId", nameId), sql.Named("objectId", objectId)).Scan(&IDTemp, &NameIdTemp, &ObjectIdTemp, &ActionFlagsTemp)
	return MakeRole(IDTemp, NameIdTemp, ObjectIdTemp, ActionFlagsTemp), err
}
func GetRolesByNameAndObjectIds(pool QueryerContext, ctx context.Context, name string, objectIds []uint64) ([]Role, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var IDTemp uint64
	var NameIdTemp uint64
	var ObjectIdTemp uint64
	var ActionFlagsTemp uint8
	rows, err := pool.QueryContext(ctx, "select r.i_d, r.name_id, r.object_id, r.action_flags from roles as r, role_names as n where r.name_id = n.id and n.name = :name and r.object_id in (:objectIds);", sql.Named("name", name), sql.Named("objectIds", objectIds))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []Role{}
	for rows.Next() {
		err := rows.Scan(&IDTemp, &NameIdTemp, &ObjectIdTemp, &ActionFlagsTemp)
		if err != nil {
			return nil, err
		}
		results = append(results, MakeRole(IDTemp, NameIdTemp, ObjectIdTemp, ActionFlagsTemp))
	}
	return results, nil
}
