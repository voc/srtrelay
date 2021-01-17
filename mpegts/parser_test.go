package mpegts

import (
	"encoding/json"
	"io/ioutil"
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
	cmd := exec.Command("ffprobe", "-hide_banner", "-of", "json", "-show_frames", filename)
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
			log.Printf("%s\n", buf[:res])
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

func TestParser_ParseH264_basic(t *testing.T) {
	// Parse 1s complete MPEG-TS with NAL at start
	data, err := ioutil.ReadFile("h264.ts")
	if err != nil {
		t.Fatal("failed to open test file")
	}
	p := NewParser()
	numPackets := len(data) / PacketLen
	offset := 0
	for i := 0; i < numPackets; i++ {
		packet := data[i*PacketLen : (i+1)*PacketLen]
		offset = (i + 1) * PacketLen
		err = p.Parse(packet)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if p.hasInit() {
			break
		}
	}
	if !p.hasInit() {
		t.Error("Should find init")
	}
	pkts, err := p.InitData()
	if err != nil {
		t.Fatalf("Init data failed: %v", err)
	}
	if pkts == nil {
		t.Error("Init should not be nil")
	}
	initCount := len(pkts)
	if initCount != 4 {
		t.Errorf("Expected 4 init packets, got %d", initCount)
	}

	file, err := ioutil.TempFile("", "srttest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	for i := range pkts {
		buf := pkts[i]
		file.Write(buf)
	}
	file.Write(data[offset:])
	file.Sync()

	// compare ffprobe results with original
	wanted, err := ffprobe("h264.ts")
	if err != nil {
		t.Fatal(err)
	}
	got, err := ffprobe(file.Name())
	log.Println(got, err)
	if len(wanted.Frames) != len(got.Frames) {
		t.Errorf("Wrong number of frames")
	}
}
