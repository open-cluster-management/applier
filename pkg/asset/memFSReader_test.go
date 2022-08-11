// Copyright Red Hat

package asset

import (
	"reflect"
	"testing"
)

func TestMemFS_AssetNames(t *testing.T) {
	type fields struct {
		data map[string][]byte
	}
	type args struct {
		excluded []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "2 files no execluded",
			fields: fields{
				data: map[string][]byte{
					"file1": []byte("file1"),
					"file2": []byte("file2"),
				},
			},
			args: args{
				excluded: []string{},
			},
			want:    []string{"file1", "file2"},
			wantErr: false,
		},
		{
			name: "2 files 1 execluded",
			fields: fields{
				data: map[string][]byte{
					"file1": []byte("file1"),
					"file2": []byte("file2"),
				},
			},
			args: args{
				excluded: []string{"file2"},
			},
			want:    []string{"file1"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MemFS{
				data: tt.fields.data,
			}
			got, err := r.AssetNames(tt.args.excluded)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemFS.AssetNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemFS.AssetNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemFS_Asset(t *testing.T) {
	type fields struct {
		data map[string][]byte
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "2 files no execluded",
			fields: fields{
				data: map[string][]byte{
					"file1": []byte("file1content"),
					"file2": []byte("file2content"),
				},
			},
			args: args{
				name: "file1",
			},
			want:    []byte("file1content"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MemFS{
				data: tt.fields.data,
			}
			got, err := r.Asset(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemFS.Asset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemFS.Asset() = %v, want %v", got, tt.want)
			}
		})
	}
}
