package stream

import (
	"fmt"
	"testing"
)

func TestParseStreamID(t *testing.T) {
	tests := []struct {
		name         string
		streamID     string
		wantMode     Mode
		wantName     string
		wantPass     string
		wantUsername string
		wantErr      error
	}{
		// Old school
		{"MissingSlash", "s1", 0, "", "", "", InvalidSlashes},
		{"InvalidName", "play//s1", 0, "", "", "", MissingName},
		{"InvalidMode", "foobar/bla", 0, "", "", "", InvalidMode},
		{"InvalidSlash", "foobar/bla//", 0, "", "", "", InvalidSlashes},
		{"EmptyPass", "play/s1/", ModePlay, "s1", "", "", nil},
		{"ValidPass", "play/s1/#![äöü", ModePlay, "s1", "#![äöü", "", nil},
		{"ValidPlay", "play/s1", ModePlay, "s1", "", "", nil},
		{"ValidPublish", "publish/abcdef", ModePublish, "abcdef", "", "", nil},
		{"ValidPlaySpace", "play/bla fasel", ModePlay, "bla fasel", "", "", nil},
		// New hotness - Bad
		{"NewInvalidPubEmptyName", "#!::m=publish", ModePublish, "", "", "", MissingName},
		{"NewInvalidPlayEmptyName", "#!::m=request", ModePlay, "", "", "", MissingName},
		{"NewInvalidPubBadKey", "#!::m=publish,y=bar", ModePublish, "", "", "", fmt.Errorf("unsupported key '%s'", "y")},
		{"NewInvalidPlayBadKey", "#!::m=request,x=foo", ModePlay, "", "", "", fmt.Errorf("unsupported key '%s'", "x")},
		{"NewInvalidPubNoEquals", "#!::m=publish,r", ModePublish, "abc", "", "", InvalidValue},
		{"NewInvalidPlayNoEquals", "#!::m=request,r", ModePlay, "abc", "", "", InvalidValue},
		{"NewInvalidPubNoValue", "#!::m=publish,r=", ModePublish, "abc", "", "", MissingName},
		{"NewInvalidPlayNoValue", "#!::m=request,s=", ModePlay, "abc", "", "", MissingName},
		{"NewInvalidPubBadKey", "#!::m=publish,x=", ModePublish, "abc", "", "", fmt.Errorf("unsupported key '%s'", "x")},
		{"NewInvalidPlayBadKey", "#!::m=request,y=", ModePlay, "abc", "", "", fmt.Errorf("unsupported key '%s'", "y")},
		// New hotness - Standard
		{"NewValidNameRequest", "#!::m=publish,r=abc", ModePublish, "abc", "", "", nil},
		{"NewValidPlay", "#!::m=request,r=abc", ModePlay, "abc", "", "", nil},
		{"NewValidNameRequestRev", "#!::r=abc,m=publish", ModePublish, "abc", "", "", nil},
		{"NewValidPlayRev", "#!::r=abc,m=request", ModePlay, "abc", "", "", nil},
		{"NewValidPassPub", "#!::m=publish,r=abc,s=bob", ModePublish, "abc", "bob", "", nil},
		{"NewValidPassPlay", "#!::m=request,r=abc,s=alice", ModePlay, "abc", "alice", "", nil},
		{"NewValidPassPubOrder", "#!::s=bob,m=publish,r=abc123", ModePublish, "abc123", "bob", "", nil},
		{"NewValidPassPlayOrder", "#!::m=request,s=alice,r=def", ModePlay, "def", "alice", "", nil},
		{"NewValidPubUsername", "#!::s=bob,m=publish,r=abc123,u=eve", ModePublish, "abc123", "bob", "eve", nil},
		{"NewValidPlayUsername", "#!::m=request,s=alice,r=def,u=bar", ModePlay, "def", "alice", "bar", nil},
		{"NewValidPubUsernameOrder", "#!::s=bob,m=publish,u=eve,r=abc123", ModePublish, "abc123", "bob", "eve", nil},
		{"NewValidPlayUsernameOrder", "#!::m=request,u=bar,s=alice,r=def", ModePlay, "def", "alice", "bar", nil},
		// New Hotness - Unicode
		{"NewValidUnicodePub", "#!::m=publish,r=#![äöü,s=bob", ModePublish, "#![äöü", "bob", "", nil},
		{"NewValidUnicodePlay", "#!::m=request,r=#![äöü,s=alice", ModePlay, "#![äöü", "alice", "", nil},
		{"NewValidUnicodePassPub", "#!::m=publish,s=#![äöü,r=bob", ModePublish, "bob", "#![äöü", "", nil},
		{"NewValidUnicodePassPlay", "#!::m=request,s=#![äöü,r=alice", ModePlay, "alice", "#![äöü", "", nil},
		{"NewValidUnicodeUserPub", "#!::s=bye,m=publish,u=#![äöü,r=art", ModePublish, "art", "bye", "#![äöü", nil},
		{"NewValidUnicodeUserPlay", "#!::m=request,u=#![äöü,r=eve,s=hai", ModePlay, "eve", "hai", "#![äöü", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var streamid StreamID
			err := streamid.FromString(tt.streamID)

			if err != nil {
				if err.Error() != tt.wantErr.Error() { // Only really care about str value for this, otherwise: if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseStreamID() error = %v, wantErr %v", err, tt.wantErr)
				}
				if streamid.String() != "" {
					t.Error("str should be empty on failed parse")
				}
				return
			}
			if name := streamid.Name(); name != tt.wantName {
				t.Errorf("ParseStreamID() got Name = %v, want %v", name, tt.wantName)
			}
			if mode := streamid.Mode(); mode != tt.wantMode {
				t.Errorf("ParseStreamID() got Mode = %v, want %v", mode, tt.wantMode)
			}
			if password := streamid.Password(); password != tt.wantPass {
				t.Errorf("ParseStreamID() got Password = %v, want %v", password, tt.wantMode)
			}
			if str := streamid.String(); str != tt.streamID {
				t.Errorf("String() got String = %v, want %v", str, tt.streamID)
			}
			if str := streamid.Username(); str != tt.wantUsername {
				t.Errorf("Username() got String = %v, want %v", str, tt.wantUsername)
			}
		})
	}
}

func TestNewStreamID(t *testing.T) {
	tests := []struct {
		name         string
		argName      string
		argMode      Mode
		argPassword  string
		wantStreamID string
		wantErr      error
	}{
		{"InvalidMode", "s1", 0, "", "", InvalidMode},
		{"InvalidName", "s1/", ModePlay, "", "", InvalidNamePassword},
		{"InvalidPass", "s1", ModePlay, "foo/bar", "", InvalidNamePassword},
		{"ValidPlay", "s1", ModePlay, "", "play/s1", nil},
		{"ValidPublish", "s1", ModePublish, "", "publish/s1", nil},
		{"ValidPlayPass", "s1", ModePlay, "foo", "play/s1/foo", nil},
		{"ValidPublishPass", "s1", ModePublish, "foo", "publish/s1/foo", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := NewStreamID(tt.argName, tt.argPassword, tt.argMode)
			if err != tt.wantErr {
				t.Errorf("ParseStreamID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if id != nil {
					t.Error("id should be nil on failed parse")
				}
				return
			}
			if err != nil {
				t.Error(err)
			}
			if str := id.String(); str != tt.wantStreamID {
				t.Errorf("NewStreamID() got String = %v, want %v", str, tt.wantStreamID)
			}
		})
	}
}

func TestStreamID_Match(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		pattern string
		want    bool
	}{
		{"MatchAll", "publish/foo/bar", "*", true},
		{"FlatMatch", "publish/foo/bar", "pub*bar", true},
		{"CompleteMatch", "play/one/two", "play/one/two", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s StreamID
			err := s.FromString(tt.id)
			if err != nil {
				t.Error(err)
			}
			if got := s.Match(tt.pattern); got != tt.want {
				t.Errorf("StreamID.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
