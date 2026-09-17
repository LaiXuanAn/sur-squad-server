package matchmaking

import "testing"

func TestMatchQueryUsesNakamaBooleanToken(t *testing.T) {
	query := matchQuery(Request{Mode: "survival"})
	want := "+label.mode:survival +label.joinable:T +label.max_players:32"
	if query != want {
		t.Fatalf("unexpected query: got %q want %q", query, want)
	}
}

func TestParseRequestDefaults(t *testing.T) {
	request, err := parseRequest("")
	if err != nil {
		t.Fatal(err)
	}
	if request.Mode != "survival" {
		t.Fatalf("unexpected defaults: %+v", request)
	}
}
