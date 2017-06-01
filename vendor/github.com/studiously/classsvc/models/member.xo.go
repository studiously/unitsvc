// Package models contains the types for schema 'public'.
package models

// GENERATED BY XO. DO NOT EDIT.

import (
	"errors"

	"github.com/google/uuid"
)

// Member represents a row from 'public.members'.
type Member struct {
	ID      uuid.UUID `json:"id"`       // id
	UserID  uuid.UUID `json:"user_id"`  // user_id
	ClassID uuid.UUID `json:"class_id"` // class_id
	Role    UserRole  `json:"role"`     // role
	Owner   bool      `json:"owner"`    // owner

	// xo fields
	_exists, _deleted bool
}

// Exists determines if the Member exists in the database.
func (m *Member) Exists() bool {
	return m._exists
}

// Deleted provides information if the Member has been deleted from the database.
func (m *Member) Deleted() bool {
	return m._deleted
}

// Insert inserts the Member to the database.
func (m *Member) Insert(db XODB) error {
	var err error

	// if already exist, bail
	if m._exists {
		return errors.New("insert failed: already exists")
	}

	// sql insert query, primary key must be provided
	const sqlstr = `INSERT INTO public.members (` +
		`id, user_id, class_id, role, owner` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5` +
		`)`

	// run query
	XOLog(sqlstr, m.ID, m.UserID, m.ClassID, m.Role, m.Owner)
	err = db.QueryRow(sqlstr, m.ID, m.UserID, m.ClassID, m.Role, m.Owner).Scan(&m.ID)
	if err != nil {
		return err
	}

	// set existence
	m._exists = true

	return nil
}

// Update updates the Member in the database.
func (m *Member) Update(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !m._exists {
		return errors.New("update failed: does not exist")
	}

	// if deleted, bail
	if m._deleted {
		return errors.New("update failed: marked for deletion")
	}

	// sql query
	const sqlstr = `UPDATE public.members SET (` +
		`user_id, class_id, role, owner` +
		`) = ( ` +
		`$1, $2, $3, $4` +
		`) WHERE id = $5`

	// run query
	XOLog(sqlstr, m.UserID, m.ClassID, m.Role, m.Owner, m.ID)
	_, err = db.Exec(sqlstr, m.UserID, m.ClassID, m.Role, m.Owner, m.ID)
	return err
}

// Save saves the Member to the database.
func (m *Member) Save(db XODB) error {
	if m.Exists() {
		return m.Update(db)
	}

	return m.Insert(db)
}

// Upsert performs an upsert for Member.
//
// NOTE: PostgreSQL 9.5+ only
func (m *Member) Upsert(db XODB) error {
	var err error

	// if already exist, bail
	if m._exists {
		return errors.New("insert failed: already exists")
	}

	// sql query
	const sqlstr = `INSERT INTO public.members (` +
		`id, user_id, class_id, role, owner` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5` +
		`) ON CONFLICT (id) DO UPDATE SET (` +
		`id, user_id, class_id, role, owner` +
		`) = (` +
		`EXCLUDED.id, EXCLUDED.user_id, EXCLUDED.class_id, EXCLUDED.role, EXCLUDED.owner` +
		`)`

	// run query
	XOLog(sqlstr, m.ID, m.UserID, m.ClassID, m.Role, m.Owner)
	_, err = db.Exec(sqlstr, m.ID, m.UserID, m.ClassID, m.Role, m.Owner)
	if err != nil {
		return err
	}

	// set existence
	m._exists = true

	return nil
}

// Delete deletes the Member from the database.
func (m *Member) Delete(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !m._exists {
		return nil
	}

	// if deleted, bail
	if m._deleted {
		return nil
	}

	// sql query
	const sqlstr = `DELETE FROM public.members WHERE id = $1`

	// run query
	XOLog(sqlstr, m.ID)
	_, err = db.Exec(sqlstr, m.ID)
	if err != nil {
		return err
	}

	// set deleted
	m._deleted = true

	return nil
}

// Class returns the Class associated with the Member's ClassID (class_id).
//
// Generated from foreign key 'members_class_id_fkey'.
func (m *Member) Class(db XODB) (*Class, error) {
	return ClassByID(db, m.ClassID)
}

// MembersByClassID retrieves a row from 'public.members' as a Member.
//
// Generated from index 'members_class_id_idx'.
func MembersByClassID(db XODB, classID uuid.UUID) ([]*Member, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`id, user_id, class_id, role, owner ` +
		`FROM public.members ` +
		`WHERE class_id = $1`

	// run query
	XOLog(sqlstr, classID)
	q, err := db.Query(sqlstr, classID)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*Member{}
	for q.Next() {
		m := Member{
			_exists: true,
		}

		// scan
		err = q.Scan(&m.ID, &m.UserID, &m.ClassID, &m.Role, &m.Owner)
		if err != nil {
			return nil, err
		}

		res = append(res, &m)
	}

	return res, nil
}

// MemberByID retrieves a row from 'public.members' as a Member.
//
// Generated from index 'members_pkey'.
func MemberByID(db XODB, id uuid.UUID) (*Member, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`id, user_id, class_id, role, owner ` +
		`FROM public.members ` +
		`WHERE id = $1`

	// run query
	XOLog(sqlstr, id)
	m := Member{
		_exists: true,
	}

	err = db.QueryRow(sqlstr, id).Scan(&m.ID, &m.UserID, &m.ClassID, &m.Role, &m.Owner)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

// MemberByUserIDClassID retrieves a row from 'public.members' as a Member.
//
// Generated from index 'members_user_id_class_id_idx'.
func MemberByUserIDClassID(db XODB, userID uuid.UUID, classID uuid.UUID) (*Member, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`id, user_id, class_id, role, owner ` +
		`FROM public.members ` +
		`WHERE user_id = $1 AND class_id = $2`

	// run query
	XOLog(sqlstr, userID, classID)
	m := Member{
		_exists: true,
	}

	err = db.QueryRow(sqlstr, userID, classID).Scan(&m.ID, &m.UserID, &m.ClassID, &m.Role, &m.Owner)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

// MembersByUserID retrieves a row from 'public.members' as a Member.
//
// Generated from index 'members_user_id_idx'.
func MembersByUserID(db XODB, userID uuid.UUID) ([]*Member, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`id, user_id, class_id, role, owner ` +
		`FROM public.members ` +
		`WHERE user_id = $1`

	// run query
	XOLog(sqlstr, userID)
	q, err := db.Query(sqlstr, userID)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*Member{}
	for q.Next() {
		m := Member{
			_exists: true,
		}

		// scan
		err = q.Scan(&m.ID, &m.UserID, &m.ClassID, &m.Role, &m.Owner)
		if err != nil {
			return nil, err
		}

		res = append(res, &m)
	}

	return res, nil
}
