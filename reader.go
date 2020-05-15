// This code is adapted from https://github.com/Velocidex/go-ntfs/blob/master/parser/reader.go

package pagedreader

import (
	"io"
	"sync"
	"www.velocidex.com/golang/go-ntfs/parser"
)

type PagedReader struct {
	mu sync.Mutex

	reader   io.ReaderAt
	pagesize int64
	lru      *parser.LRU

	Hits int64
	Miss int64
}

func New(reader io.ReaderAt, pagesize int64, cacheSize int) (*PagedReader, error) {
	cache, err := parser.NewLRU(cacheSize, nil)
	if err != nil {
		return nil, err
	}

	return &PagedReader{
		reader:   reader,
		pagesize: pagesize,
		lru:      cache,
	}, nil
}

func (r *PagedReader) ReadAt(buf []byte, offset int64) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	bufIdx := 0
	for {
		// How much is left in this page to read?
		toRead := int(r.pagesize - offset%r.pagesize)

		// How much do we need to read into the buffer?
		if toRead > len(buf)-bufIdx {
			toRead = len(buf) - bufIdx
		}

		// Are we done?
		if toRead == 0 {
			return bufIdx, nil
		}

		var pageBuf []byte

		page := offset - offset%r.pagesize
		cachedPageBuf, pres := r.lru.Get(int(page))
		if !pres {
			r.Miss += 1
			// Read this page into memory.
			pageBuf = make([]byte, r.pagesize)
			got, err := r.reader.ReadAt(pageBuf, page)
			readEnough := err == io.EOF && got >= toRead
			if err != nil && !readEnough {
				return bufIdx, err
			}

			r.lru.Add(int(page), pageBuf)
		} else {
			r.Hits += 1
			pageBuf = cachedPageBuf.([]byte)
		}

		// Copy the relevant data from the page.
		pageOffset := int(offset % r.pagesize)
		copy(buf[bufIdx:bufIdx+toRead], pageBuf[pageOffset:pageOffset+toRead])

		offset += int64(toRead)
		bufIdx += toRead
	}
}
