package sampler

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	"github.com/gordonklaus/portaudio"
	"go.uber.org/zap"
)

func New(log *zap.Logger, sampleRate float64, splitPeriod time.Duration, encoderOpts ...EncoderOption) Sampler {
	s := &sampler{
		e:         NewEncoder(encoderOpts...),
		log:       log,
		sample:    sampleRate,
		splitFreq: splitPeriod,
		chunk:     make(chan chunk, 5),
		err:       make(chan error),
	}
	return s
}

type Sampler interface {
	Sample(ctx context.Context)
}

type sampler struct {
	e         Encoder
	log       *zap.Logger
	sample    float64
	splitFreq time.Duration
	chunk     chan chunk
	err       chan error
}

func (s sampler) Sample(ctx context.Context) {
	go func() {
		for err := range s.err {
			s.log.Error("sample failed", zap.Error(err))
		}
	}()
	go func() {
		if err := s.consume(ctx); err != nil {
			s.err <- fmt.Errorf("consume: %w", err)
		}
	}()
	if err := s.stream(ctx); err != nil {
		s.err <- fmt.Errorf("stream: %w", err)
	}
}

func (s sampler) consume(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			s.log.Error("consume failed", zap.Error(err))
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case c := <-s.chunk:
			pr := s.e.outPath(c.ts, outRaw)
			s.log.Info("writing new chunk", zap.String("path", pr))
			f, err := os.Create(pr)
			if err != nil {
				return fmt.Errorf("os create %q: %w", pr, err)
			}
			defer func() {
				err = f.Close()
			}()
			if _, err = f.Write(c.raw); err != nil {
				return fmt.Errorf("write chunk %q: %w", pr, err)
			}
			if err = f.Close(); err != nil {
				return fmt.Errorf("close chunk %q: %w", pr, err)
			}
			err = convert(s.e, c.ts, strconv.Itoa(int(s.sample)))
			if err != nil {
				return fmt.Errorf("convert: %w", err)
			}
			if err := os.Remove(pr); err != nil {
				return fmt.Errorf("remove %q: %w", pr, err)
			}
		default:
			time.Sleep(time.Second)
		}
	}
}

func (s sampler) stream(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			s.log.Error("sample failed", zap.Error(err))
		}
	}()

	if err := portaudio.Initialize(); err != nil {
		return fmt.Errorf("portaudio Initialize: %w", err)
	}
	defer func() {
		err = portaudio.Terminate()
	}()

	in := make([]int16, 64)
	stream, err := portaudio.OpenDefaultStream(1, 0, float64(s.sample), len(in), in)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}
	defer func() {
		err = stream.Close()
	}()

	if err := stream.Start(); err != nil {
		return fmt.Errorf("stream start: %w", err)
	}

	t := time.NewTicker(time.Duration(s.splitFreq))

	var b bytes.Buffer

loop:
	for {
		if err := stream.Read(); err != nil {
			return fmt.Errorf("stream read: %w", err)
		}
		if err := binary.Write(&b, binary.LittleEndian, in); err != nil {
			return fmt.Errorf("binary write: %w", err)
		}
		select {
		case <-ctx.Done():
			break loop
		case <-t.C:
			ts := time.Now()
			s.chunk <- chunk{
				raw: b.Bytes(),
				ts:  ts,
			}
			b.Reset()
		default:
		}
	}

	if err := stream.Stop(); err != nil {
		return fmt.Errorf("stream stop: %w", err)
	}
	return nil
}

type outType int

const (
	outFlac outType = iota
	outMp3
	outRaw
)

func (t outType) String() string {
	if t == outMp3 {
		return "mp3"
	}
	if t == outFlac {
		return "flac"
	}
	return "raw"
}

func (e Encoder) outPath(ts time.Time, t outType) string {
	filename := fmt.Sprintf("%s.%s", ts.Format("20060102150405"), t)
	return path.Join(e.root, filename)
}

type chunk struct {
	raw []byte
	ts  time.Time
}

func convert(e Encoder, ts time.Time, freq string) (err error) {
	raw, mp3 := e.outPath(ts, outRaw), e.outPath(ts, outMp3)
	args := []string{
		"-f", "s16le",
		"-ar", freq,
		"-ac", "1",
		"-i", raw, mp3,
	}
	if err = e.encode(mp3, args...); err != nil {
		return fmt.Errorf("ffmpeg %q->%q: %w", raw, mp3, err)
	}
	flac := e.outPath(ts, outFlac)
	args = []string{
		"-i", mp3,
		"-ar", freq,
		"-ac", "1",
		"-map", "0:a",
		"-c:a", "flac",
		flac,
	}
	if err = e.encode(flac, args...); err != nil {
		return fmt.Errorf("ffmpeg %q->%q: %w", mp3, flac, err)
	}
	return nil
}

type Encoder struct {
	ffmpegPath string
	root       string
}

type EncoderOption func(Encoder) Encoder

// EncoderOptionPath specifies ffmpeg path
func EncoderOptionPath(path string) EncoderOption {
	return func(e Encoder) Encoder {
		e.ffmpegPath = path
		return e
	}
}

// EncoderOptionRoot specifies were to the encoded samples
func EncoderOptionRoot(root string) EncoderOption {
	return func(e Encoder) Encoder {
		e.root = root
		return e
	}
}

func NewEncoder(opts ...EncoderOption) Encoder {
	encoder := Encoder{ffmpegPath: "ffmpeg"}
	for _, opt := range opts {
		encoder = opt(encoder)
	}
	return encoder
}

func (e Encoder) encode(rm string, args ...string) (err error) {
	_, err = os.Stat(rm)
	if err == nil {
		if err = os.Remove(rm); err != nil {
			return fmt.Errorf("remove %q: %w", rm, err)
		}
	}
	fmt.Println(append([]string{"ffmpeg"}, args...))
	{
		cmd := exec.Command(e.ffmpegPath, args...)
		//cmd.Stdout = os.Stdout
		//cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("exec ffmpeg: %w", err)
		}
		return nil
	}
}
