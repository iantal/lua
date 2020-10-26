package service

import (
	"testing"

	"github.com/iantal/lua/protos/lua"
	"github.com/stretchr/testify/assert"
)

func TestMatching(t *testing.T) {
	declarations := []string{
		"org.import1",
		"org.import2",
		"org.import3",
		"org.import4",
		"org.import5",
		"org.import15",
		"org.import25",
	}

	libraries := []*lua.Library{
		&lua.Library{
			Name: "lib1",
			Classes: []string{
				"org.import2",
				"class_x",
				"class_y",
			},
		},
		&lua.Library{
			Name: "lib2",
			Classes: []string{
				"org.import1",
				"class_a",
			},
		},
		&lua.Library{
			Name: "lib3",
			Classes: []string{
				"class_m",
				"org.import6",
				"org.import7",
			},
		},
		&lua.Library{
			Name: "lib4",
			Classes: []string{
				"org.import3",
				"org.import4",
				"org.import5",
				"class_n",
				"org.import8",
			},
		},
	}

	dependencies := match(declarations, libraries)

	assert.Equal(t, "lib1", dependencies[0].Name)
	assert.Equal(t, "org.import2", dependencies[0].Classes)

	assert.Equal(t, "lib2", dependencies[1].Name)
	assert.Equal(t, "org.import1", dependencies[1].Classes)

	assert.Equal(t, "lib4", dependencies[2].Name)
	assert.Equal(t, "org.import3,org.import4,org.import5", dependencies[2].Classes)

}
