// Package audioExport provides structures for creating uncompressed audio
// files without linking to external C libraries.
package audioExport

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
)

// WaveFile is used to create uncompressed .wav files.
type WaveFile struct {
	file         *os.File
	description  AudioDescription
	bytesWritten uint32
}

// Open creates the file and writes the necessary headers.  The corresponding
// Close method should always be called when you're done writing data.
func (w *WaveFile) Open(fileName string, description AudioDescription) error {
	var err error

	w.file, err = os.Create(fileName)
	if err != nil {
		return err
	}

	w.description = description

	buffer := new(bytes.Buffer)
	err = w.writeHeader(buffer)
	if err != nil {
		return err
	}

	_, err = w.file.Write(buffer.Bytes())
	return err
}

// WriteBytes writes the binary waveform to the file.  It expects muxed data
// in the format specified by the audio description.  In most cases,
// WriteChannels is more suitable because it will convert and mux the data for
// you.  WriteBytes can be called several times, so long as the file doesn't
// reach its 4GB limit.
func (w *WaveFile) WriteBytes(bytes []byte) error {
	n, err := w.file.Write(bytes)
	w.bytesWritten += uint32(n)
	return err
}

// WriteChannels muxes and writes the channels to the file.  Each channel
// should be a float64 slice where each item in the array ranges from -1 to 1.
// Any values beyond these bounds will be automatically clipped.  WriteChannels
// can be called several times, so long as the file doesn't reach its 4GB
// limit.
func (w *WaveFile) WriteChannels(channels ...[]float64) error {
	var err error

	// If too many channels are given, return an error.
	if len(channels) != int(w.description.NumChannels) {
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
			err = w.writeFloatToBuffer(channels[j][i], buffer)
			if err != nil {
				return err
			}
		}
	}

	return w.WriteBytes(buffer.Bytes())
}

// Close completes the headers and closes the file.  Close should always be
// called when you're done writing data.
func (w *WaveFile) Close() error {
	var err error

	err = w.closeDataChunk()
	if err != nil {
		return err
	}

	err = w.closeRIFFChunk()
	if err != nil {
		return err
	}

	return w.file.Close()
}

// AudioDescription acts as a getter for the AudioDescription provided to the
// Open method.
func (w *WaveFile) AudioDescription() AudioDescription {
	return w.description
}

/*****************************************************************************/
/****************************** Private Methods ******************************/
/*****************************************************************************/

// writeHeader writes the header chunks to the buffer.
func (w *WaveFile) writeHeader(buffer *bytes.Buffer) error {
	var err error

	err = w.writeRIFFChunk(buffer)
	if err != nil {
		return err
	}

	err = w.writeFmtChunk(buffer)
	if err != nil {
		return err
	}

	err = w.startDataChunk(buffer)
	if err != nil {
		return err
	}

	return nil
}

// writeRIFFChunk writes the container (RIFF) chunk to the buffer.
func (w *WaveFile) writeRIFFChunk(buffer *bytes.Buffer) error {
	var err error

	// Chunk ID (RIFF)
	_, err = buffer.WriteString("RIFF")
	if err != nil {
		return err
	}

	// Chunk size (Unknown at this time)
	err = binary.Write(buffer, binary.LittleEndian, uint32(0))
	if err != nil {
		return err
	}

	// Format (WAVE)
	_, err = buffer.WriteString("WAVE")
	if err != nil {
		return err
	}

	return nil
}

// writeFmtChunk writes the mandatory fmt chunk to the buffer.
func (w *WaveFile) writeFmtChunk(buffer *bytes.Buffer) error {
	var err error

	// Chunk ID (fmt )
	_, err = buffer.WriteString("fmt ")
	if err != nil {
		return err
	}

	// Chunk size (always 16)
	err = binary.Write(buffer, binary.LittleEndian, uint32(16))
	if err != nil {
		return err
	}

	// Audio format (1 = uncompressed PCM)
	err = binary.Write(buffer, binary.LittleEndian, uint16(1))
	if err != nil {
		return err
	}

	// Number of channels
	err = binary.Write(buffer, binary.LittleEndian, w.description.NumChannels)
	if err != nil {
		return err
	}

	// Sample rate
	err = binary.Write(buffer, binary.LittleEndian, w.description.SampleRate)
	if err != nil {
		return err
	}

	blockAlign := w.description.NumChannels * w.description.BitsPerSample / 8
	byteRate := w.description.SampleRate * uint32(blockAlign)

	// Byte rate
	err = binary.Write(buffer, binary.LittleEndian, byteRate)
	if err != nil {
		return err
	}

	// Block align
	err = binary.Write(buffer, binary.LittleEndian, blockAlign)
	if err != nil {
		return err
	}

	// Bits per sample
	err = binary.Write(buffer, binary.LittleEndian, w.description.BitsPerSample)
	if err != nil {
		return err
	}

	return nil
}

// startDataChunk writes the start of the data chunk to the buffer.
func (w *WaveFile) startDataChunk(buffer *bytes.Buffer) error {
	var err error

	// Chunk ID (data)
	_, err = buffer.WriteString("data")
	if err != nil {
		return err
	}

	// Chunk size (unknown at this time)
	err = binary.Write(buffer, binary.LittleEndian, uint32(0))
	if err != nil {
		return err
	}

	return nil
}

// closeDataChunk writes the size of the data chunk to its header.
func (w *WaveFile) closeDataChunk() error {
	var err error

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, w.bytesWritten)
	if err != nil {
		return err
	}

	// The offset of the size of the data chunk is always 40 bytes.
	_, err = w.file.WriteAt(buffer.Bytes(), 40)
	if err != nil {
		return err
	}

	return nil
}

// closeRIFFChunk writes the size of the RIFF chunk to its header.
func (w *WaveFile) closeRIFFChunk() error {
	var err error

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, w.bytesWritten+36)
	if err != nil {
		return err
	}

	// The offset of the size of the RIFF chunk is always 4 bytes.
	_, err = w.file.WriteAt(buffer.Bytes(), 4)
	if err != nil {
		return err
	}

	return nil
}

// writeFloatToBuffer determines which method to call in order to write the
// data to the buffer at the right bit depth.
func (w *WaveFile) writeFloatToBuffer(data float64, buffer *bytes.Buffer) error {
	switch w.description.BitsPerSample {
	case BPS8:
		return w.write8BitToBuffer(data, buffer)
	case BPS16:
		return w.write16BitToBuffer(data, buffer)
	case BPS32:
		return w.write32BitToBuffer(data, buffer)
	default:
		return errors.New("Invalid bit depth.")
	}
	return nil
}

// write8BitToBuffer writes an 8-bit unsigned integer to the buffer.
func (w *WaveFile) write8BitToBuffer(data float64, buffer *bytes.Buffer) error {
	res := uint8(data*127 + 127)
	return binary.Write(buffer, binary.LittleEndian, res)
}

// write16BitToBuffer writes a 16-bit integer to the buffer.
func (w *WaveFile) write16BitToBuffer(data float64, buffer *bytes.Buffer) error {
	res := int16(data * 32767)
	return binary.Write(buffer, binary.LittleEndian, res)
}

// write32BitToBuffer writes a 32-bit integer to the buffer.
func (w *WaveFile) write32BitToBuffer(data float64, buffer *bytes.Buffer) error {
	res := int32(data * 2147483647)
	return binary.Write(buffer, binary.LittleEndian, res)
}
