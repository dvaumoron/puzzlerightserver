// Generated from model.crn - do not edit.

package model

import "context"

type UserRole struct {
	Id     uint64
	UserId uint64
	RoleId uint64
}

func MakeUserRole(id uint64, userid uint64, roleid uint64) UserRole {
	return UserRole{
		Id:     id,
		RoleId: roleid,
		UserId: userid,
	}
}

func (o UserRole) Create(pool ExecerContext, ctx context.Context) error {
	_, err := createUserRole(pool, ctx, o.Id, o.UserId, o.RoleId)
	return err
}

func ReadUserRole(pool RowQueryerContext, ctx context.Context, Id uint64) (UserRole, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "select o.id, o.user_id, o.role_id from user_roles as o where o.id = $1;"
	var IdTemp uint64
	var UserIdTemp uint64
	var RoleIdTemp uint64
	err := pool.QueryRowContext(ctx, query, Id).Scan(&IdTemp, &UserIdTemp, &RoleIdTemp)
	return MakeUserRole(IdTemp, UserIdTemp, RoleIdTemp), err
}

func (o UserRole) Update(pool ExecerContext, ctx context.Context) error {
	_, err := updateUserRole(pool, ctx, o.Id, o.UserId, o.RoleId)
	return err
}

func (o UserRole) Delete(pool ExecerContext, ctx context.Context) error {
	_, err := deleteUserRole(pool, ctx, o.Id)
	return err
}

func createUserRole(pool ExecerContext, ctx context.Context, Id uint64, UserId uint64, RoleId uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "insert into user_roles(user_id, role_id) values($2, $3);"
	result, err := pool.ExecContext(ctx, query, Id, UserId, RoleId)
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func updateUserRole(pool ExecerContext, ctx context.Context, Id uint64, UserId uint64, RoleId uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "update user_roles set user_id = $2, role_id = $3 where id = $1;"
	result, err := pool.ExecContext(ctx, query, Id, UserId, RoleId)
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func deleteUserRole(pool ExecerContext, ctx context.Context, Id uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "delete from user_roles where id = $1;"
	result, err := pool.ExecContext(ctx, query, Id)
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}

func DeleteUserRolesByUserId(pool ExecerContext, ctx context.Context, userId uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "delete from user_roles where user_id = $1;"
	result, err := pool.ExecContext(ctx, query, userId)
	if err != nil {
		return int64(0), err
	}
	return result.RowsAffected()
}
