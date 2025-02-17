package meta

type Album struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Artist int    `json:"artist"`
}

type AlbumMeta struct {
	Title  string      `json:"title"`
	Artist string      `json:"artist"`
	Tracks []TrackMeta `json:"tracks"`
}

type AddAlbumResp struct {
	AlbumId  int64   `json:"album_id"`
	ArtistId int64   `json:"artist_id"`
	TrackIds []int64 `json:"track_ids"`
}
