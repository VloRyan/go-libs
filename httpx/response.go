package httpx

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"strings"
)

func IsOkStatus(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

func IsClientErrorStatus(statusCode int) bool {
	return statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError
}

func StreamFileAndReplaceToken(w http.ResponseWriter, dir fs.FS, file, old, new string) {
	f, err := dir.Open(file)
	defer func(f fs.File) {
		_ = f.Close()
	}(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Read file failed: " + err.Error()))
		return
	}
	b := make([]byte, 2048)
	i, err := f.Read(b)
	if err != nil && !errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Read file failed: " + err.Error()))
		return
	}
	if i > 0 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

	for i > 0 {
		s := strings.ReplaceAll(string(b[:i]), old, new)
		i = len(s)
		// TODO handling if old is in two chunks
		/*if i > len(old) {
			if partialMatch := stringx.MatchPartial(s[i-len(old):], old); partialMatch > 0 {
				nextBytes := make([]byte, len(old)-partialMatch)
				nextBytesRead, err := f.Read(nextBytes)
				if err != nil && !errors.Is(err, io.EOF) {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte("Read file failed: " + err.Error()))
					return
				}
				s = strings.Replace(s+string(nextBytes[:nextBytesRead]), old, new, 1)
			}
		}*/
		b = []byte(s)
		if _, err = w.Write(b[:i]); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Write file failed: " + err.Error()))
			return
		}
		i, err = f.Read(b)
		if err != nil && !errors.Is(err, io.EOF) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Read file failed: " + err.Error()))
			break
		}
	}
}
