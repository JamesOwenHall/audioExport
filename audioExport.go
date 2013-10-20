// Package audioExport provides structures for creating uncompressed audio
// files without linking to external C libraries.
package audioExport

type AudioFile interface {
	Open(fileName string, description AudioDescription) error
	WriteChannels(channels ...[]float64) error
	Close() error
}

// AudioDescription describes the format of the audio data.
type AudioDescription struct {
	NumChannels   int16
	SampleRate    uint32
	BitsPerSample int16
}

// The SampleRate constants provide a list of the most common sample rates.
// For most solutions, 48k should be sufficient.
const (
	SampleRate32k   uint32 = 32000
	SampleRate44_1k uint32 = 44100
	SampleRate48k   uint32 = 48000
	SampleRate96k   uint32 = 96000
	SampleRate192k  uint32 = 192000
)

// The BPS constants list the possible values for the BitsPerSample member of
// the Audio Description struct.
const (
	BPS8  int16 = 8
	BPS16 int16 = 16
	BPS32 int16 = 32
)
