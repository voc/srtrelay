package server

import (
	"C"
	"testing"
)

func TestParseStreamID(t *testing.T) {
	tests := []struct {
		name     string
		streamID string
		wantName string
		wantMode Mode
		wantErr  error
	}{
		{"MissingSlash", "s1", "", 0, InvalidStreamID},
		{"InvalidSlash", "s1//play", "", 0, InvalidStreamID},
		{"InvalidSlash2", "s1/play/", "", 0, InvalidStreamID},
		{"InvalidMode", "foobar/bla", "", 0, InvalidMode},
		{"ValidPlay", "s1/play", "s1", ModePlay, nil},
		{"ValidPublish", "s1/publish", "s1", ModePublish, nil},
		{"ValidPlaySpace", "bla fasel/play", "bla fasel", ModePlay, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotMode, err := ParseStreamID(tt.streamID)
			if err != tt.wantErr {
				t.Errorf("ParseStreamID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("ParseStreamID() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotMode != tt.wantMode {
				t.Errorf("ParseStreamID() gotMode = %v, want %v", gotMode, tt.wantMode)
			}
		})
	}
}

// struct TestRelay{}
// func (tr* TestRelay) Publish(string) (chan<- []byte, error) {

// }

// func (tr* TestRelay) Subscribe(string) (<-chan []byte, UnsubscribeFunc, error)

// }
func TestServerImpl_play(t *testing.T) {
	s := &ServerImpl{ps: TestRelay}
}
