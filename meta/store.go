/*

TODO:
- think about using prepared statments

*/

package meta

import (
	"context"
	"database/sql"
	"os"
)

type MetaStore struct {
	db  *sql.DB
	ctx context.Context
}

func NewMetaStore(ctx context.Context) (ms MetaStore, err error) {
	os.Remove("./.db") // for now, just to keep dev easier
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

func (ms *MetaStore) AddAlbum(albumMeta *AlbumMeta) (
	resp AddAlbumResp,
	err error,
) {
	// setup transaction
	tx, err := ms.db.BeginTx(ms.ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return
	}

	// first we upsert the artist
	row := tx.QueryRowContext(ms.ctx, `
		insert into artists(name)
		values (?)
		on conflict (name) do update set name = name
		returning id
	`, albumMeta.Artist)
	err = row.Scan(&resp.ArtistId)
	if err != nil {
		return
	}

	// then we insert the album, if it exists, we return an error
	row = tx.QueryRowContext(ms.ctx, `
		insert into albums(title, artist)
		values (?, ?)
		returning id
	`, albumMeta.Title, resp.ArtistId)
	err = row.Scan(&resp.AlbumId)
	if err != nil {
		return
	}

	// then for each track we insert, if exists -> return error
	stmt, err := tx.PrepareContext(ms.ctx, `
		insert into tracks(title, track_number, album)
		values (?, ?, ?)
		returning id
	`)
	if err != nil {
		return
	}
	for _, track := range albumMeta.Tracks {
		row = stmt.QueryRowContext(
			ms.ctx,
			track.Title,
			track.TrackNumber,
			resp.AlbumId,
		)
		var trackId int64
		err = row.Scan(&trackId)
		if err != nil {
			return
		}
		resp.TrackIds = append(resp.TrackIds, trackId)
	}

	// commit transaction and return
	err = tx.Commit()
	return
}
