package stream

import (
	"testing"
)

func TestParseStreamID(t *testing.T) {
	tests := []struct {
		name     string
		streamID string
		wantMode Mode
		wantName string
		wantPass string
		wantErr  error
	}{
		{"MissingSlash", "s1", 0, "", "", InvalidStreamID},
		{"InvalidName", "play//s1", 0, "", "", InvalidStreamID},
		{"InvalidMode", "foobar/bla", 0, "", "", InvalidMode},
		{"InvalidSlash", "foobar/bla//", 0, "", "", InvalidStreamID},
		{"EmptyPass", "play/s1/", ModePlay, "s1", "", nil},
		{"ValidPass", "play/s1/#![äöü", ModePlay, "s1", "#![äöü", nil},
		{"ValidPlay", "play/s1", ModePlay, "s1", "", nil},
		{"ValidPublish", "publish/abcdef", ModePublish, "abcdef", "", nil},
		{"ValidPlaySpace", "play/bla fasel", ModePlay, "bla fasel", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var streamid StreamID
			err := streamid.FromString(tt.streamID)
			if err != tt.wantErr {
				t.Errorf("ParseStreamID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
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
