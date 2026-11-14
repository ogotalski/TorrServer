package torr

import (
	"encoding/json"

	"server/settings"
	"server/torr/state"
	"server/torr/utils"

	"github.com/anacrolix/torrent/metainfo"
)

type tsFiles struct {
	TorrServer struct {
		Files []*state.TorrentFileStat `json:"Files"`
	} `json:"TorrServer"`
}

// withFileList is a torrent's data with its file list for the web UI ({"TorrServer":{"Files":…}}):
// the list itself when the data is empty, added to a client's JSON object that has none (the
// NothingOn apps keep their link to a title there), any other data left as it is.
func withFileList(data string, files []*state.TorrentFileStat) string {
	if data == "" {
		list := new(tsFiles)
		list.TorrServer.Files = files
		if buf, err := json.Marshal(list); err == nil {
			return string(buf)
		}
		return data
	}
	if len(files) == 0 {
		return data
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &fields); err != nil || fields == nil {
		return data
	}
	if _, ok := fields["TorrServer"]; ok {
		return data
	}
	list, err := json.Marshal(struct {
		Files []*state.TorrentFileStat `json:"Files"`
	}{files})
	if err != nil {
		return data
	}
	fields["TorrServer"] = list
	buf, err := json.Marshal(fields)
	if err != nil {
		return data
	}
	return string(buf)
}

func AddTorrentDB(torr *Torrent) {
	t := new(settings.TorrentDB)
	t.TorrentSpec = torr.TorrentSpec
	t.Title = torr.Title
	t.Category = torr.Category
	t.Data = withFileList(torr.Data, torr.Status().FileStats)
	torr.Data = t.Data

	t.Poster = torr.Poster
	if t.Poster != "" {
		var existing string
		if db := GetTorrentDB(torr.Hash()); db != nil {
			existing = db.Poster
		}
		if existing != t.Poster {
			ok, verified := utils.CheckImgUrl(t.Poster)
			if !ok || (existing != "" && !verified) {
				t.Poster = existing
			}
		}
	}
	t.Size = torr.Size
	if t.Size == 0 && torr.Torrent != nil {
		t.Size = torr.Torrent.Length()
	}
	// don't override timestamp from DB on edit
	t.Timestamp = torr.Timestamp // time.Now().Unix()
	t.OfflineState = torr.OfflineState()
	t.OfflineDir = torr.offlineDirRaw()
	t.OfflineUsedAt = torr.offlineUsedAt()
	if t.OfflineState != "" {
		// A live torrent object does not carry these: keep the stored values. The pin is
		// changed only by SetOfflinePinned, never by saving the torrent.
		if old := GetTorrentDB(torr.Hash()); old != nil {
			t.OfflinePinned = old.OfflinePinned()
			if t.OfflineUsedAt == 0 {
				t.OfflineUsedAt = old.offlineUsedAt()
			}
		}
	}

	settings.AddTorrent(t)
}

func GetTorrentDB(hash metainfo.Hash) *Torrent {
	list := settings.ListTorrent()
	for _, db := range list {
		if hash == db.InfoHash {
			torr := new(Torrent)
			torr.TorrentSpec = db.TorrentSpec
			torr.Title = db.Title
			torr.Poster = db.Poster
			torr.Category = db.Category
			torr.Timestamp = db.Timestamp
			torr.Size = db.Size
			torr.Data = db.Data
			torr.setOfflineState(db.OfflineState)
			torr.setOfflineDir(db.OfflineDir)
			torr.setOfflineUsedAt(db.OfflineUsedAt)
			torr.setOfflinePinned(db.OfflinePinned)
			torr.Stat = state.TorrentInDB
			return torr
		}
	}
	return nil
}

func RemTorrentDB(hash metainfo.Hash) {
	settings.RemTorrent(hash)
}

func ListTorrentsDB() map[metainfo.Hash]*Torrent {
	ret := make(map[metainfo.Hash]*Torrent)
	list := settings.ListTorrent()
	for _, db := range list {
		torr := new(Torrent)
		torr.TorrentSpec = db.TorrentSpec
		torr.Title = db.Title
		torr.Poster = db.Poster
		torr.Category = db.Category
		torr.Timestamp = db.Timestamp
		torr.Size = db.Size
		torr.Data = db.Data
		torr.setOfflineState(db.OfflineState)
		torr.setOfflineDir(db.OfflineDir)
		torr.setOfflineUsedAt(db.OfflineUsedAt)
		torr.setOfflinePinned(db.OfflinePinned)
		torr.Stat = state.TorrentInDB
		ret[torr.TorrentSpec.InfoHash] = torr
	}
	return ret
}
