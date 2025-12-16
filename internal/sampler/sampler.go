package sampler

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gordonklaus/portaudio"
	"go.uber.org/zap"
)

func New(log *zap.Logger, sampleRate float64, splitPeriod time.Duration) Sampler {
	s := &sampler{
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
			pr := outPath(c.ts, outRaw)
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
			p3 := outPath(c.ts, outMp3)
			s.log.Info("converting new chunk", zap.String("path", p3))
			err = convert(pr, p3, strconv.Itoa(int(s.sample)/2))
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
	outMp3 outType = iota
	outRaw
)

func outPath(ts time.Time, t outType) string {
	ext := "mp3"
	if t == outRaw {
		ext = "raw"
	}
	return fmt.Sprintf("%s.%s", ts.Format("20060102150405"), ext)
}

type chunk struct {
	raw []byte
	ts  time.Time
}

func convert(raw, mp3, freq string) (err error) {
	_, err = os.Stat(mp3)
	if err == nil {
		if err = os.Remove(mp3); err != nil {
			return fmt.Errorf("remove %q: %w", mp3, err)
		}
	}
	args := []string{
		"-f", "s16le",
		"-ar", freq,
		"-ac", "2",
		"-i", raw, mp3,
	}
	fmt.Println(append([]string{"ffmpeg"}, args...))
	cmd := exec.Command("ffmpeg", args...)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("exec ffmpeg: %w", err)
	}
	return nil
}
