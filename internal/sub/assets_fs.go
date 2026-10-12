package sub

import (
	"io/fs"
	"net/http"
)

// filesOnlyFS hides directories so http.FileServer can never render a listing
// of the Vite bundle, which would identify the panel to port scanners.
type filesOnlyFS struct {
	http.FileSystem
}

func (f filesOnlyFS) Open(name string) (http.File, error) {
	file, err := f.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	if info.IsDir() {
		_ = file.Close()
		return nil, fs.ErrNotExist
	}
	return file, nil
}
