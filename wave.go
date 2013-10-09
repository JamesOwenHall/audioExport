/**
 * AudioExport is a package for exporting raw audio data to uncompressed .wav
 * files.
 *
 * @author James Hall
 * github.com/JamesOwenHall
 */

package audioExport

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
)

/**
 * The WaveFile struct is used to create uncompressed .wav files.
 */
type WaveFile struct {
	file         *os.File
	description  AudioDescription
	bytesWritten uint32
}

/**
 * Creates the file and writes the necessary headers.  The corresponding
 * "Close" method should always be called when you're done with the file.
 * @param {string} fileName The name of the file to be created.
 * @return {error} Non-nil if an error occured.
 */
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

/**
 * Writes the binary waveform to the file.  This can be called several times,
 * so long as the file doesn't reach its 4GB limit.
 * @param {[]byte} bytes The waveform data to be written to the file.  For
 *                       multichannel audio, the data should already be muxed.
 * @return {error} Non-nil if an error occured.
 */
func (w *WaveFile) WriteBytes(bytes []byte) error {
	n, err := w.file.Write(bytes)
	w.bytesWritten += uint32(n)
	return err
}

/**
 * Muxes and writes the channels to the file.  This can be called several
 * times, so long as the file doesn't reach its 4GB limit.
 * @param {[]float64} channels The audio data in each channel.
 * @return {error} Non-nil if an error occured.
 */
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

/**
 * Closes the file.  This should always be called when you're done writing
 * data.
 * @return {error} Non-nil if an error occured.
 */
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

/***** Private Methods *****/

/**
 * Writes the header chunks to the buffer.
 * @return {error} Non-nil if an error occured.
 */
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

/**
 * Writes the container (RIFF) chunk to the buffer.
 * @return {error} Non-nil if an error occured.
 */
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

/**
 * Writes the mandatory fmt chunk to the buffer.
 */
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

/**
 * Writes the start of the data chunk to the buffer.
 */
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

/**
 * Writes the size of the data chunk to its header.
 */
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

/**
 * Writes the size of the RIFF chunk to its header.
 */
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

/**
 * Determines which method to call in order to write the data to the buffer at
 * the right bit depth.
 */
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

/**
 * Writes an 8-bit unsigned integer to the buffer.
 */
func (w *WaveFile) write8BitToBuffer(data float64, buffer *bytes.Buffer) error {
	res := uint8(data*127 + 127)
	return binary.Write(buffer, binary.LittleEndian, res)
}

/**
 * Writes a 16-bit integer to the buffer.
 */
func (w *WaveFile) write16BitToBuffer(data float64, buffer *bytes.Buffer) error {
	res := int16(data * 32767)
	return binary.Write(buffer, binary.LittleEndian, res)
}

/**
 * Writes a 32-bit integer to the buffer.
 */
func (w *WaveFile) write32BitToBuffer(data float64, buffer *bytes.Buffer) error {
	res := int32(data * 2147483647)
	return binary.Write(buffer, binary.LittleEndian, res)
}
