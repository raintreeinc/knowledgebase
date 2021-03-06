package pgdb

import (
	"errors"

	"github.com/raintreeinc/knowledgebase/kb"
)

type Access struct{ Context }

func (db Access) VerifyUser(user kb.User) error {
	if !db.BoolQuery(`
		SELECT FROM Users
		WHERE ID = $1
		  AND AuthID = $2
		  AND AuthProvider = $3
		  AND Email = $4`,
		user.ID, user.AuthID, user.AuthProvider, user.Email) {
		return errors.New("Invalid user.")
	}
	return nil
}

func (db Access) IsAdmin(user kb.Slug) bool {
	return db.BoolQuery(`SELECT FROM Users WHERE ID = $1 AND Admin`, user)
}

func (db Access) SetAdmin(user kb.Slug, isAdmin bool) error {
	r, err := db.Exec(`UPDATE Users SET Admin = $2 WHERE ID = $1`, user, isAdmin)
	if err != nil {
		return err
	}
	affected, _ := r.RowsAffected()
	if affected == 0 {
		return kb.ErrUserNotExist
	}
	return nil
}

func (db Access) Rights(group, user kb.Slug) kb.Rights {
	var rights string

	// If a person is a direct member of the owner group,
	// then he has MaxAccess possible
	err := db.QueryRow(`
		SELECT Access FROM AccessView
		WHERE GroupID = $1 AND UserID = $2
	`, group, user).Scan(&rights)
	if err == nil {
		return kb.Rights(rights)
	}
	return kb.Blocked
}

func (db Access) AddUser(group, user kb.Slug) error {
	_, err := db.Exec(`
		INSERT INTO
		Membership (GroupID, UserID)
		VALUES ($1, $2)
	`, group, user)

	return err
}

func (db Access) RemoveUser(group, user kb.Slug) error {
	_, err := db.Exec(`
		DELETE FROM Membership
		WHERE GroupID = $1 AND UserID = $2
	`, group, user)
	return err
}

func (db Access) CommunityAdd(group, member kb.Slug, rights kb.Rights) error {
	_, err := db.Exec(`
		INSERT INTO
		Community (GroupID, MemberID, Access)
		VALUES ($1, $2, $3)
	`, group, member, string(rights))
	if dupkey(err) {
		_, err = db.Exec(`
			UPDATE Community
			SET Access = $3
			WHERE GroupID = $1 AND MemberID = $2
		`, group, member, string(rights))
	}
	return err
}

func (db Access) CommunityRemove(group, member kb.Slug) error {
	_, err := db.Exec(`
		DELETE FROM Community
		WHERE GroupID = $1 AND MemberID = $2
	`, group, member)
	return err
}

//TODO: fix this for OwnerID, GroupID
func (db Access) List(group kb.Slug) (members []kb.Member, err error) {
	rows, err := db.Query(`
	SELECT Membership.UserID, Users.Name, False, Users.MaxAccess
		FROM Membership
		JOIN Users ON Membership.UserID = Users.ID
		WHERE Membership.GroupID = $1
	UNION
	SELECT Groups.ID, Groups.Name, True, Community.Access
		FROM Community
		JOIN Groups ON Community.MemberID = Groups.ID
		WHERE Community.GroupID = $1
	`, group)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var member kb.Member
		var access string
		err := rows.Scan(&member.ID, &member.Name, &member.IsGroup, &access)
		member.Access = kb.Rights(access)
		if err != nil {
			return members, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}
