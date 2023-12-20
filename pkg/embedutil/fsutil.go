package embedutil

import "git.act3-ace.com/ace/go-common/pkg/fsutil"

// simple wrapper for initializing an fsutil without creating a temp directory
func newFS(dir string) (*fsutil.FSUtil, error) {
	filesys := &fsutil.FSUtil{
		RootDir: dir,
	}
	err := filesys.AddDir(dir)
	return filesys, err
}

// renamed version of NewFSUtil for clearer code
// since this package usually does not want to
// create temporary directories
var newTempFS = fsutil.NewFSUtil
