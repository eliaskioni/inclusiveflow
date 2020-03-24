package gettext_test

import (
	"strings"
	"testing"
	"time"

	"github.com/nyaruka/goflow/utils/gettext"

	"github.com/stretchr/testify/assert"
)

func TestComments(t *testing.T) {
	c := gettext.Comment{}
	assert.Equal(t, "", c.String())

	c = gettext.Comment{
		Translator: "translator",
		Extracted:  "extracted",
		References: []string{"src/foo.go"},
		Flags:      []string{"fuzzy"},
	}
	assert.Equal(t, "#  translator\n#. extracted\n#: src/foo.go\n#, fuzzy\n", c.String())
}

func TestPOs(t *testing.T) {
	po := gettext.NewPO("Generated for testing", time.Date(2020, 3, 25, 11, 50, 30, 123456789, time.UTC), "es")

	po.AddEntry(&gettext.Entry{
		MsgID:  "Yes",
		MsgStr: "",
	})
	po.AddEntry(&gettext.Entry{
		MsgID:  "Yes",
		MsgStr: "Si",
	})

	po.AddEntry(&gettext.Entry{
		MsgContext: "context1",
		MsgID:      "No",
		MsgStr:     "",
	})
	po.AddEntry(&gettext.Entry{
		Comment: gettext.Comment{
			Extracted: "has_text",
		},
		MsgContext: "context1",
		MsgID:      "No",
		MsgStr:     "No",
	})

	b := &strings.Builder{}
	po.Write(b)

	assert.Equal(t, 2, len(po.Entries))
	assert.Equal(
		t, `# Generated for testing
#
#, fuzzy
msgid ""
msgstr ""
"POT-Creation-Date: 2020-03-25 11:50+0000\n"
"Language: es\n"
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"

msgid "Yes"
msgstr "Si"

#. has_text
msgctxt "context1"
msgid "No"
msgstr "No"

`,
		b.String())
}

func TestEncodePOString(t *testing.T) {
	assert.Equal(t, `""`, gettext.EncodePOString(""))
	assert.Equal(t, `"FOO"`, gettext.EncodePOString("FOO"))
	assert.Equal(
		t,
		`""
"FOO\n"
"BAR"`,
		gettext.EncodePOString("FOO\nBAR"),
	)
	assert.Equal(
		t,
		`""
"\n"
"FOO\n"
"\n"
"BAR\n"`,
		gettext.EncodePOString("\nFOO\n\nBAR\n"),
	)
	assert.Equal(
		t,
		`""
"FOO\n"
"\n"
"\n"`,
		gettext.EncodePOString("FOO\n\n\n"),
	)
	assert.Equal(t, `"FOO\tBAR"`, gettext.EncodePOString("FOO\tBAR"))
}
