package torr

import (
	"encoding/json"
	"testing"

	"server/torr/state"
)

var someFiles = []*state.TorrentFileStat{{Id: 1, Path: "Show/Show.S01E01.mkv", Length: 100}}

func filesIn(t *testing.T, data string) []*state.TorrentFileStat {
	t.Helper()
	var f tsFiles
	if err := json.Unmarshal([]byte(data), &f); err != nil {
		t.Fatalf("data is not JSON: %q", data)
	}
	return f.TorrServer.Files
}

func TestWithFileListFillsEmptyData(t *testing.T) {
	data := withFileList("", someFiles)
	if got := filesIn(t, data); len(got) != 1 || got[0].Path != "Show/Show.S01E01.mkv" {
		t.Fatalf("want the file list, got %q", data)
	}
}

func TestWithFileListAddsTheFilesToAClientsData(t *testing.T) {
	// A client's own key (the NothingOn apps' link to a title) stays; the web UI gets its files.
	data := withFileList(`{"nothingon":{"key":"Torrent: x"}}`, someFiles)
	if got := filesIn(t, data); len(got) != 1 {
		t.Fatalf("want the file list added, got %q", data)
	}
	var all map[string]json.RawMessage
	_ = json.Unmarshal([]byte(data), &all)
	if string(all["nothingon"]) != `{"key":"Torrent: x"}` {
		t.Fatalf("the client's key must stay as it was, got %q", data)
	}
}

func TestWithFileListLeavesOtherDataAlone(t *testing.T) {
	for _, data := range []string{
		`{"TorrServer":{"Files":[]},"nothingon":{}}`, // has its file list already
		`not json`,
		`["a list"]`,
	} {
		if got := withFileList(data, someFiles); got != data {
			t.Fatalf("%q must stay as it is, got %q", data, got)
		}
	}
	if got := withFileList(`{"nothingon":{}}`, nil); got != `{"nothingon":{}}` {
		t.Fatalf("no files known: unchanged, got %q", got)
	}
}
