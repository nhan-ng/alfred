package app

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/nhan-ng/alfred/pkg/util"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var fs = afero.NewOsFs()

type globCloneOptions struct {
	repoURL        string
	globPattern    string
	outdir         string
	excludePattern string
}

func newGlobCloneCommand() *cobra.Command {
	opts := &globCloneOptions{}

	cmd := &cobra.Command{
		Use:   "gclone",
		Short: "Clone a git using glob pattern",
		Run: func(c *cobra.Command, args []string) {
			runGlobClone(opts)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&opts.repoURL, "repo", "r", "https://github.com/example-owner/example-repo", "defines source repo URL to clone.")
	cmd.Flags().StringVarP(&opts.globPattern, "glob", "g", "**", "defines glob pattern to clone.")
	cmd.Flags().StringVarP(&opts.outdir, "outdir", "o", "_gclone", "(optional) output directory to save the files to. Default values: '_gclone'.")
	cmd.Flags().StringVarP(&opts.excludePattern, "exclude", "e", "", "(optional) defines glob pattern to exclude.")

	return cmd
}

func runGlobClone(opts *globCloneOptions) {
	// Clone the git repo in memory
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:   opts.repoURL,
		Depth: 1,
	})
	util.CheckIfError(err)

	// Get the HEAD commit
	ref, err := repo.Head()
	util.CheckIfError(err)
	commit, err := repo.CommitObject(ref.Hash())
	util.CheckIfError(err)
	fmt.Println(commit)

	// Get the tree
	tree, err := commit.Tree()
	util.CheckIfError(err)

	// Create a file system
	afs := &afero.Afero{Fs: afero.NewBasePathFs(fs, opts.outdir)}

	// Initialize workers to clone
	numWorkers := runtime.NumCPU()
	wg := &sync.WaitGroup{}
	wg.Add(numWorkers)
	matched := make(chan *object.File)
	for w := 1; w <= numWorkers; w++ {
		go cloneFileWorker(afs, matched, wg)
	}

	// Iterate through the tree and add the files matching glob to be cloned
	tree.Files().ForEach(func(f *object.File) error {
		// Try to match with the glob pattern
		if match, err := filepath.Match(opts.globPattern, f.Name); err == nil && match {
			// The file matches the pattern, but we check against the exclude pattern, only
			// accept the file if either the exclude pattern is not defined, or it doesn't
			// match
			if len(opts.excludePattern) > 0 {
				if exclude, err := filepath.Match(opts.excludePattern, f.Name); err == nil && !exclude {
					matched <- f
				}
				return err
			}
			matched <- f
		}
		return err
	})
	close(matched)
	wg.Wait()
}

func cloneFileWorker(fs *afero.Afero, jobs <-chan *object.File, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	for f := range jobs {
		content, err := f.Blob.Reader()
		if err != nil {
			util.Warning("There was an error %s processing file '%s'", err, f.Name)
			continue
		}

		fs.SafeWriteReader(f.Name, content)
		fmt.Printf("%s\n", f.Name)
	}
}
