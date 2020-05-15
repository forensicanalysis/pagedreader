package pagedreader

import (
	"bytes"
	"reflect"
	"testing"
)

func TestPagedReader_ReadAt(t *testing.T) {
	type fields struct {
		buf       []byte
		pagesize  int64
		cacheSize int
	}
	type args struct {
		buf    []byte
		offset int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{"read 2", fields{buf: []byte("abcd"), pagesize: 1, cacheSize: 100}, args{buf: make([]byte, 2), offset: 0}, []byte("ab"), false},
		{"unmatching page size", fields{buf: []byte("abcd"), pagesize: 3, cacheSize: 100}, args{buf: make([]byte, 1), offset: 3}, []byte("d"), false},
		{"matching page size", fields{buf: []byte("abcd"), pagesize: 2, cacheSize: 100}, args{buf: make([]byte, 1), offset: 3}, []byte("d"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self, err := New(bytes.NewReader(tt.fields.buf), tt.fields.pagesize, tt.fields.cacheSize)
			if err != nil {
				t.Fatal(err)
			}
			_, err = self.ReadAt(tt.args.buf, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.args.buf, tt.want) {
				t.Errorf("ReadAt() got = %v, want %v", tt.args.buf, tt.want)
			}
		})
	}
}
