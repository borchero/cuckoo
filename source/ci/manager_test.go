package ci

import (
	"fmt"
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestTagFromTemplate(t *testing.T) {
	env := ReadEnvironment()

	env.Commit.Tag = "1.4.3"
	env.Commit.Hash = "38d3ff0737d2514f789dc39c2a5d8ca44821a077f0e71f4024cd83b9ba936a1f"
	manager := NewManager(env)

	tag, _ := manager.TagFromTemplate("%t")
	assert.Equal(t, tag, "1.4.3", "%t does not yield full tag.")

	tag, _ = manager.TagFromTemplate("%m")
	assert.Equal(t, tag, "1", "%m does not yield <major> tag.")

	tag, _ = manager.TagFromTemplate("%n")
	assert.Equal(t, tag, "1.4", "%n does not yield <major>.<minor> tag.")

	tag, _ = manager.TagFromTemplate("%h")
	assert.Equal(t, tag, "38d3ff0", "%h does not yield first seven chars of commit hash.")

	date := time.Now().Format("2006-01-02")
	tag, _ = manager.TagFromTemplate("%d")
	assert.Equal(t, tag, date, "%d does not yield current date.")

	tag, _ = manager.TagFromTemplate("%t-rc-%d-%h")
	expected := fmt.Sprintf("1.4.3-rc-%s-38d3ff0", date)
	assert.Equal(t, tag, expected, "Combined tag does not generate correct output.")
}
