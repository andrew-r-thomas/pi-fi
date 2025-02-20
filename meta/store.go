/*

TODO:
- think about using prepared statments

*/

package meta

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
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

type GetLibResp struct {
	Albums  []GetLibRespAlbum  `json:"albums"`
	Artists []GetLibRespArtist `json:"artists"`
	Tracks  []GetLibRespTrack  `json:"tracks"`
}
type GetLibRespAlbum struct {
	Id       int64   `json:"id"`
	Title    string  `json:"title"`
	ArtistId int64   `json:"artist_id"`
	TrackIds []int64 `json:"track_ids"`
}
type GetLibRespArtist struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}
type GetLibRespTrack struct {
	Id          int64  `json:"id"`
	Title       string `json:"title"`
	ArtistId    int64  `json:"artist_id"`
	AlbumId     int64  `json:"album_id"`
	TrackNumber uint32 `json:"track_number"`
}

func (ms *MetaStore) GetLibrary() (*GetLibResp, error) {
	resp := &GetLibResp{
		Albums:  make([]GetLibRespAlbum, 0),
		Artists: make([]GetLibRespArtist, 0),
		Tracks:  make([]GetLibRespTrack, 0),
	}

	// add artist info to return payload
	artistRows, err := ms.db.Query("select id, name from artists")
	if err != nil {
		return nil, err
	}
	defer artistRows.Close()
	for artistRows.Next() {
		var artist GetLibRespArtist
		err = artistRows.Scan(&artist.Id, &artist.Name)
		if err != nil {
			return nil, err
		}
		resp.Artists = append(resp.Artists, artist)
	}
	err = artistRows.Err()
	if err != nil {
		return nil, err
	}

	// add album into to return payload
	albumRows, err := ms.db.Query(`
		select
			a.id,
			a.title,
			a.artist as artist_id,
			group_concat(t.id) as track_ids
		from albums a
		left join tracks t on t.album = a.id
		group by a.id, a.title, a.artist
	`)
	if err != nil {
		return nil, err
	}
	defer albumRows.Close()
	for albumRows.Next() {
		var album GetLibRespAlbum
		var trackStr sql.NullString
		err = albumRows.Scan(
			&album.Id,
			&album.Title,
			&album.ArtistId,
			&trackStr,
		)
		if err != nil {
			return nil, err
		}

		// split up track id string into ids
		if trackStr.Valid && trackStr.String != "" {
			trackStrs := strings.Split(trackStr.String, ",")
			album.TrackIds = make([]int64, 0, len(trackStrs))

			for _, idStr := range trackStrs {
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					return nil, err
				}
				album.TrackIds = append(album.TrackIds, id)
			}
		}

		resp.Albums = append(resp.Albums, album)
	}
	err = albumRows.Err()
	if err != nil {
		return nil, err
	}

	// add track data
	trackRows, err := ms.db.Query(`
		select
			t.id,
			t.title,
			t.track_number,
			t.album as album_id,
			a.artist as artist_id
		from tracks t
		join albums a on t.album = a.id
	`)
	if err != nil {
		return nil, err
	}
	defer trackRows.Close()
	for trackRows.Next() {
		var track GetLibRespTrack
		var trackNum int64
		err = trackRows.Scan(
			&track.Id,
			&track.Title,
			&trackNum,
			&track.AlbumId,
			&track.ArtistId,
		)
		if err != nil {
			return nil, err
		}

		track.TrackNumber = uint32(trackNum)
		resp.Tracks = append(resp.Tracks, track)
	}
	err = trackRows.Err()
	if err != nil {
		return nil, err
	}

	return resp, nil
}
