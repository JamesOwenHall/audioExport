package audioExport

import (
	"bytes"
	"encoding/binary"
	"os"
)

/**
 * The WaveFile struct is used to create uncompressed .wav files.
 */
type WaveFile struct {
	file        *os.File
	description WaveDescription
}

/**
 * This struct describes the format of the audio data.  See the .wav format
 * specification for a complete definition of these members.
 */
type WaveDescription struct {
	NumChannels   int16
	SampleRate    uint32
	BitsPerSample int16
}

/**
 * Creates the file and writes the necessary headers.  The corresponding
 * "Close" method should always be called when you're done with the file.
 * @param {string} fileName The name of the file to be created.
 * @return {error} Non-nil if an error occured.
 */
func (w *WaveFile) Open(fileName string, description WaveDescription) error {
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
	// TODO: Write bytes to file
	return nil
}

/**
 * Muxes and writes the channels to the file.  This can be called several
 * times, so long as the file doesn't reach its 4GB limit.
 * @param {[]byte} channels The audio data in each channel.  If the file
 *                          description calls for more channels than what is
 *                          passed, the other channels will be filled with
 *                          zeroes.
 * @return {error} Non-nil if an error occured.
 */
func (w *WaveFile) WriteChannels(channels ...[]byte) error {
	// TODO: Mux and write the bytes to the file
	return nil
}

/**
 * Closes the file.  This should always be called when you're done writing
 * data.
 * @return {error} Non-nil if an error occured.
 */
func (w *WaveFile) Close() error {
	// TODO: Adjust block sizes
	return w.file.Close()
}

/***** Private Methods *****/

/**
 * Writes the header chunks to the file.
 * @return {error} Non-nil if an error occured.
 */
func (w *WaveFile) writeHeader(buffer *bytes.Buffer) error {
	var err error

	err = w.writeRIFFChunk(buffer)
	if err != nil {
		return err
	}

	return nil
}

/**
 * Writes the container (RIFF) chunk to the file.
 * @return {error} Non-nil if an error occured.
 */
func (w *WaveFile) writeRIFFChunk(buffer *bytes.Buffer) error {
	var err error

	// ChunkID (RIFF)
	_, err = buffer.WriteString("RIFF")
	if err != nil {
		return err
	}

	// ChunkSize (Unknown at this time)
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

func (w *WaveFile) writeFmtChunk(buffer *bytes.Buffer) error {
	return nil
}
