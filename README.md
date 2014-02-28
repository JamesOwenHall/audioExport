####NOTE: This repository is not actively maintained.

AudioExport
===========

Export uncompressed audio files in Go.

##Why Use It

The Go standard library doesn't have any audio capabilities.  With AudioExport, you gain the ability to create audio files without the need to link to external C libraries.  Your application will still compile to a single binary in typical Go style.


##How To Use It

First you need to create an instance of AudioDescription &amp; your choice of audio file.  In this case, we'll use a WaveFile:

    myFile := audioExport.WaveFile{}
    desc := audioExport.AudioDescription{
		NumChannels:   2,
		SampleRate:    audioExport.SampleRate48k,
		BitsPerSample: audioExport.BPS16,
    }
    
Next, you need to open the file and specify a filename:

    err = myFile.Open("/Users/Foo/Bar/myFile.wav", desc)
    
Call the WriteChannels method as many times as you need to write the sound data to the file.  The sound data should be in the form of slices of float64s ranging from -1 to +1.

    err = myFile.WriteChannels(leftChannel, rightChannel)

Close the file and you're done.

    err = myFile.Close()

##Supported Formats

Currently, the only supported file formats are WAV and AIFF.

####File Types
- WAV
- AIFF

####Bits per Sample
- 8
- 16
- 32

####Sample Rates (Hz)
- 32,000
- 44,100
- 48,000
- 96,000
- 192,000
