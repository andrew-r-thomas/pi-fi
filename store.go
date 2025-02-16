package pifi

import (
	"database/sql"
)

type MetaStore struct {
	db *sql.DB
}

func NewMetaStore() (ms MetaStore, err error) {
	ms.db, err = sql.Open("sqlite3", "./.db")
	if err != nil {
		return
	}

	err = ms.setup()

	return
}

func (ms *MetaStore) setup() (err error) {
	_, err = ms.db.Exec(`
		create table if not exists artists(
			id integer primary key,
			name text not null unique
		);

		create table if not exists albums(
			id integer primary key,
			title text not null,
			artist integer not null,
			foreign key (artist) references artists(id)
		);

		create table if not exists tracks(
			id integer primary key,
			title text not null,
			track_number integer not null,
			album integer not null,
			foreign_key(album) references albums(id),
		);
	`)
	if err != nil {
		return
	}

	return
}

func (ms *MetaStore) AddTrack(
	title string,
	artists []string,
	album string,
	trackNumber int,
) {

}
