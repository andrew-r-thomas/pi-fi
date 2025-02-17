package meta

import (
	"context"
	"database/sql"
)

type MetaStore struct {
	db  *sql.DB
	ctx context.Context
}

func NewMetaStore(ctx context.Context) (ms MetaStore, err error) {
	ms.db, err = sql.Open("sqlite3", "./.db")
	if err != nil {
		return
	}

	ms.ctx = ctx
	err = ms.setup()

	return
}

func (ms *MetaStore) setup() (err error) {
	_, err = ms.db.ExecContext(ms.ctx, `
		create table if not exists artists(
			id integer primary key,
			name text not null unique
		);

		create table if not exists albums(
			id integer primary key,
			title text not null,
			artist integer not null,
			foreign key (artist) references artists(id),
			unique(artist,title)
		);

		create table if not exists tracks(
			id integer primary key,
			title text not null,
			track_number integer not null,
			album integer not null,
			foreign key (album) references albums(id),
			unique(track_number,album)
		);
	`)
	if err != nil {
		return
	}

	return
}

func (ms *MetaStore) AddTrack(
	title string,
	artist string,
	album string,
	trackNumber int,
) (albumId int64, err error) {
	// setup transaction
	tx, err := ms.db.BeginTx(ms.ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()

	// upsert the artist
	res, err := tx.ExecContext(ms.ctx, `
		insert into artists (name)
		values (?)
		on conflict (name) do nothing
		returning id
	`, artist)
	if err != nil {
		return
	}
	artistId, err := res.LastInsertId()
	if err != nil {
		return
	}

	// upsert the album
	res, err = tx.ExecContext(ms.ctx, `
		insert into albums (title, artist)
		values (?, ?)
		on conflict (title,artist) do nothing
		returning id
	`, album, artistId)
	if err != nil {
		return
	}
	albumId, err = res.LastInsertId()
	if err != nil {
		return
	}

	// insert track, take error if unique constraint is violated
	_, err = tx.ExecContext(ms.ctx, `
		insert into tracks (title, album, track_number)
		values (?, ?, ?)
	`, title, albumId, trackNumber)
	if err != nil {
		return
	}

	// commit transaction and return
	err = tx.Commit()
	return
}

func (ms *MetaStore) GetAlbum(id int) (Album, error) {
	var album Album

	row := ms.db.QueryRow("select * from albums where id = ?", id)
	err := row.Scan(&album.Id, &album.Title, &album.Artist)
	if err != nil {
		return album, err
	}

	return album, nil
}
