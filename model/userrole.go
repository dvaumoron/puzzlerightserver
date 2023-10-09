// Generated from model.crn - do not edit.

package model

import (
	"context"
	"database/sql"
)

type UserRole struct {
	ID     uint64
	UserId uint64
	RoleId uint64
}

func MakeUserRole(id uint64, userid uint64, roleid uint64) UserRole {
	return UserRole{
		ID:     id,
		RoleId: roleid,
		UserId: userid,
	}
}

func (o UserRole) Create(pool ExecerContext, ctx context.Context) error {
	_, err := createUserRole(pool, ctx, o.ID, o.UserId, o.RoleId)
	return err
}

func ReadUserRole(pool RowQueryerContext, ctx context.Context, ID uint64) (UserRole, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var IDTemp uint64
	var UserIdTemp uint64
	var RoleIdTemp uint64
	err := pool.QueryRowContext(ctx, "select o.i_d, o.user_id, o.role_id from user_roles as o where o.i_d = :ID;", sql.Named("ID", ID)).Scan(&IDTemp, &UserIdTemp, &RoleIdTemp)
	return MakeUserRole(IDTemp, UserIdTemp, RoleIdTemp), err
}

func (o UserRole) Update(pool ExecerContext, ctx context.Context) error {
	_, err := updateUserRole(pool, ctx, o.ID, o.UserId, o.RoleId)
	return err
}

func (o UserRole) Delete(pool ExecerContext, ctx context.Context) error {
	_, err := deleteUserRole(pool, ctx, o.ID)
	return err
}

func createUserRole(pool ExecerContext, ctx context.Context, ID uint64, UserId uint64, RoleId uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := pool.ExecContext(ctx, "insert into user_roles(i_d, user_id, role_id) values(:ID, :UserId, :RoleId);", sql.Named("ID", ID), sql.Named("UserId", UserId), sql.Named("RoleId", RoleId))
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func updateUserRole(pool ExecerContext, ctx context.Context, ID uint64, UserId uint64, RoleId uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := pool.ExecContext(ctx, "update user_roles set user_id = :UserId, role_id = :RoleId where i_d = :ID;", sql.Named("ID", ID), sql.Named("UserId", UserId), sql.Named("RoleId", RoleId))
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func deleteUserRole(pool ExecerContext, ctx context.Context, ID uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := pool.ExecContext(ctx, "delete from user_roles where i_d = :ID;", sql.Named("ID", ID))
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func DeleteUserRolesByUserId(pool ExecerContext, ctx context.Context, userId uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := pool.ExecContext(ctx, "delete from user_roles where user_id = :userId;", sql.Named("userId", userId))
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}
