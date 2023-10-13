// Generated from model.crn - do not edit.

package model

import "context"

type UserRole struct {
	Id     uint64
	UserId uint64
	RoleId uint64
}

func MakeUserRole(id uint64, userId uint64, roleId uint64) UserRole {
	return UserRole{
		Id:     id,
		RoleId: roleId,
		UserId: userId,
	}
}

func (o UserRole) Create(pool ExecerContext, ctx context.Context) error {
	_, err := createUserRole(pool, ctx, o.UserId, o.RoleId)
	return err
}

func ReadUserRole(pool RowQueryerContext, ctx context.Context, id uint64) (UserRole, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "select o.id, o.user_id, o.role_id from user_roles as o where o.id = $1;"
	var idTemp uint64
	var userIdTemp uint64
	var roleIdTemp uint64
	err := pool.QueryRowContext(ctx, query, id).Scan(&idTemp, &userIdTemp, &roleIdTemp)
	return MakeUserRole(idTemp, userIdTemp, roleIdTemp), err
}

func (o UserRole) Update(pool ExecerContext, ctx context.Context) error {
	_, err := updateUserRole(pool, ctx, o.Id, o.UserId, o.RoleId)
	return err
}

func (o UserRole) Delete(pool ExecerContext, ctx context.Context) error {
	_, err := deleteUserRole(pool, ctx, o.Id)
	return err
}

func createUserRole(pool ExecerContext, ctx context.Context, userId uint64, roleId uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "insert into user_roles(user_id, role_id) values($1, $2);"
	result, err := pool.ExecContext(ctx, query, userId, roleId)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func updateUserRole(pool ExecerContext, ctx context.Context, id uint64, userId uint64, roleId uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "update user_roles set user_id = $2, role_id = $3 where id = $1;"
	result, err := pool.ExecContext(ctx, query, id, userId, roleId)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func deleteUserRole(pool ExecerContext, ctx context.Context, id uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "delete from user_roles where id = $1;"
	result, err := pool.ExecContext(ctx, query, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func DeleteUserRolesByUserId(pool ExecerContext, ctx context.Context, userId uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := "delete from user_roles where user_id = $1;"
	result, err := pool.ExecContext(ctx, query, userId)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
