package cache

import (
	"bytes"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/alecthomas/kong"
	"github.com/pkg/errors"
	"hash/fnv"
	"io"
	"os"
)

type Cache struct {
	Hash uint32 `json:"hash"`
	Step int    `json:"step"`
}

func IsCacheExists(cacheFilePath string) bool {
	_, err := os.Stat(cacheFilePath)
	return !os.IsNotExist(err)
}

func AskToContinueExecution(kongCtx *kong.Context) bool {
	var response bool
	kongCtx.FatalIfErrorf(survey.AskOne(&survey.Confirm{
		Message: "An uncompleted plan execution was found. Would you like to continue executing this plan from where you left off?",
	}, &response), "")
	return response
}

func ClearCache(cacheFilePath string, kongCtx *kong.Context) {
	err := os.Remove(cacheFilePath)
	if err != nil {
		kongCtx.FatalIfErrorf(err, "Failed to remove cache file.")
	}
}

func CalculateHash(buff []byte) (uint32, error) {
	h := fnv.New32a()
	buffer := &bytes.Buffer{}
	buffer.Write(buff)
	if _, err := io.Copy(h, buffer); err != nil {
		return 0, errors.Wrap(err, "cannot copy file content")
	}
	return h.Sum32(), nil
}

func (c *Cache) String() string {
	return fmt.Sprintf("hash: %d\nstep: %d", c.Hash, c.Step)
}
