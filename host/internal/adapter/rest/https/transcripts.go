package https

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/malikbenkirane/groq-whisper/host/internal/domain/session"
	"github.com/malikbenkirane/groq-whisper/host/internal/domain/transcript"
	"github.com/malikbenkirane/groq-whisper/host/internal/repo"
)

func (a adapter) handlePostTranscript() customHandler {
	return func(w http.ResponseWriter, r *http.Request) (errUser error, errSys error) {
		if r.Header.Get("Content-Type") != "application/json" {
			return errExpectedContentTypeJSON, fmt.Errorf(
				"%w: got %q", errExpectedContentTypeJSON, r.Header.Get("Content-Type"))
		}

		var s session.Id
		{
			intId, err := strconv.Atoi(r.PathValue("session"))
			if err != nil {
				return errBadRequest, fmt.Errorf("%w: %w", errStrconvSession, err)
			}
			s = session.Id(intId)
		}

		var tx struct {
			Tx string
			Ts string
		}

		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			return errInternalError, fmt.Errorf("%w: %w: %w", errDecodeTxPayload, errJsonDecode, err)
		}

		t, err := time.Parse(iso8601, tx.Ts)
		if err != nil {
			return errBadRequest, fmt.Errorf("parse iso8601:  %w", err)
		}

		if err := a.repo.SaveTranscriptChunk(transcript.Chunk{
			Text:      tx.Tx,
			Timestamp: t,
		}, s); err != nil {
			return errInternalError, fmt.Errorf("%w: %w", repo.ErrSaveTranscriptChunk, err)
		}

		return
	}
}

const iso8601 = "2006-01-02T15:04:05.000"
