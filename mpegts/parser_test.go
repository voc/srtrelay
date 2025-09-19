package mpegts

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"testing"
)

type ffprobeFrame struct {
	PictureNumber int    `json:"coded_picture_number"`
	KeyFrame      int    `json:"key_frame"`
	PacketSize    string `json:"pkt_size"`
	PictType      string `json:"pict_type"`
	PTS           int    `json:"pkt_pts"`
	DTS           int    `json:"pkjt_dts"`
}

type ffprobeOutput struct {
	Frames []ffprobeFrame `json:"frames"`
}

func ffprobe(filename string) (ffprobeOutput, error) {
	cmd := exec.Command("ffprobe", "-v", "debug", "-hide_banner", "-of", "json", "-show_frames", filename)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	go func() {
		buf := make([]byte, 1500)
		for {
			res, err := stderr.Read(buf)
			if err != nil {
				break
			}
			log.Printf("%s", buf[:res])
		}
	}()
	var output ffprobeOutput
	if err := json.NewDecoder(stdout).Decode(&output); err != nil {
		log.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	return output, err
}

func checkParser(t *testing.T, p *Parser, data []byte, name string, numFrames int) {
	numPackets := len(data) / PacketLen
	offset := 0
	for i := 0; i < numPackets; i++ {
		packet := data[i*PacketLen : (i+1)*PacketLen]
		offset = (i + 1) * PacketLen
		err := p.Parse(packet)
		if err != nil {
			t.Fatalf("%s - Parse failed: %v", name, err)
		}
		if p.hasInit() {
			break
		}
	}
	if !p.hasInit() {
		t.Errorf("%s - Should find init", name)
	}
	pkts, err := p.InitData()
	if err != nil {
		t.Fatalf("%s - Init data failed: %v", name, err)
	}
	if pkts == nil {
		t.Errorf("%s - Init should not be nil", name)
	}

	file, err := os.CreateTemp("", "srttest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	for i := range pkts {
		buf := pkts[i]
		if _, err := file.Write(buf); err != nil {
			t.Error(err)
		}
	}
	if _, err := file.Write(data[offset:]); err != nil {
		t.Fatal(err)
	}
	if err := file.Sync(); err != nil {
		t.Fatal(err)
	}

	// compare ffprobe results with original
	got, err := ffprobe(file.Name())
	log.Println(got, err)
	if numFrames != len(got.Frames) {
		t.Errorf("%s - Failed ffprobe, got %d frames, expected %d", name, len(got.Frames), numFrames)
	}
}

func TestParser_ParseH264_basic(t *testing.T) {
	// Parse 1s complete MPEG-TS with NAL at start
	data, err := os.ReadFile("h264.ts")
	if err != nil {
		t.Fatalf("failed to open test file")
	}
	p := NewParser()
	checkParser(t, p, data, "simple", 25)
}

func TestParser_ParseH264_complex(t *testing.T) {
	// Parse 1s complete MPEG-TS with NAL at start
	data, err := os.ReadFile("h264_long.ts")
	if err != nil {
		t.Fatalf("failed to open test file")
	}
	log.Println("numpackets", len(data)/PacketLen)

	tests := []struct {
		name           string
		offset         int
		expectedFrames int
	}{
		{"NoOffset", 0, 15},
		{"GOPOffset", 50 * PacketLen, 10},
		{"2GOPOffset", 100 * PacketLen, 5},
	}
	for _, tt := range tests {
		p := NewParser()
		checkParser(t, p, data[tt.offset:], tt.name, tt.expectedFrames)
	}
}

func TestParser_ParseH265_basic(t *testing.T) {
	// Parse 1s complete MPEG-TS with NAL at start
	data, err := os.ReadFile("h265.ts")
	if err != nil {
		t.Fatalf("failed to open test file")
	}
	p := NewParser()
	checkParser(t, p, data, "simple", 25)
}

func TestParser_ParseH265_complex(t *testing.T) {
	// Parse 1s complete MPEG-TS with NAL at start
	data, err := os.ReadFile("h265_long.ts")
	if err != nil {
		t.Fatalf("failed to open test file")
	}
	log.Println("numpackets", len(data)/PacketLen)

	tests := []struct {
		name           string
		offset         int
		expectedFrames int
	}{
		{"NoOffset", 0, 15},
		{"GOPOffset", 50 * PacketLen, 10},
		{"2GOPOffset", 100 * PacketLen, 5},
	}
	for _, tt := range tests {
		p := NewParser()
		checkParser(t, p, data[tt.offset:], tt.name, tt.expectedFrames)
	}
}
