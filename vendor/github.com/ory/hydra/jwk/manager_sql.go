package jwk

import (
	"database/sql"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/ory/hydra/pkg"
	"github.com/pkg/errors"
	"github.com/rubenv/sql-migrate"
	"github.com/square/go-jose"
)

type SQLManager struct {
	DB     *sqlx.DB
	Cipher *AEAD
}

var migrations = &migrate.MemoryMigrationSource{
	Migrations: []*migrate.Migration{
		{
			Id: "1",
			Up: []string{
				`CREATE TABLE IF NOT EXISTS hydra_jwk (
	sid     varchar(255) NOT NULL,
	kid 	varchar(255) NOT NULL,
	version int NOT NULL DEFAULT 0,
	keydata text NOT NULL,
	PRIMARY KEY (sid, kid)
)`,
			},
			Down: []string{
				"DROP TABLE hydra_jwk",
			},
		},
	},
}

type sqlData struct {
	Set     string `db:"sid"`
	KID     string `db:"kid"`
	Version int    `db:"version"`
	Key     string `db:"keydata"`
}

func (s *SQLManager) CreateSchemas() (int, error) {
	migrate.SetTable("hydra_jwk_migration")
	n, err := migrate.Exec(s.DB.DB, s.DB.DriverName(), migrations, migrate.Up)
	if err != nil {
		return 0, errors.Wrapf(err, "Could not migrate sql schema, applied %d migrations", n)
	}
	return n, nil
}

func (m *SQLManager) AddKey(set string, key *jose.JsonWebKey) error {
	out, err := json.Marshal(key)
	if err != nil {
		return errors.WithStack(err)
	}

	encrypted, err := m.Cipher.Encrypt(out)
	if err != nil {
		return errors.WithStack(err)
	}

	if _, err = m.DB.NamedExec(`INSERT INTO hydra_jwk (sid, kid, version, keydata) VALUES (:sid, :kid, :version, :keydata)`, &sqlData{
		Set:     set,
		KID:     key.KeyID,
		Version: 0,
		Key:     encrypted,
	}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (m *SQLManager) AddKeySet(set string, keys *jose.JsonWebKeySet) error {
	tx, err := m.DB.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}

	for _, key := range keys.Keys {
		out, err := json.Marshal(key)
		if err != nil {
			if re := tx.Rollback(); re != nil {
				return errors.Wrap(err, re.Error())
			}
			return errors.WithStack(err)
		}

		encrypted, err := m.Cipher.Encrypt(out)
		if err != nil {
			if re := tx.Rollback(); re != nil {
				return errors.Wrap(err, re.Error())
			}
			return errors.WithStack(err)
		}

		if _, err = tx.NamedExec(`INSERT INTO hydra_jwk (sid, kid, version, keydata) VALUES (:sid, :kid, :version, :keydata)`, &sqlData{
			Set:     set,
			KID:     key.KeyID,
			Version: 0,
			Key:     encrypted,
		}); err != nil {
			if re := tx.Rollback(); re != nil {
				return errors.Wrap(err, re.Error())
			}
			return errors.WithStack(err)
		}
	}

	if err := tx.Commit(); err != nil {
		if re := tx.Rollback(); re != nil {
			return errors.Wrap(err, re.Error())
		}
		return errors.WithStack(err)
	}
	return nil
}

func (m *SQLManager) GetKey(set, kid string) (*jose.JsonWebKeySet, error) {
	var d sqlData
	if err := m.DB.Get(&d, m.DB.Rebind("SELECT * FROM hydra_jwk WHERE sid=? AND kid=?"), set, kid); err == sql.ErrNoRows {
		return nil, errors.Wrap(pkg.ErrNotFound, "")
	} else if err != nil {
		return nil, errors.WithStack(err)
	}

	key, err := m.Cipher.Decrypt(d.Key)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var c jose.JsonWebKey
	if err := json.Unmarshal(key, &c); err != nil {
		return nil, errors.WithStack(err)
	}

	return &jose.JsonWebKeySet{
		Keys: []jose.JsonWebKey{c},
	}, nil
}

func (m *SQLManager) GetKeySet(set string) (*jose.JsonWebKeySet, error) {
	var ds []sqlData
	if err := m.DB.Select(&ds, m.DB.Rebind("SELECT * FROM hydra_jwk WHERE sid=?"), set); err == sql.ErrNoRows {
		return nil, errors.Wrap(pkg.ErrNotFound, "")
	} else if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(ds) == 0 {
		return nil, errors.Wrap(pkg.ErrNotFound, "")
	}

	keys := &jose.JsonWebKeySet{Keys: []jose.JsonWebKey{}}
	for _, d := range ds {
		key, err := m.Cipher.Decrypt(d.Key)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		var c jose.JsonWebKey
		if err := json.Unmarshal(key, &c); err != nil {
			return nil, errors.WithStack(err)
		}
		keys.Keys = append(keys.Keys, c)
	}

	return keys, nil
}

func (m *SQLManager) DeleteKey(set, kid string) error {
	if _, err := m.DB.Exec(m.DB.Rebind(`DELETE FROM hydra_jwk WHERE sid=? AND kid=?`), set, kid); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (m *SQLManager) DeleteKeySet(set string) error {
	if _, err := m.DB.Exec(m.DB.Rebind(`DELETE FROM hydra_jwk WHERE sid=?`), set); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
