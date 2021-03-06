package photoprism

import (
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/query"
)

// Files represents a list of already indexed file names and their unix modification timestamps.
type Files struct {
	files query.FileMap
	mutex sync.RWMutex
}

// NewFiles returns a new Files instance pointer.
func NewFiles() *Files {
	m := &Files{
		files: make(query.FileMap),
	}

	return m
}

// Init fetches the list from the database once.
func (m *Files) Init() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.files) > 0 {
		// Already initialized.
		return nil
	}

	files, err := query.IndexedFiles()

	if err != nil {
		return fmt.Errorf("%s (query indexed files)", err.Error())
	} else {
		m.files = files
		return nil
	}
}

// Ignore tests of a file requires indexing, file name must be relative to the originals path.
func (m *Files) Ignore(fileName, fileRoot string, modTime time.Time, rescan bool) bool {
	timestamp := modTime.Unix()
	key := path.Join(fileRoot, fileName)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if rescan {
		m.files[key] = timestamp
		return false
	}

	mod, ok := m.files[key]

	if ok && mod == timestamp {
		return true
	} else {
		m.files[key] = timestamp
		return false
	}
}

// Indexed tests of a file was already indexed without modifying the files map.
func (m *Files) Indexed(fileName, fileRoot string, modTime time.Time, rescan bool) bool {
	if rescan {
		return false
	}

	timestamp := modTime.Unix()
	key := path.Join(fileRoot, fileName)

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	mod, ok := m.files[key]

	if ok && mod == timestamp {
		return true
	} else {
		return false
	}
}
