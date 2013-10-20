package audioExport

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
)

// AiffFile is used to create uncompressed .aiff files.
type AiffFile struct {
	file         *os.File
	description  AudioDescription
	bytesWritten int32
}

// Open creates the file and writes the necessary headers.  The corresponding
// Close method should always be called when you're done writing data.
func (a *AiffFile) Open(fileName string, description AudioDescription) error {
	var err error

	a.file, err = os.Create(fileName)
	if err != nil {
		return err
	}

	a.description = description

	buffer := new(bytes.Buffer)
	err = a.writeHeader(buffer)
	if err != nil {
		return err
	}

	_, err = a.file.Write(buffer.Bytes())
	return err
}

// WriteBytes writes the binary waveform to the file.  It expects muxed data
// in the format specified by the audio description.  In most cases,
// WriteChannels is more suitable because it will convert and mux the data for
// you.  WriteBytes can be called several times, so long as the file doesn't
// reach its 4GB limit.
func (a *AiffFile) WriteBytes(bytes []byte) error {
	n, err := a.file.Write(bytes)
	a.bytesWritten += int32(n)
	return err
}

// WriteChannels muxes and writes the channels to the file.  Each channel
// should be a float64 slice where each item in the array ranges from -1 to 1.
// Any values beyond these bounds will be automatically clipped.  WriteChannels
// can be called several times, so long as the file doesn't reach its 4GB
// limit.
func (a *AiffFile) WriteChannels(channels ...[]float64) error {
	var err error

	// If too many channels are given, return an error.
	if len(channels) != int(a.description.NumChannels) {
		return errors.New("The number of audio channels doesn't equal the number of streams supplied.")
	}

	// Make sure the data streams are all of the same length
	var chanLength int = -1
	for i := range channels {
		if chanLength == -1 {
			chanLength = len(channels[i])
			continue
		}

		if len(channels[i]) != chanLength {
			return errors.New("The channels have different amounts of audio data.")
		}
	}

	buffer := new(bytes.Buffer)

	// Write to the buffer
	for i := 0; i < chanLength; i++ {
		for j := range channels {
			err = a.writeFloatToBuffer(channels[j][i], buffer)
			if err != nil {
				return err
			}
		}
	}

	return a.WriteBytes(buffer.Bytes())
}

// Close completes the headers and closes the file.  Close should always be
// called when you're done writing data.
func (a *AiffFile) Close() error {
	var err error

	err = a.closeDataChunk()
	if err != nil {
		return err
	}

	err = a.closeCommonChunk()
	if err != nil {
		return err
	}

	err = a.closeContainerChunk()
	if err != nil {
		return err
	}

	return a.file.Close()
}

/*****************************************************************************/
/****************************** Private Methods ******************************/
/*****************************************************************************/

// writeHeader writes the header chunks to the buffer.
func (a *AiffFile) writeHeader(buffer *bytes.Buffer) error {
	var err error

	err = a.writeContainerChunk(buffer)
	if err != nil {
		return err
	}

	err = a.writeCommonChunk(buffer)
	if err != nil {
		return err
	}

	err = a.startDataChunk(buffer)

	return err
}

// writeContainerChunk writes the container chunk to the buffer.
func (a *AiffFile) writeContainerChunk(buffer *bytes.Buffer) error {
	var err error

	// Chunk ID (FORM)
	_, err = buffer.WriteString("FORM")
	if err != nil {
		return err
	}

	// Chunk size (Unknown at this time)
	err = binary.Write(buffer, binary.BigEndian, int32(0))
	if err != nil {
		return err
	}

	// Format (WAVE)
	_, err = buffer.WriteString("AIFF")
	return err
}

// writeCommonChunk writes the mandatory common chunk to the buffer.
func (a *AiffFile) writeCommonChunk(buffer *bytes.Buffer) error {
	var err error

	// Chunk ID (COMM)
	_, err = buffer.WriteString("COMM")
	if err != nil {
		return err
	}

	// Chunk size (always 18)
	err = binary.Write(buffer, binary.BigEndian, int32(18))
	if err != nil {
		return err
	}

	// Number of channels
	err = binary.Write(buffer, binary.BigEndian, a.description.NumChannels)
	if err != nil {
		return err
	}

	// Number of sample frames (unknown at this time)
	err = binary.Write(buffer, binary.BigEndian, uint32(0))
	if err != nil {
		return err
	}

	// Bits per sample
	err = binary.Write(buffer, binary.BigEndian, a.description.BitsPerSample)
	if err != nil {
		return err
	}

	// Sample rate
	sampleRate, err := a.convertSampleRate()
	if err != nil {
		return err
	}
	_, err = buffer.Write(sampleRate)
	if err != nil {
		return err
	}

	return nil
}

// startDataChunk writes the start of the data chunk to the buffer.
func (a *AiffFile) startDataChunk(buffer *bytes.Buffer) error {
	var err error

	// Chunk ID (SSND)
	_, err = buffer.WriteString("SSND")
	if err != nil {
		return err
	}

	// Chunk size (unknown at this time)
	err = binary.Write(buffer, binary.BigEndian, int32(0))
	if err != nil {
		return err
	}

	// Offset
	err = binary.Write(buffer, binary.BigEndian, uint32(0))
	if err != nil {
		return err
	}

	// Block size
	err = binary.Write(buffer, binary.BigEndian, uint32(0))
	if err != nil {
		return err
	}

	return nil
}

// closeDataChunk writes the size of the data chunk to its header.
func (a *AiffFile) closeDataChunk() error {
	var err error

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.BigEndian, a.bytesWritten+8)
	if err != nil {
		return err
	}

	// The offset of the size of the data chunk is always 42 bytes.
	_, err = a.file.WriteAt(buffer.Bytes(), 42)
	if err != nil {
		return err
	}

	return nil
}

// closeCommonChunk writes the number of sample frames to the common chunk.
func (a *AiffFile) closeCommonChunk() error {
	var err error

	numSampleFrames := uint32(a.bytesWritten / int32(a.description.BitsPerSample))

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.BigEndian, numSampleFrames)
	if err != nil {
		return err
	}

	_, err = a.file.WriteAt(buffer.Bytes(), 22)
	if err != nil {
		return err
	}

	return nil
}

// closeContainerChunk writes the size of the container chunk to its header.
func (a *AiffFile) closeContainerChunk() error {
	var err error

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.BigEndian, a.bytesWritten+46)
	if err != nil {
		return err
	}

	// The offset of the size of the container chunk is always 4 bytes.
	_, err = a.file.WriteAt(buffer.Bytes(), 4)
	if err != nil {
		return err
	}

	return nil
}

// writeFloatToBuffer determines which method to call in order to write the
// data to the buffer at the right bit depth.
func (a *AiffFile) writeFloatToBuffer(data float64, buffer *bytes.Buffer) error {
	switch a.description.BitsPerSample {
	case BPS8:
		return a.write8BitToBuffer(data, buffer)
	case BPS16:
		return a.write16BitToBuffer(data, buffer)
	case BPS32:
		return a.write32BitToBuffer(data, buffer)
	default:
		return errors.New("Invalid bit depth")
	}
	return nil
}

// write8BitToBuffer writes an 8-bit unsigned integer to the buffer.
func (a *AiffFile) write8BitToBuffer(data float64, buffer *bytes.Buffer) error {
	res := uint8(data*127 + 127)
	return binary.Write(buffer, binary.BigEndian, res)
}

// write16BitToBuffer writes a 16-bit integer to the buffer.
func (a *AiffFile) write16BitToBuffer(data float64, buffer *bytes.Buffer) error {
	res := int16(data * 32767)
	return binary.Write(buffer, binary.BigEndian, res)
}

// write32BitToBuffer writes a 32-bit integer to the buffer.
func (a *AiffFile) write32BitToBuffer(data float64, buffer *bytes.Buffer) error {
	res := int32(data * 2147483647)
	return binary.Write(buffer, binary.BigEndian, res)
}

// convertSampleRate generates the 80-bit byte slice corresponding to the
// selected sample rate.
func (a *AiffFile) convertSampleRate() ([]byte, error) {
	switch a.description.SampleRate {
	case SampleRate32k:
		return []byte{64, 13, 250, 0, 0, 0, 0, 0, 0, 0}, nil
	case SampleRate44_1k:
		return []byte{64, 14, 172, 68, 0, 0, 0, 0, 0, 0}, nil
	case SampleRate48k:
		return []byte{64, 14, 187, 128, 0, 0, 0, 0, 0, 0}, nil
	case SampleRate96k:
		return []byte{64, 15, 187, 128, 0, 0, 0, 0, 0, 0}, nil
	case SampleRate192k:
		return []byte{64, 16, 187, 128, 0, 0, 0, 0, 0, 0}, nil
	default:
		return nil, errors.New("Invalid sample rate")
	}
}
